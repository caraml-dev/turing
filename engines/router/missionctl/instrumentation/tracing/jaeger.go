package tracing

import (
	"context"
	"fmt"
	"net/http"

	ot "github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

// JaegerTracer implements the Tracer interface using the classic Jaeger client
// library (OpenTracing API, Thrift transport over a UDP agent or an HTTP
// collector). Every trace.Tracer/trace.Span call made by the rest of this
// codebase is translated into the equivalent ot.Tracer/ot.Span (OpenTracing)
// call by this file, so spans are still physically created by the classic
// Jaeger client and exported over Thrift, while callers keep working against
// the same trace.Span type regardless of which backend is active.
//
// Note: go.opentelemetry.io/otel/bridge/opentracing only bridges in the
// opposite direction to what's needed here -- it forwards OpenTracing-API
// calls onto an OpenTelemetry trace.Tracer of choice (see its NewTracerPair,
// which takes a trace.Tracer, not an ot.Tracer), so it cannot adapt an
// existing ot.Tracer (the classic Jaeger client) into a trace.Tracer. This
// file does that translation directly instead.
//
// Deprecated: this backend exists only to support downstream systems that
// still require Jaeger's native Thrift ingestion; it will be removed in a
// future release. New deployments should use OtelTracer (OtelConfig) instead.
type JaegerTracer struct {
	tracer     ot.Tracer
	propagator propagation.TextMapPropagator
}

// newJaegerTracer builds a classic Jaeger client from cfg (sampling all requests,
// matching this backend's pre-OTel-migration behaviour, and generating 128-bit
// trace IDs so they line up byte-for-byte with the OTel trace.TraceID format used
// elsewhere in this codebase), and returns a Tracer that adapts it to the
// trace.Tracer/trace.Span API, alongside a ShutdownFunc that closes the underlying
// Jaeger reporter.
func newJaegerTracer(name string, cfg *config.JaegerConfig) (Tracer, ShutdownFunc, error) {
	jCfg := jaegercfg.Configuration{
		ServiceName: name,
		Disabled:    !cfg.Enabled,
		Gen128Bit:   true,
		Reporter: &jaegercfg.ReporterConfig{
			CollectorEndpoint:  cfg.CollectorEndpoint,
			LocalAgentHostPort: fmt.Sprintf("%s:%d", cfg.ReporterAgentHost, cfg.ReporterAgentPort),
			LogSpans:           true,
		},
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}

	nativeTracer, closer, err := jCfg.NewTracer()
	if err != nil {
		return nil, nil, err
	}

	t := &JaegerTracer{
		tracer: nativeTracer,
		// Reuse the same propagator OtelTracer uses for incoming header extraction,
		// so both backends agree on how a remote parent span context is decoded.
		propagator: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{}, b3.New(),
		),
	}

	return t, func(context.Context) error { return closer.Close() }, nil
}

// IsEnabled satisfies the Tracer interface, always returning true
func (*JaegerTracer) IsEnabled() bool {
	return true
}

// StartSpanFromRequestHeader extracts a remote span context from the request header
// and starts a child span from it.
func (t *JaegerTracer) StartSpanFromRequestHeader(
	ctx context.Context,
	opName string,
	header http.Header,
) (trace.Span, context.Context) {
	ctx = t.propagator.Extract(ctx, propagation.HeaderCarrier(header))
	return t.startSpan(ctx, opName)
}

// StartSpanFromContext starts a new / child span associated with the given context.
func (t *JaegerTracer) StartSpanFromContext(
	ctx context.Context,
	opName string,
) (trace.Span, context.Context) {
	return t.startSpan(ctx, opName)
}

// startSpan starts a native Jaeger span, using whatever OTel trace.SpanContext is
// already present in ctx (either injected by StartSpanFromRequestHeader above, or
// left there by a previous call to this same JaegerTracer) as its ChildOf parent,
// so the classic Jaeger client's own trace/span IDs stay derived from -- and thus
// consistent with -- the trace.SpanContext exposed to the rest of this codebase.
func (t *JaegerTracer) startSpan(ctx context.Context, opName string) (trace.Span, context.Context) {
	var opts []ot.StartSpanOption
	if parent := trace.SpanContextFromContext(ctx); parent.IsValid() {
		if parentSC, err := jaegerSpanContextFrom(parent); err == nil {
			opts = append(opts, ot.ChildOf(parentSC))
		}
	}

	span := &jaegerSpan{span: t.tracer.StartSpan(opName, opts...)}
	return span, trace.ContextWithSpan(ctx, span)
}

