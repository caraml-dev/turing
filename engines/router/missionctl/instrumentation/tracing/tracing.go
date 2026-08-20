package tracing

import (
	"context"
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/trace"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

// ShutdownFunc flushes and shuts down the tracer provider(s) created by InitGlobalTracer.
type ShutdownFunc func(context.Context) error

// Tracer represents a generic tracer that supports creation of OpenTelemetry spans.
// Concrete tracers are constructed via their own typed newXTracer function (see
// newOtelTracer, newJaegerTracer, newMultiTracer) rather than through this interface,
// since different backends take different config types.
type Tracer interface {
	IsEnabled() bool
	StartSpanFromRequestHeader(
		context.Context,
		string,
		http.Header,
	) (trace.Span, context.Context)
	StartSpanFromContext(context.Context, string) (trace.Span, context.Context)
}

// globalTracer is initialised to a Nop tracer, calling InitGlobalTracer will reset this
var globalTracer = newNopTracer()

// InitGlobalTracer initialises whichever of jaegerCfg/otelCfg are enabled, and sets the
// global tracer to: the Nop tracer if neither is enabled, that single backend's tracer
// if exactly one is enabled, or a MultiTracer fanning out to both if both are enabled.
// The returned ShutdownFunc is always non-nil and safe to call, even when a non-nil
// error is also returned, so callers can unconditionally defer it.
func InitGlobalTracer(
	name string,
	jaegerCfg *config.JaegerConfig, //nolint:staticcheck
	otelCfg *config.OtelConfig,
) (ShutdownFunc, error) {
	var tracers []Tracer
	var shutdowns []ShutdownFunc

	// NOTE: order matters here. Otel is appended before Jaeger so that, when both are
	// enabled, tracers[0] (and thus spans[0] in the resulting multiSpan) is always the
	// Otel tracer/span. multiSpan's SpanContext()/IsRecording()/TracerProvider() delegate
	// to spans[0] as the composite's "primary" identity (see multi.go's multiSpan doc
	// comment) -- reordering these two blocks would silently flip which backend that
	// primary identity comes from, so don't reorder casually.
	if otelCfg != nil && otelCfg.Enabled {
		t, shutdown, err := newOtelTracer(name, otelCfg)
		if err != nil {
			return aggregateShutdown(shutdowns), err
		}
		tracers = append(tracers, t)
		shutdowns = append(shutdowns, shutdown)
	}

	// Must stay after the Otel block above -- see the ordering note there.
	if jaegerCfg != nil && jaegerCfg.Enabled {
		t, shutdown, err := newJaegerTracer(name, jaegerCfg)
		if err != nil {
			return aggregateShutdown(shutdowns), err
		}
		tracers = append(tracers, t)
		shutdowns = append(shutdowns, shutdown)
	}

	switch len(tracers) {
	case 0:
		globalTracer = newNopTracer()
	case 1:
		globalTracer = tracers[0]
	default:
		globalTracer = newMultiTracer(tracers)
	}

	return aggregateShutdown(shutdowns), nil
}

// aggregateShutdown returns a ShutdownFunc that always attempts every shutdown func
// in shutdowns (even if an earlier one errors), combining any errors.
func aggregateShutdown(shutdowns []ShutdownFunc) ShutdownFunc {
	return func(ctx context.Context) error {
		var errs []error
		for _, shutdown := range shutdowns {
			if err := shutdown(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}
}

// Glob returns the global tracer
func Glob() Tracer {
	return globalTracer
}

// SetGlob sets the global tracer, for testing
func SetGlob(t Tracer) {
	globalTracer = t
}
