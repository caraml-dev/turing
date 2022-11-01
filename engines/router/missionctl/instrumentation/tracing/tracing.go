package tracing

import (
	"context"
	"io"
	"net/http"

	"github.com/opentracing/opentracing-go"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

// TraceConfig exposes getters and setters for some of the configuration applied to the tracer
type TraceConfig interface {
	IsStartNewSpans() bool
	SetStartNewSpans(bool)
}

// tracingConfig implements the TraceConfig interface and captures the subset of the configuration
// applied to tracing in the app, which determines whether a span should be created or skipped.
// startNewSpans determines if all requests must be traced (startNewSpans = true) or only those
// that already have tracing span info in the request header (in which case, a span is only
// continued with a child span and entirely new spans are not created).
type tracingConfig struct {
	startNewSpans bool
}

// IsStartNewSpans tells whether the tracer can start new spans
func (c *tracingConfig) IsStartNewSpans() bool {
	return c.startNewSpans
}

// SetStartNewSpans is a setter for the global tracer config's startNewSpans
func (c *tracingConfig) SetStartNewSpans(newSpans bool) {
	c.startNewSpans = newSpans
}

// Tracer represents a generic tracer that supports initialization of a global
// tracing client and creation of opentracing spans
type Tracer interface {
	TraceConfig
	InitGlobalTracer(string, *config.JaegerConfig) (io.Closer, error)
	IsEnabled() bool
	StartSpanFromRequestHeader(
		context.Context,
		string,
		http.Header,
	) (opentracing.Span, context.Context)
	StartSpanFromContext(context.Context, string) (opentracing.Span, context.Context)
}

// baseTracer partially implements the Tracer interface and can be used by all tracers
// to support creation of opentracing spans
type baseTracer struct {
	*tracingConfig
}

// StartSpanFromRequestHeader attempts to extract span info from the request header and creates a
// new / child span accordingly, which is associated to the given context.Context object.
func (t *baseTracer) StartSpanFromRequestHeader(
	ctx context.Context,
	opName string,
	header http.Header,
) (opentracing.Span, context.Context) {
	tr := opentracing.GlobalTracer()
	spanCtx, _ := tr.Extract(
		opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))
	return t.newOrChildSpan(ctx, spanCtx, opName)
}

// StartSpanFromContext attempts to extract span info from the given context.Context and creates
// a new / child span accordingly, which is associated to the same context object.
func (t *baseTracer) StartSpanFromContext(
	ctx context.Context,
	opName string,
) (opentracing.Span, context.Context) {
	var spanCtx opentracing.SpanContext
	// Retrieve the span from given context
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		spanCtx = span.Context()
	}
	return t.newOrChildSpan(ctx, spanCtx, opName)
}

// newOrChildSpan creates a child span from the given span context if not empty. If
// it is empty, a new span will be created based on the value of IsStartNewSpans().
// If a span is created, it is associated to the given context.Context object before returing.
func (t *baseTracer) newOrChildSpan(
	ctx context.Context,
	spanCtx opentracing.SpanContext,
	opName string,
) (opentracing.Span, context.Context) {
	var sp opentracing.Span
	if spanCtx != nil {
		// Start child span
		sp = opentracing.StartSpan(opName, opentracing.ChildOf(spanCtx))
	} else if t.IsStartNewSpans() {
		// There is no current span info, create a new one if permitted
		sp = opentracing.StartSpan(opName)
	}
	if sp != nil {
		// A (new / child) span has been created, add it to the context
		ctx = opentracing.ContextWithSpan(ctx, sp)
	}
	return sp, ctx
}

// globalTracer is initialised to a Nop tracer, calling InitGlobalTracer will reset this
var globalTracer = newNopTracer()

// InitGlobalTracer creates a new Jaeger tracer, and sets it as global tracer.
func InitGlobalTracer(name string, jaegerCfg *config.JaegerConfig) (io.Closer, error) {
	// If jaeger config has been set and the tracing enabled, initialise the JaegerTracer
	if jaegerCfg != nil && jaegerCfg.Enabled {
		globalTracer = newJaegerTracer()
	}
	// Initialise the tracer
	return globalTracer.InitGlobalTracer(name, jaegerCfg)
}

// Glob returns the global tracer
func Glob() Tracer {
	return globalTracer
}

// SetGlob sets the global tracer, for testing
func SetGlob(t Tracer) {
	globalTracer = t
}
