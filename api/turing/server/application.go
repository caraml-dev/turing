package server

import (
	"flag"
	"net/http"
	"strings"
	"time"

	"github.com/gojek/mlp/api/pkg/authz/enforcer"
	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/gorilla/mux"

	"github.com/caraml-dev/turing/api/turing/api"
	batchrunner "github.com/caraml-dev/turing/api/turing/batch/runner"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/database"
	"github.com/caraml-dev/turing/api/turing/log"
	"github.com/caraml-dev/turing/api/turing/middleware"
	"github.com/caraml-dev/turing/api/turing/vault"
)

type configFlags []string

func (c *configFlags) String() string {
	return strings.Join(*c, ",")
}

func (c *configFlags) Set(value string) error {
	*c = append(*c, value)
	return nil
}

func Run() {
	var configFlags configFlags
	flag.Var(&configFlags, "config", "Path to a configuration file. This flag can be specified multiple "+
		"times to load multiple configurations.")
	flag.Parse()

	if len(configFlags) < 1 {
		log.Panicf("Must specify at least one config path using -config")
	}

	cfg, err := config.Load(configFlags...)
	if err != nil {
		log.Panicf("%s", err)
	}
	err = cfg.Validate()
	if err != nil {
		log.Panicf("Failed validating config: %s", err)
	}

	// Configure global logger
	if len(cfg.LogLevel) > 0 {
		if err = log.SetLogLevelAt(cfg.LogLevel); err != nil {
			log.Panicf("Failed to configure global logger: %s", err)
		}
	}

	// init db
	db, err := database.InitDB(cfg.DbConfig)
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	// Initialise NewRelic
	if err := newrelic.InitNewRelic(cfg.NewRelicConfig); err != nil {
		log.Errorf("Failed to initialize newrelic: %s", err)
	}
	defer newrelic.Shutdown(5 * time.Second)

	// Init Authorizer
	var authorizer *middleware.Authorizer
	apiPathPrefix := "/v1"
	if cfg.AuthConfig.Enabled {
		// Use product mlp as the policies are shared across the mlp products.
		authEnforcer, err := enforcer.NewEnforcerBuilder().Product("mlp").URL(cfg.AuthConfig.URL).Build()
		if err != nil {
			log.Panicf("Failed initializing authorization enforcer %v", err)
		}
		authorizer, err = middleware.NewAuthorizer(authEnforcer, apiPathPrefix)
		if err != nil {
			log.Panicf("Failed initializing Authorizer %v", err)
		}
	}

	// Init Vault client
	vaultClient, err := vault.NewClientFromConfig(cfg)
	if err != nil {
		log.Panicf("Failed initializing vault client: %v", err)
	}

	// Init app context
	appCtx, err := api.NewAppContext(db, cfg, authorizer, vaultClient)
	if err != nil {
		log.Panicf("Failed initializing application context: %v", err)
	}

	// Init Sentry client
	if cfg.Sentry.Enabled {
		cfg.Sentry.Labels = map[string]string{
			"environment": cfg.DeployConfig.EnvironmentType,
			"app":         "turing-api",
		}
		if err := sentry.InitSentry(cfg.Sentry); err != nil {
			log.Errorf("Failed initializing sentry client: %s", err)
		}
		defer sentry.Close()
	}

	// Run batch runners
	go batchrunner.RunBatchRunners(appCtx.BatchRunners)

	// Register handlers
	r := mux.NewRouter()

	// HealthCheck Handler
	AddHealthCheckHandler(r, "/v1/internal", sqlDB)

	// Write to a file so that Swagger UI can use it
	err = cfg.OpenapiConfig.GenerateSpecFile()
	if err != nil {
		log.Panicf("failed to write openapi yaml file: %s", err)
	}

	// API Handler
	err = AddAPIRoutesHandler(r, apiPathPrefix, appCtx, cfg)
	if err != nil {
		log.Panicf("Failed to configure API routes: %v", err)
	}

	// Serve Swagger UI
	if spaCfg := cfg.OpenapiConfig.SwaggerUIConfig; spaCfg != nil && len(spaCfg.ServingDirectory) > 0 {
		log.Infof("Serving Swagger UI at: %s", spaCfg.ServingPath)
		ServeSinglePageApplication(r,
			spaCfg.ServingPath,
			spaCfg.ServingDirectory)
	} else {
		log.Warnf("Swagger UI not configured to run.")
	}

	// Serve Turing UI
	if spaCfg := cfg.TuringUIConfig; spaCfg != nil && len(spaCfg.ServingDirectory) > 0 {
		log.Infof("Serving Turing UI at: %s", spaCfg.ServingPath)
		ServeSinglePageApplication(r,
			spaCfg.ServingPath,
			spaCfg.ServingDirectory)
	} else {
		log.Warnf("Turing UI not configured to run.")
	}

	log.Infof("Listening on port %d", cfg.Port)
	if err := http.ListenAndServe(cfg.ListenAddress(), r); err != nil {
		log.Errorf("Failed to start turing-api: %s", err)
	}
}
