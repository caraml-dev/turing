package server

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gojek/mlp/api/pkg/authz/enforcer"
	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/gojek/turing/api/turing/api"
	batchrunner "github.com/gojek/turing/api/turing/batch/runner"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/middleware"
	"github.com/gojek/turing/api/turing/vault"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
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

	// Init db
	db, err := gorm.Open(
		"postgres",
		fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
			cfg.DbConfig.Host,
			cfg.DbConfig.Port,
			cfg.DbConfig.User,
			cfg.DbConfig.Database,
			cfg.DbConfig.Password))
	if err != nil {
		panic(err)
	}
	db.LogMode(false)
	defer db.Close()

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
	AddHealthCheckHandler(r, "/v1/internal", db)

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
	}

	// Serve Turing UI
	if spaCfg := cfg.TuringUIConfig; spaCfg != nil && len(spaCfg.ServingDirectory) > 0 {
		log.Infof("Serving Turing UI at: %s", spaCfg.ServingPath)
		ServeSinglePageApplication(r,
			spaCfg.ServingPath,
			spaCfg.ServingDirectory)
	}

	log.Infof("Listening on port %d", cfg.Port)
	if err := http.ListenAndServe(cfg.ListenAddress(), r); err != nil {
		log.Errorf("Failed to start turing-api: %s", err)
	}
}
