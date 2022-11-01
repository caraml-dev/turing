package server

import (
	"fmt"
	"io"
	"net"
	"net/http"

	// TODO: justify this
	_ "net/http/pprof"

	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/soheilhy/cmux"

	"github.com/caraml-dev/turing/engines/router/missionctl"
	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/http/handlers"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/upi"

	// Turing router will support these experiment runners: nop
	_ "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner/nop"
	// TODO: justify this
	_ "gopkg.in/confluentinc/confluent-kafka-go.v1/kafka/librdkafka"
)

func Run() {
	// Read env vars
	cfg, err := config.InitConfigEnv()
	if err != nil {
		log.Glob().Panicf("Failed initializing config: %v", err)
	}

	// Init logger
	log.InitGlobalLogger(cfg.AppConfig)
	defer func() {
		_ = log.Glob().Sync()
	}()

	// Init Turing result logger
	err = resultlog.InitTuringResultLogger(cfg.AppConfig)
	if err != nil {
		log.Glob().Fatalf("Failed initializing Turing Result Logger: %v", err)
	}

	// Init instrumentation, defer closing tracer
	defer initInstrumentation(cfg)()
	// Init Sentry, defer closing client
	defer initSentryClient(cfg)()

	switch cfg.RouterConfig.Protocol {
	case config.UPI:
		// Init mission control
		missionCtl, err := missionctl.NewMissionControlUPI(
			cfg.RouterConfig.ConfigFile,
			cfg.AppConfig.FiberDebugLog,
		)
		if err != nil {
			log.Glob().Panicf("Failed initializing Mission Control: %v", err)
		}

		l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
		if err != nil {
			log.Glob().Panicf("Failed to listen on port: %v", cfg.Port)
		}

		upiServer := upi.NewUPIServer(missionCtl)
		m := cmux.New(l)
		grpcListener := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
		httpListener := m.Match(cmux.Any())

		mux := http.NewServeMux()
		mux.Handle("/v1/internal/", http.StripPrefix(
			"/v1/internal",
			handlers.NewInternalAPIHandler([]string{}),
		))
		if cfg.AppConfig.CustomMetrics {
			mux.Handle("/metrics", promhttp.Handler())
		}
		httpServer := &http.Server{Handler: mux}

		log.Glob().Infof("Starting UPI Router in port %d", cfg.Port)
		go upiServer.Run(grpcListener)
		go func() {
			if err := httpServer.Serve(httpListener); err != nil {
				log.Glob().Errorf("Failed to serve http server: %s", err)
			}
		}()
		if err := m.Serve(); err != nil {
			log.Glob().Errorf("Failed to serve cmux: %s", err)
		}
	case config.HTTP:
		// Init mission control
		missionCtl, err := missionctl.NewMissionControl(
			nil,
			cfg.EnrichmentConfig,
			cfg.RouterConfig,
			cfg.EnsemblerConfig,
			cfg.AppConfig,
		)
		if err != nil {
			log.Glob().Panicf("Failed initializing Mission Control: %v", err)
		}
		// Register handlers
		http.Handle("/v1/internal/", http.StripPrefix(
			"/v1/internal",
			handlers.NewInternalAPIHandler([]string{}),
		))
		http.Handle("/v1/predict", sentry.Recoverer(handlers.NewHTTPHandler(missionCtl)))
		http.Handle("/v1/batch_predict", sentry.Recoverer(handlers.NewBatchHTTPHandler(missionCtl)))
		// Register metrics handler
		if cfg.AppConfig.CustomMetrics {
			http.Handle("/metrics", promhttp.Handler())
		}
		// Serve
		log.Glob().Infof("listening at port %d", cfg.Port)
		if err := http.ListenAndServe(cfg.ListenAddress(), http.DefaultServeMux); err != nil {
			log.Glob().Errorf("Failed to start Turing Mission Control API: %s", err)
		}
	default:
		log.Glob().Panicf("router protocol %s not supported", cfg.RouterConfig.Protocol)
	}
}

// initInstrumentation initializes the metrics collector and tracing client
func initInstrumentation(cfg *config.Config) func() {
	var tracingCloser io.Closer
	var err error

	// Init metrics collector
	err = metrics.InitMetricsCollector(cfg.AppConfig.CustomMetrics)
	if err != nil {
		log.Glob().Fatalf("Failed initializing Metrics Collector: %v", err)
	}

	// Init tracing client
	tracingCloser, err = tracing.InitGlobalTracer(cfg.AppConfig.Name, cfg.AppConfig.Jaeger)
	if err != nil {
		log.Glob().Fatalf("Failed initializing Tracer: %v", err)
	}
	// Return closer function
	return func() {
		if err := tracingCloser.Close(); err != nil {
			panic(err)
		}
	}
}

// initSentryClient initializes the Sentry client for error logging
func initSentryClient(cfg *config.Config) func() {
	if cfg.AppConfig.Sentry.Enabled {
		cfg.AppConfig.Sentry.Labels = map[string]string{
			"environment": cfg.AppConfig.Environment,
			"app":         fmt.Sprintf("turing-router-%s", cfg.AppConfig.Name),
		}
		if err := sentry.InitSentry(cfg.AppConfig.Sentry); err != nil {
			log.Glob().Fatalf("Failed initializing sentry client: %s", err)
		}
		return func() {
			defer sentry.Close()
		}
	}
	// Sentry not enabled, return dummy close function
	return func() {}
}
