package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

// MultiTracer fans out span creation to every wrapped Tracer, so a single logical
// operation is exported to all of them. Used when more than one backend is enabled
// simultaneously (e.g. both JaegerConfig and OtelConfig).
type MultiTracer struct {
	tracers []Tracer
}

// newMultiTracer wraps tracers (expected to have at least 2 elements -- for exactly
// one enabled backend, InitGlobalTracer uses it directly instead) in a MultiTracer.
func newMultiTracer(tracers []Tracer) Tracer {
	return &MultiTracer{tracers: tracers}
}

// IsEnabled satisfies the Tracer interface, always returning true (MultiTracer is
// only constructed when at least one wrapped tracer is enabled).
func (*MultiTracer) IsEnabled() bool {
	return true
}

// multiParentsKey is the context key under which each backend's own previous span
// is tracked independently (see multiSpan's doc comment for why a single shared
// composite identity can't be used for this).
type multiParentsKey struct{}

// StartSpanFromContext starts a span on every wrapped tracer, each parented from
// its OWN previous span (tracked independently per backend via multiParentsKey,
// not the shared composite -- see multiSpan's doc comment for why that matters),
// and returns a composite span representing all of them.
func (m *MultiTracer) StartSpanFromContext(ctx context.Context, opName string) (trace.Span, context.Context) {
	parents, _ := ctx.Value(multiParentsKey{}).([]trace.Span)
	spans := make([]trace.Span, len(m.tracers))
	for i, t := range m.tracers {
		backendCtx := ctx
		// Guard against a length mismatch (e.g. if SetGlob swapped the global tracer,
		// changing the number of backends, mid-request): fall back to the plain ctx
		// for this backend rather than panicking on an out-of-range index.
		if parents != nil && len(parents) == len(m.tracers) {
			backendCtx = trace.ContextWithSpan(ctx, parents[i])
		}
		spans[i], _ = t.StartSpanFromContext(backendCtx, opName)
	}
	return m.newCompositeCtx(ctx, spans)
}

// StartSpanFromRequestHeader starts a span on every wrapped tracer, each extracting
// its own remote parent from header, and returns a composite span representing all
// of them. (No multiParentsKey lookup needed here: this is always the root of a
// request, so there's no prior per-backend parent to seed each backendCtx with --
// each backend independently extracts its own parent straight from header instead.)
func (m *MultiTracer) StartSpanFromRequestHeader(
	ctx context.Context,
	opName string,
	header http.Header,
) (trace.Span, context.Context) {
	spans := make([]trace.Span, len(m.tracers))
	for i, t := range m.tracers {
		spans[i], _ = t.StartSpanFromRequestHeader(ctx, opName, header)
	}
	return m.newCompositeCtx(ctx, spans)
}

// newCompositeCtx wraps spans in a multiSpan and seeds ctx with both the composite
// (for a caller's trace.SpanFromContext(ctx).End() to find) and the raw per-backend
// spans keyed by multiParentsKey (for the next nested call to hand each backend its
// own correct parent, per the doc comment above).
func (m *MultiTracer) newCompositeCtx(ctx context.Context, spans []trace.Span) (trace.Span, context.Context) {
	composite := &multiSpan{spans: spans}
	ctx = context.WithValue(ctx, multiParentsKey{}, spans)
	return composite, trace.ContextWithSpan(ctx, composite)
}

// multiSpan implements trace.Span by fanning every mutating call out to every
// wrapped span. SpanContext/IsRecording/TracerProvider delegate to the first
// wrapped span only -- but that is now purely about what identity the composite
// itself reports to any external caller that inspects it directly (e.g. logging
// the "current" trace ID). It no longer has anything to do with parent-chaining
// for a later nested StartSpanFromContext call: that lookup goes through
// multiParentsKey instead, precisely because delegating parent-lookup to spans[0]
// unconditionally would give every backend after the first the wrong parent (the
// first backend's span identity, not its own).
type multiSpan struct {
	embedded.Span

	spans []trace.Span
}

func (m *multiSpan) End(options ...trace.SpanEndOption) {
	for _, s := range m.spans {
		s.End(options...)
	}
}

func (m *multiSpan) AddEvent(name string, options ...trace.EventOption) {
	for _, s := range m.spans {
		s.AddEvent(name, options...)
	}
}

func (m *multiSpan) IsRecording() bool {
	return m.spans[0].IsRecording()
}

func (m *multiSpan) RecordError(err error, options ...trace.EventOption) {
	for _, s := range m.spans {
		s.RecordError(err, options...)
	}
}

func (m *multiSpan) SpanContext() trace.SpanContext {
	return m.spans[0].SpanContext()
}

func (m *multiSpan) SetStatus(code codes.Code, description string) {
	for _, s := range m.spans {
		s.SetStatus(code, description)
	}
}

func (m *multiSpan) SetName(name string) {
	for _, s := range m.spans {
		s.SetName(name)
	}
}

func (m *multiSpan) SetAttributes(kv ...attribute.KeyValue) {
	for _, s := range m.spans {
		s.SetAttributes(kv...)
	}
}

func (m *multiSpan) TracerProvider() trace.TracerProvider {
	return m.spans[0].TracerProvider()
}
