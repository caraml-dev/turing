package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

func TestGetTracingConfig(t *testing.T) {
	cfg := tracingConfig{
		startNewSpans: true,
	}

	assert.Equal(t, true, cfg.IsStartNewSpans())
}

func TestSetTracingConfig(t *testing.T) {
	cfg := tracingConfig{
		startNewSpans: true,
	}
	cfg.SetStartNewSpans(false)
	assert.Equal(t, false, cfg.startNewSpans)
}

func TestGetGlob(t *testing.T) {
	// Save globalTracer in a temp var and reset after the test
	tempTracer := globalTracer
	defer func() {
		globalTracer = tempTracer
	}()

	// Test
	globalTracer = &NopTracer{}
	assert.Equal(t, globalTracer, Glob())
}

func TestSetGlob(t *testing.T) {
	// Save globalTracer in a temp var and reset after the test
	tempTracer := globalTracer
	defer func() {
		globalTracer = tempTracer
	}()

	// Test
	tr := &NopTracer{}
	SetGlob(tr)
	assert.Equal(t, tr, globalTracer)
}

func TestInitGlobalTracerNop(t *testing.T) {
	// Save globalTracer in a temp var and reset after the test
	tempTracer := globalTracer
	defer func() {
		globalTracer = tempTracer
	}()

	_, err := InitGlobalTracer("test", &config.JaegerConfig{})
	assert.NoError(t, err)
	assert.Equal(t, false, globalTracer.IsEnabled())
}

func TestInitGlobalTracerJaeger(t *testing.T) {
	// Save globalTracer in a temp var and reset after the test
	tempTracer := globalTracer
	defer func() {
		globalTracer = tempTracer
	}()

	_, err := InitGlobalTracer("test", &config.JaegerConfig{
		Enabled:           true,
		ReporterAgentHost: "localhost",
		ReporterAgentPort: 1000,
	})
	assert.NoError(t, err)
	assert.Equal(t, true, globalTracer.IsEnabled())
}
