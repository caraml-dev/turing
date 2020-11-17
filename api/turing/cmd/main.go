package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/cors"

	"github.com/gojek/mlp/pkg/authz/enforcer"
	"github.com/gojek/mlp/pkg/instrumentation/sentry"
	"github.com/gojek/turing/api/turing/api"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/vault"
	"github.com/gojek/turing/api/turing/web"
	"github.com/heptiolabs/healthcheck"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type configFlags []string

func (c *configFlags) String() string {
	return strings.Join(*c, ",")
}

func (c *configFlags) Set(value string) error {
	*c = append(*c, value)
	return nil
}

func main() {
	var configFlags configFlags
	flag.Var(&configFlags, "config", "Path to a configuration file. This flag can be specified multiple "+
		"times to load multiple configurations.")
	flag.Parse()

	if len(configFlags) < 1 {
		log.Panicf("Must specify at least one config path using -config")
	}

	cfg, err := config.FromFiles(configFlags...)
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

	// Init auth enforcer, vault client
	var authEnforcer *enforcer.Enforcer
	if cfg.AuthConfig.Enabled {
		// Use product mlp as the policies are shared across the mlp products.
		ae, err := enforcer.NewEnforcerBuilder().Product("mlp").
			URL(cfg.AuthConfig.URL).Build()
		if err != nil {
			log.Panicf("Failed initializing authorization enforcer %v", err)
		}
		authEnforcer = &ae
	}
	vaultClient, err := vault.NewClientFromConfig(cfg)
	if err != nil {
		log.Panicf("Failed initializing vault client: %v", err)
	}

	// Init app context
	appCtx, err := api.NewAppContext(db, cfg, authEnforcer, vaultClient)
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

	// Register handlers
	health := healthcheck.NewHandler()
	health.AddReadinessCheck("database", healthcheck.DatabasePingCheck(db.DB(), 1*time.Second))

	mux := http.NewServeMux()
	mux.Handle("/v1/internal/", http.StripPrefix("/v1/internal", health))
	mux.Handle("/v1/", http.StripPrefix("/v1", api.NewRouter(appCtx)))
	// Serve Swagger Spec
	mux.Handle("/swagger.yaml", web.FileHandler("./swagger.yaml"))
	// Serve UI
	if cfg.TuringUIConfig.AppDirectory != "" {
		log.Infof(
			"Serving Turing UI from %s at %s",
			cfg.TuringUIConfig.AppDirectory,
			cfg.TuringUIConfig.Homepage)
		web.ServeReactApp(mux, cfg.TuringUIConfig.Homepage, cfg.TuringUIConfig.AppDirectory)
	}

	log.Infof("Listening on port %d", cfg.Port)
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: cfg.AllowedOrigins,
	})
	if err := http.ListenAndServe(cfg.ListenAddress(), corsHandler.Handler(mux)); err != nil {
		log.Errorf("Failed to start turing-api: %s", err)
	}
}
