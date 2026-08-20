package tracing

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

func TestNewOtelTracer_IsEnabled(t *testing.T) {
	tr, shutdown, err := newOtelTracer("test", &config.OtelConfig{
		Enabled:           true,
		CollectorEndpoint: "http://localhost:4318",
	})
	require.NoError(t, err)
	defer func() { _ = shutdown(context.Background()) }()

	assert.Equal(t, true, tr.IsEnabled())
}

func TestNewOtelTracer(t *testing.T) {
	tr, shutdown, err := newOtelTracer("test", &config.OtelConfig{
		Enabled:           true,
		CollectorEndpoint: "http://localhost:4318",
		SamplingRatio:     0.5,
	})
	require.NoError(t, err)
	require.NotNil(t, tr)
	require.NotNil(t, shutdown)

	defer func() { _ = shutdown(context.Background()) }()
}

func TestNewOtelTracer_HTTPS(t *testing.T) {
	tr, shutdown, err := newOtelTracer("test", &config.OtelConfig{
		Enabled:           true,
		CollectorEndpoint: "https://localhost:4318",
		SamplingRatio:     0.5,
	})
	require.NoError(t, err)
	defer func() { _ = shutdown(context.Background()) }()

	span, ctx := tr.StartSpanFromContext(context.Background(), "test-op")
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	span.End()
}

func TestOtelTracer_StartSpanFromContext(t *testing.T) {
	tr, shutdown, err := newOtelTracer("test", &config.OtelConfig{
		Enabled:           true,
		CollectorEndpoint: "http://localhost:4318",
	})
	require.NoError(t, err)
	defer func() { _ = shutdown(context.Background()) }()

	span, ctx := tr.StartSpanFromContext(context.Background(), "test-op")
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	span.End()
}

func TestOtelTracer_StartSpanFromRequestHeader(t *testing.T) {
	tr, shutdown, err := newOtelTracer("test", &config.OtelConfig{
		Enabled:           true,
		CollectorEndpoint: "http://localhost:4318",
	})
	require.NoError(t, err)
	defer func() { _ = shutdown(context.Background()) }()

	header := http.Header{}
	header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	span, ctx := tr.StartSpanFromRequestHeader(context.Background(), "test-op", header)
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", span.SpanContext().TraceID().String())
	span.End()
}

// TestOtelTracer_StartSpanFromRequestHeader_B3 confirms that inbound requests carrying B3
// trace-context headers (Istio/Knative's default propagation format) are joined into the
// same trace, rather than silently producing a disconnected root span.
func TestOtelTracer_StartSpanFromRequestHeader_B3(t *testing.T) {
	tr, shutdown, err := newOtelTracer("test", &config.OtelConfig{
		Enabled:           true,
		CollectorEndpoint: "http://localhost:4318",
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
	assert.True(t, span.SpanContext().IsSampled())
	span.End()
}

func TestNewOtelTracer_MissingHost(t *testing.T) {
	_, _, err := newOtelTracer("test", &config.OtelConfig{
		Enabled:           true,
		CollectorEndpoint: "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing host")
}
