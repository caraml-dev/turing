package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/api/turing/config"
)

func TestInitTracer_Disabled(t *testing.T) {
	shutdown, err := initTracer(config.OtelConfig{Enabled: false})
	require.NoError(t, err)
	require.NotNil(t, shutdown)
	assert.NoError(t, shutdown(context.Background()))
}

func TestInitTracer_Enabled(t *testing.T) {
	shutdown, err := initTracer(config.OtelConfig{
		Enabled:       true,
		OtlpEndpoint:  "http://localhost:4318",
		SamplingRatio: 0.5,
	})
	require.NoError(t, err)
	require.NotNil(t, shutdown)
	assert.NoError(t, shutdown(context.Background()))
}

func TestInitProfiler_Disabled(t *testing.T) {
	profiler, err := initProfiler(config.PyroscopeConfig{Enabled: false})
	require.NoError(t, err)
	assert.Nil(t, profiler)
}

func TestInitProfiler_Enabled(t *testing.T) {
	profiler, err := initProfiler(config.PyroscopeConfig{
		Enabled:       true,
		ServerAddress: "http://localhost:4040",
	})
	require.NoError(t, err)
	require.NotNil(t, profiler)
	defer func() { _ = profiler.Stop() }()
}

func TestInitProfiler_EnabledWithHTTPHeaders(t *testing.T) {
	profiler, err := initProfiler(config.PyroscopeConfig{
		Enabled:       true,
		ServerAddress: "http://localhost:4040",
		HTTPHeaders:   map[string]string{"Authorization": "Bearer token"},
	})
	require.NoError(t, err)
	require.NotNil(t, profiler)
	defer func() { _ = profiler.Stop() }()
}

func TestInitProfiler_EnabledEmptyServerAddress(t *testing.T) {
	profiler, err := initProfiler(config.PyroscopeConfig{
		Enabled:       true,
		ServerAddress: "",
	})
	require.Error(t, err)
	assert.Nil(t, profiler)
}
