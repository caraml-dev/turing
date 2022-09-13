package tracing

import (
	"context"
	"io"
	"net/http"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/opentracing/opentracing-go"
)

// NopTracer implements the Tracer interface with dummy methods
type NopTracer struct{}

// IsStartNewSpans satisfies the Tracer interface, always returning false
func (*NopTracer) IsStartNewSpans() bool {
	return false
}

// SetStartNewSpans is a nop method that satisfies the Tracer interface
func (*NopTracer) SetStartNewSpans(_ bool) {}

// InitGlobalTracer satisfies the Tracer interface and returns a Nop closer
func (*NopTracer) InitGlobalTracer(_ string, _ *config.JaegerConfig) (io.Closer, error) {
	return io.NopCloser(nil), nil
}

// IsEnabled satisfies the Tracer interface, always returning false
func (*NopTracer) IsEnabled() bool {
	return false
}

// StartSpanFromRequestHeader satisfies the Tracer interface, returning the context as
// is and an empty span
func (*NopTracer) StartSpanFromRequestHeader(
	ctx context.Context,
	_ string,
	_ http.Header,
) (opentracing.Span, context.Context) {
	return nil, ctx
}

// StartSpanFromContext satisfies the Tracer interface, returning the context as is
// and an empty span
func (*NopTracer) StartSpanFromContext(
	ctx context.Context,
	_ string,
) (opentracing.Span, context.Context) {
	return nil, ctx
}

func newNopTracer() Tracer {
	return &NopTracer{}
}
