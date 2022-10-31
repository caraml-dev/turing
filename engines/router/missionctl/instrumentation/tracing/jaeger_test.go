package tracing

import (
	"context"
	"net/http"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

func TestIsEnabled(t *testing.T) {
	tr := newJaegerTracer()
	assert.Equal(t, true, tr.IsEnabled())
}

func TestStartSpanFromRequestHeader(t *testing.T) {
	tr := newJaegerTracer()

	// Init global tracer using Jaeger client
	_, _ = tr.InitGlobalTracer("test", &config.JaegerConfig{
		Enabled:           true,
		StartNewSpans:     false,
		ReporterAgentHost: "localhost",
		ReporterAgentPort: 6832,
	})

	// Set span related attributes to request header
	header := http.Header{}
	header.Set("X-B3-Sampled", "1")
	header.Set("X-B3-Spanid", "a30ec88c39471716")
	header.Set("X-B3-Traceid", "950f2de0b8430e9fa30ec88c39471716")
	header.Set("X-Request-Id", "26787b1c-bf6e-97a8-8b30-36675e4effa0")

	// Verify that a span can be extracted and a child span created
	sp, _ := tr.StartSpanFromRequestHeader(context.Background(), "test", header)
	assert.NotNil(t, sp)
}

func TestStartSpanFromContext(t *testing.T) {
	tr := newJaegerTracer()

	// Init global tracer using Jaeger client
	_, _ = tr.InitGlobalTracer("test", &config.JaegerConfig{
		Enabled:           true,
		StartNewSpans:     false,
		ReporterAgentHost: "localhost",
		ReporterAgentPort: 6832,
	})

	// Associate a new span to a context
	_, ctx := opentracing.StartSpanFromContext(context.Background(), "test")
	// Verify that a span can be extracted and a child span created
	sp, _ := tr.StartSpanFromContext(ctx, "test")
	assert.NotNil(t, sp)
}

func TestBuildConfig(t *testing.T) {
	cfg := &config.JaegerConfig{
		Enabled:           true,
		StartNewSpans:     true,
		ReporterAgentHost: "localhost",
		ReporterAgentPort: 2000,
		CollectorEndpoint: "test_endpoint",
	}

	jaegerCfg := buildConfig(cfg)

	// Test jaeger config values
	assert.Equal(t, false, jaegerCfg.Disabled)
	assert.Equal(t, "test_endpoint", jaegerCfg.Reporter.CollectorEndpoint)
	assert.Equal(t, "localhost:2000", jaegerCfg.Reporter.LocalAgentHostPort)
	assert.Equal(t, true, jaegerCfg.Reporter.LogSpans)
	assert.Equal(t, "const", jaegerCfg.Sampler.Type)
	assert.Equal(t, float64(1), jaegerCfg.Sampler.Param)
}
