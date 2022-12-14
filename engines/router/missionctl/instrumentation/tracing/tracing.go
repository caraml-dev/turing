package tracing

import (
	"context"
	"io"
	"net/http"

	"github.com/opentracing/opentracing-go"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

// Tracer represents a generic tracer that supports initialization of a global
// tracing client and creation of opentracing spans
type Tracer interface {
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
	var sp opentracing.Span
	if spanCtx != nil {
		// Start child span
		sp = opentracing.StartSpan(opName, opentracing.ChildOf(spanCtx))
		if sp != nil {
			// A (new / child) span has been created, add it to the context
			ctx = opentracing.ContextWithSpan(ctx, sp)
		}
	}
	return sp, ctx
}

// StartSpanFromContext attempts to extract span info from the given context.Context and creates
// a new / child span accordingly, which is associated to the same context object.
func (t *baseTracer) StartSpanFromContext(
	ctx context.Context,
	opName string,
) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContext(ctx, opName)
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
