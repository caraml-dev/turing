package profiling

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildTags(t *testing.T) {
	t.Run("only router_name when pod env vars are unset", func(t *testing.T) {
		assert.Equal(t, map[string]string{"router_name": "test-router"}, buildTags("test-router", nil, true))
	})

	t.Run("includes pod_name and pod_namespace when set and includePodTags is true", func(t *testing.T) {
		t.Setenv(envPodName, "test-router-abc123")
		t.Setenv(envPodNamespace, "test-namespace")

		assert.Equal(t, map[string]string{
			"router_name":   "test-router",
			"pod_name":      "test-router-abc123",
			"pod_namespace": "test-namespace",
		}, buildTags("test-router", nil, true))
	})

	t.Run("omits pod_name and pod_namespace when includePodTags is false, even if set", func(t *testing.T) {
		t.Setenv(envPodName, "test-router-abc123")
		t.Setenv(envPodNamespace, "test-namespace")

		assert.Equal(t, map[string]string{"router_name": "test-router"}, buildTags("test-router", nil, false))
	})

	t.Run("merges in customTags", func(t *testing.T) {
		assert.Equal(t, map[string]string{
			"router_name": "test-router",
			"team":        "fraud",
			"env":         "staging",
		}, buildTags("test-router", map[string]string{"team": "fraud", "env": "staging"}, true))
	})

	t.Run("built-in tags win over a colliding customTags key", func(t *testing.T) {
		t.Setenv(envPodName, "test-router-abc123")

		assert.Equal(t, map[string]string{
			"router_name": "test-router",
			"pod_name":    "test-router-abc123",
		}, buildTags("test-router", map[string]string{
			"router_name": "should-be-overridden",
			"pod_name":    "should-also-be-overridden",
		}, true))
	})
}
