package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/caraml-dev/turing/engines/router/missionctl"
	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/http/handlers"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber/protocol"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	// Turing router will support these experiment runners: nop
	_ "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner/nop"
	// TODO: justify this
	_ "gopkg.in/confluentinc/confluent-kafka-go.v1/kafka/librdkafka"
	// TODO: justify this
	_ "net/http/pprof"
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

	if strings.EqualFold(cfg.RouterConfig.Protocol, string(protocol.GRPC)) {
		// Init mission control
		missionCtl, err := missionctl.NewMissionControlGrpc(
			cfg.RouterConfig.ConfigFile,
			cfg.AppConfig.FiberDebugLog,
		)
		if err != nil {
			log.Glob().Panicf("Failed initializing Mission Control: %v", err)
		}
		s := grpc.NewServer()
		upiv1.RegisterUniversalPredictionServiceServer(s, missionCtl)
		reflection.Register(s)
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
		if err != nil {
			log.Glob().Errorf("Failed to listen on port %d: %s", cfg.Port, err)
		}
		if err := s.Serve(listener); err != nil {
			log.Glob().Errorf("Failed to start Turing Mission Control API: %s", err)
		}
	} else {
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
			handlers.NewInternalAPIHandler([]string{
				cfg.EnsemblerConfig.Endpoint,
				cfg.EnrichmentConfig.Endpoint,
			}),
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
