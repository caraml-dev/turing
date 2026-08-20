package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

func TestGetGlob(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	globalTracer = &NopTracer{}
	assert.Equal(t, globalTracer, Glob())
}

func TestSetGlob(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	tr := &NopTracer{}
	SetGlob(tr)
	assert.Equal(t, tr, globalTracer)
}

func TestInitGlobalTracer_Nop(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	_, err := InitGlobalTracer("test", &config.JaegerConfig{}, &config.OtelConfig{}) //nolint:staticcheck
	assert.NoError(t, err)
	assert.Equal(t, false, globalTracer.IsEnabled())
}

func TestInitGlobalTracer_OtelOnly(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	shutdown, err := InitGlobalTracer("test", &config.JaegerConfig{}, &config.OtelConfig{ //nolint:staticcheck
		Enabled:           true,
		CollectorEndpoint: "http://localhost:4318",
	})
	require.NoError(t, err)
	assert.Equal(t, true, globalTracer.IsEnabled())
	assert.IsType(t, &OtelTracer{}, globalTracer)
	assert.NoError(t, shutdown(context.Background()))
}

// TestInitGlobalTracer_JaegerOnly relies on the classic Jaeger client's UDP agent
// reporter not dialling eagerly -- InitGlobalTracer succeeds even with no agent
// listening at the configured host:port, exactly like the OTel exporter today (see
// TestNewOtelTracer in otel_test.go).
func TestInitGlobalTracer_JaegerOnly(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	shutdown, err := InitGlobalTracer("test", &config.JaegerConfig{ //nolint:staticcheck
		Enabled:           true,
		ReporterAgentHost: "localhost",
		ReporterAgentPort: 6831,
	}, &config.OtelConfig{})
	require.NoError(t, err)
	assert.Equal(t, true, globalTracer.IsEnabled())
	assert.IsType(t, &JaegerTracer{}, globalTracer)
	assert.NoError(t, shutdown(context.Background()))
}

func TestInitGlobalTracer_Multi(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	shutdown, err := InitGlobalTracer("test",
		&config.JaegerConfig{ //nolint:staticcheck
			Enabled:           true,
			ReporterAgentHost: "localhost",
			ReporterAgentPort: 6831,
		},
		&config.OtelConfig{
			Enabled:           true,
			CollectorEndpoint: "http://localhost:4318",
		},
	)
	require.NoError(t, err)
	assert.Equal(t, true, globalTracer.IsEnabled())
	assert.IsType(t, &MultiTracer{}, globalTracer)
	assert.NoError(t, shutdown(context.Background()))
}

func TestInitGlobalTracer_OtelError_ReturnsNonNilShutdown(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	shutdown, err := InitGlobalTracer("test", &config.JaegerConfig{}, &config.OtelConfig{ //nolint:staticcheck
		Enabled:           true,
		CollectorEndpoint: "",
	})
	assert.Error(t, err)
	require.NotNil(t, shutdown)
	assert.NoError(t, shutdown(context.Background()))
}
