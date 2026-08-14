package tracing

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

func TestNewJaegerTracer_IsEnabled(t *testing.T) {
	tr, shutdown, err := newJaegerTracer("test", &config.JaegerConfig{
		Enabled:           true,
		ReporterAgentHost: "localhost",
		ReporterAgentPort: 6831,
	})
	require.NoError(t, err)
	defer func() { _ = shutdown(context.Background()) }()

	assert.Equal(t, true, tr.IsEnabled())
}

func TestNewJaegerTracer_StartSpanFromContext(t *testing.T) {
	tr, shutdown, err := newJaegerTracer("test", &config.JaegerConfig{
		Enabled:           true,
		ReporterAgentHost: "localhost",
		ReporterAgentPort: 6831,
	})
	require.NoError(t, err)
	defer func() { _ = shutdown(context.Background()) }()

	span, ctx := tr.StartSpanFromContext(context.Background(), "test-op")
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	span.End()
}

func TestNewJaegerTracer_NestedSpans_ShareTraceID(t *testing.T) {
	tr, shutdown, err := newJaegerTracer("test", &config.JaegerConfig{
		Enabled:           true,
		ReporterAgentHost: "localhost",
		ReporterAgentPort: 6831,
	})
	require.NoError(t, err)
	defer func() { _ = shutdown(context.Background()) }()

	outerSpan, ctx := tr.StartSpanFromContext(context.Background(), "outer")
	innerSpan, _ := tr.StartSpanFromContext(ctx, "inner")

	assert.Equal(t, outerSpan.SpanContext().TraceID(), innerSpan.SpanContext().TraceID())
	innerSpan.End()
	outerSpan.End()
}

func TestNewJaegerTracer_StartSpanFromRequestHeader(t *testing.T) {
	tr, shutdown, err := newJaegerTracer("test", &config.JaegerConfig{
		Enabled:           true,
		ReporterAgentHost: "localhost",
		ReporterAgentPort: 6831,
	})
	require.NoError(t, err)
	defer func() { _ = shutdown(context.Background()) }()

	header := http.Header{}
	header.Set("X-B3-Traceid", "4bf92f3577b34da6a3ce929d0e0e4736")
	header.Set("X-B3-Spanid", "00f067aa0ba902b7")
	header.Set("X-B3-Sampled", "1")

	span, ctx := tr.StartSpanFromRequestHeader(context.Background(), "test-op", header)
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", span.SpanContext().TraceID().String())
	span.End()
}
