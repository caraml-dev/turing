package server

import (
	"context"
	"fmt"

	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/caraml-dev/turing/api/turing/config"
)

const appName = "turing-api"

// initTracer initializes the global OpenTelemetry tracer provider for api, exporting
// spans via OTLP HTTP, and returns its shutdown function. When cfg.Enabled is false,
// it returns a no-op shutdown function and leaves the OTel globals untouched. The
// returned shutdown function is always non-nil and safe to call, even when a non-nil
// error is also returned, so callers can unconditionally defer it.
func initTracer(cfg config.OtelConfig) (func(context.Context) error, error) {
	noopShutdown := func(context.Context) error { return nil }

	if !cfg.Enabled {
		return noopShutdown, nil
	}

	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(cfg.OtlpEndpoint))
	if err != nil {
		return noopShutdown, err
	}

	ratio := cfg.SamplingRatio
	if ratio <= 0 {
		ratio = 1
	}

	res := resource.NewSchemaless(attribute.String("service.name", appName))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// initProfiler starts continuous profiling via pyroscope-go if enabled in cfg. Returns
// a nil profiler and nil error when profiling is disabled. pyroscope.Start does not itself
// error on an empty ServerAddress -- it happily constructs a client that fails silently on
// every upload -- so an empty address is rejected explicitly here instead.
func initProfiler(cfg config.PyroscopeConfig) (*pyroscope.Profiler, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.ServerAddress == "" {
		return nil, fmt.Errorf("pyroscope profiling is enabled but ServerAddress is empty")
	}

	return pyroscope.Start(pyroscope.Config{
		ApplicationName: appName,
		ServerAddress:   cfg.ServerAddress,
		HTTPHeaders:     cfg.HTTPHeaders,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
		},
	})
}
