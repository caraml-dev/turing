package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// NopTracer implements the Tracer interface with dummy methods
type NopTracer struct{}

// IsEnabled satisfies the Tracer interface, always returning false
func (*NopTracer) IsEnabled() bool {
	return false
}

// StartSpanFromRequestHeader satisfies the Tracer interface, returning the context as
// is and a no-op span
func (*NopTracer) StartSpanFromRequestHeader(
	ctx context.Context,
	_ string,
	_ http.Header,
) (trace.Span, context.Context) {
	return trace.SpanFromContext(ctx), ctx
}

// StartSpanFromContext satisfies the Tracer interface, returning the context as is
// and a no-op span
func (*NopTracer) StartSpanFromContext(
	ctx context.Context,
	_ string,
) (trace.Span, context.Context) {
	return trace.SpanFromContext(ctx), ctx
}

func newNopTracer() Tracer {
	return &NopTracer{}
}
