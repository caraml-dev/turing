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

func TestInitProfiler_EnabledWithCustomTags(t *testing.T) {
	profiler, err := initProfiler(config.PyroscopeConfig{
		Enabled:       true,
		ServerAddress: "http://localhost:4040",
		CustomTags:    map[string]string{"team": "fraud"},
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

func TestBuildTags(t *testing.T) {
	t.Run("empty when pod env vars are unset and no custom tags", func(t *testing.T) {
		assert.Equal(t, map[string]string{}, buildTags(config.PyroscopeConfig{IncludePodTags: true}))
	})

	t.Run("includes pod_name and pod_namespace when set and IncludePodTags is true", func(t *testing.T) {
		t.Setenv(envPodName, "turing-api-abc123")
		t.Setenv(envPodNamespace, "test-namespace")

		assert.Equal(t, map[string]string{
			"pod_name":      "turing-api-abc123",
			"pod_namespace": "test-namespace",
		}, buildTags(config.PyroscopeConfig{IncludePodTags: true}))
	})

	t.Run("omits pod_name and pod_namespace when IncludePodTags is false, even if set", func(t *testing.T) {
		t.Setenv(envPodName, "turing-api-abc123")
		t.Setenv(envPodNamespace, "test-namespace")

		assert.Equal(t, map[string]string{}, buildTags(config.PyroscopeConfig{IncludePodTags: false}))
	})

	t.Run("merges in CustomTags", func(t *testing.T) {
		assert.Equal(t, map[string]string{"team": "fraud", "env": "staging"}, buildTags(config.PyroscopeConfig{
			CustomTags: map[string]string{"team": "fraud", "env": "staging"},
		}))
	})

	t.Run("built-in pod tags win over a colliding CustomTags key", func(t *testing.T) {
		t.Setenv(envPodName, "turing-api-abc123")

		assert.Equal(t, map[string]string{"pod_name": "turing-api-abc123"}, buildTags(config.PyroscopeConfig{
			IncludePodTags: true,
			CustomTags:     map[string]string{"pod_name": "should-be-overridden"},
		}))
	})
}
