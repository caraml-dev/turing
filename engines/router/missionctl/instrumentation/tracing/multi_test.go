package tracing

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// recordingTracer is a minimal Tracer that records every span it starts via an
// in-memory exporter, so tests can assert on what a MultiTracer produced.
type recordingTracer struct {
	tracer   trace.Tracer
	exporter *tracetest.InMemoryExporter
}

func newRecordingTracer(name string) *recordingTracer {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	return &recordingTracer{tracer: tp.Tracer(name), exporter: exporter}
}

func (*recordingTracer) IsEnabled() bool { return true }

func (r *recordingTracer) StartSpanFromContext(ctx context.Context, opName string) (trace.Span, context.Context) {
	ctx, span := r.tracer.Start(ctx, opName)
	return span, ctx
}

func (r *recordingTracer) StartSpanFromRequestHeader(
	ctx context.Context,
	opName string,
	_ http.Header,
) (trace.Span, context.Context) {
	return r.StartSpanFromContext(ctx, opName)
}

func TestMultiTracer_IsEnabled(t *testing.T) {
	mt := newMultiTracer([]Tracer{newRecordingTracer("a"), newRecordingTracer("b")})
	assert.Equal(t, true, mt.IsEnabled())
}

func TestMultiTracer_StartSpanFromContext_RecordsOnEveryBackend(t *testing.T) {
	a := newRecordingTracer("a")
	b := newRecordingTracer("b")
	mt := newMultiTracer([]Tracer{a, b})

	span, ctx := mt.StartSpanFromContext(context.Background(), "test-op")
	require.NotNil(t, span)
	require.NotNil(t, ctx)
	span.End()

	require.Len(t, a.exporter.GetSpans(), 1)
	require.Len(t, b.exporter.GetSpans(), 1)
	assert.Equal(t, "test-op", a.exporter.GetSpans()[0].Name)
	assert.Equal(t, "test-op", b.exporter.GetSpans()[0].Name)
}

func TestMultiTracer_NestedSpans_EachBackendKeepsItsOwnParentChain(t *testing.T) {
	a := newRecordingTracer("a")
	b := newRecordingTracer("b")
	mt := newMultiTracer([]Tracer{a, b})

	outerSpan, ctx := mt.StartSpanFromContext(context.Background(), "outer")
	innerSpan, ctx := mt.StartSpanFromContext(ctx, "inner")
	innerSpan.End()
	outerSpan.End()
	_ = ctx

	for _, exp := range []*tracetest.InMemoryExporter{a.exporter, b.exporter} {
		spans := exp.GetSpans()
		require.Len(t, spans, 2)
		var outer, inner tracetest.SpanStub
		for _, s := range spans {
			if s.Name == "outer" {
				outer = s
			} else {
				inner = s
			}
		}
		assert.Equal(t, outer.SpanContext.SpanID(), inner.Parent.SpanID())
		assert.Equal(t, outer.SpanContext.TraceID(), inner.SpanContext.TraceID())
	}
}
