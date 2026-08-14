package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/trace"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

// ShutdownFunc flushes and shuts down the tracer provider created by InitGlobalTracer.
type ShutdownFunc func(context.Context) error

// Tracer represents a generic tracer that supports initialization of a global
// tracing client and creation of OpenTelemetry spans
type Tracer interface {
	InitGlobalTracer(string, *config.JaegerConfig) (ShutdownFunc, error)
	IsEnabled() bool
	StartSpanFromRequestHeader(
		context.Context,
		string,
		http.Header,
	) (trace.Span, context.Context)
	StartSpanFromContext(context.Context, string) (trace.Span, context.Context)
}

// globalTracer is initialised to a Nop tracer, calling InitGlobalTracer will reset this
var globalTracer Tracer = newNopTracer()

// InitGlobalTracer creates a new OTel tracer exporting via OTLP HTTP, and sets it as global
// tracer. The returned ShutdownFunc is always non-nil and safe to call, even when a non-nil
// error is also returned, so callers can unconditionally defer it rather than relying on a
// Fatal-on-error caller to skip the call.
func InitGlobalTracer(name string, jaegerCfg *config.JaegerConfig) (ShutdownFunc, error) {
	noopShutdown := func(context.Context) error { return nil }

	// If jaeger config has not been set or tracing is not enabled, just (re-)initialise
	// whatever tracer is currently global (typically the Nop tracer).
	if jaegerCfg == nil || !jaegerCfg.Enabled {
		return globalTracer.InitGlobalTracer(name, jaegerCfg)
	}

	// Tracing is enabled: initialise a new OtelTracer, but only swap it in as the
	// global tracer if initialisation succeeds. Otherwise leave the previous
	// globalTracer (e.g. the Nop tracer) in place, so callers never observe a
	// half-initialised OtelTracer whose IsEnabled() returns true but whose
	// tracer/propagator fields are nil (which would panic on use).
	otelTracer := newOtelTracer()
	shutdown, err := otelTracer.InitGlobalTracer(name, jaegerCfg)
	if err != nil {
		return noopShutdown, err
	}
	globalTracer = otelTracer
	return shutdown, nil
}

// Glob returns the global tracer
func Glob() Tracer {
	return globalTracer
}

// SetGlob sets the global tracer, for testing
func SetGlob(t Tracer) {
	globalTracer = t
}
