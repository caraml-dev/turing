package profiling_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/profiling"
)

func TestStart_Disabled(t *testing.T) {
	profiler, err := profiling.Start("test-router", &config.PyroscopeConfig{Enabled: false})
	require.NoError(t, err)
	assert.Nil(t, profiler)
}

func TestStart_NilConfig(t *testing.T) {
	profiler, err := profiling.Start("test-router", nil)
	require.NoError(t, err)
	assert.Nil(t, profiler)
}

func TestStart_Enabled(t *testing.T) {
	profiler, err := profiling.Start("test-router", &config.PyroscopeConfig{
		Enabled:       true,
		ServerAddress: "http://localhost:4040",
	})
	require.NoError(t, err)
	require.NotNil(t, profiler)
	defer func() { _ = profiler.Stop() }()
}

func TestStart_EnabledWithHTTPHeaders(t *testing.T) {
	profiler, err := profiling.Start("test-router", &config.PyroscopeConfig{
		Enabled:       true,
		ServerAddress: "http://localhost:4040",
		HTTPHeaders:   map[string]string{"Authorization": "Bearer token"},
	})
	require.NoError(t, err)
	require.NotNil(t, profiler)
	defer func() { _ = profiler.Stop() }()
}

func TestStart_EnabledEmptyServerAddress(t *testing.T) {
	profiler, err := profiling.Start("test-router", &config.PyroscopeConfig{
		Enabled:       true,
		ServerAddress: "",
	})
	require.Error(t, err)
	assert.Nil(t, profiler)
}