// jaegerSpanContextFrom converts an OTel trace.SpanContext into the equivalent
// jaeger.SpanContext, so a span started from an OTel-style parent (extracted from
// request headers, or carried over from a previous jaegerSpan already in ctx)
// keeps the same trace ID under the classic Jaeger client too.
func jaegerSpanContextFrom(sc trace.SpanContext) (jaeger.SpanContext, error) {
	traceID, err := jaeger.TraceIDFromString(sc.TraceID().String())
	if err != nil {
		return jaeger.SpanContext{}, err
	}
	spanID, err := jaeger.SpanIDFromString(sc.SpanID().String())
	if err != nil {
		return jaeger.SpanContext{}, err
	}
	return jaeger.NewSpanContext(traceID, spanID, 0, sc.IsSampled(), nil), nil
}

// jaegerSpan adapts a native Jaeger (OpenTracing) ot.Span to the OTel trace.Span
// interface, so every backend in this package exposes the same span type to callers.
type jaegerSpan struct {
	embedded.Span

	span ot.Span
}

// End satisfies trace.Span, finishing the underlying native Jaeger span.
func (s *jaegerSpan) End(...trace.SpanEndOption) {
	s.span.Finish()
}

// AddEvent satisfies trace.Span, logging name and any attributes as a structured
// log entry on the underlying native Jaeger span.
func (s *jaegerSpan) AddEvent(name string, options ...trace.EventOption) {
	cfg := trace.NewEventConfig(options...)
	attrs := cfg.Attributes()

	kv := make([]interface{}, 0, 2+2*len(attrs))
	kv = append(kv, "event", name)
	for _, attr := range attrs {
		kv = append(kv, string(attr.Key), attr.Value.AsInterface())
	}
	s.span.LogKV(kv...)
}

// IsRecording satisfies trace.Span. The classic Jaeger client does not expose
// whether a given span will actually be sampled/reported until Finish(), so this
// conservatively always returns true.
func (s *jaegerSpan) IsRecording() bool {
	return true
}

// RecordError satisfies trace.Span, tagging the underlying native Jaeger span as
// an error and logging err on it.
func (s *jaegerSpan) RecordError(err error, _ ...trace.EventOption) {
	otext.LogError(s.span, err)
}

// SpanContext satisfies trace.Span, translating the underlying native Jaeger
// span's own SpanContext into the equivalent OTel trace.SpanContext.
func (s *jaegerSpan) SpanContext() trace.SpanContext {
	sc, ok := s.span.Context().(jaeger.SpanContext)
	if !ok {
		return trace.SpanContext{}
	}

	// jaeger.TraceID.String() only zero-pads to 32 hex chars when High != 0 (i.e.
	// when the trace was actually generated as 128-bit); pad it out so it always
	// parses as a valid OTel TraceID.
	traceIDHex := sc.TraceID().String()
	for len(traceIDHex) < 32 {
		traceIDHex = "0" + traceIDHex
	}

	traceID, err := trace.TraceIDFromHex(traceIDHex)
	if err != nil {
		return trace.SpanContext{}
	}
	spanID, err := trace.SpanIDFromHex(sc.SpanID().String())
	if err != nil {
		return trace.SpanContext{}
	}

	var flags trace.TraceFlags
	if sc.IsSampled() {
		flags = trace.FlagsSampled
	}

	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: flags,
	})
}

// SetStatus satisfies trace.Span, tagging the underlying native Jaeger span with
// the OTel status code/description.
func (s *jaegerSpan) SetStatus(code codes.Code, description string) {
	if code == codes.Error {
		s.span.SetTag("error", true)
	}
	s.span.SetTag("otel.status_code", code.String())
	if description != "" {
		s.span.SetTag("otel.status_description", description)
	}
}

// SetName satisfies trace.Span, changing the underlying native Jaeger span's
// operation name.
func (s *jaegerSpan) SetName(name string) {
	s.span.SetOperationName(name)
}

// SetAttributes satisfies trace.Span, setting each attribute as a tag on the
// underlying native Jaeger span.
func (s *jaegerSpan) SetAttributes(kv ...attribute.KeyValue) {
	for _, attr := range kv {
		s.span.SetTag(string(attr.Key), attr.Value.AsInterface())
	}
}

// TracerProvider satisfies trace.Span. This backend doesn't construct an OTel
// SDK TracerProvider (spans are created directly against the native Jaeger
// ot.Tracer), so this returns a no-op provider.
func (s *jaegerSpan) TracerProvider() trace.TracerProvider {
	return noop.NewTracerProvider()
}
