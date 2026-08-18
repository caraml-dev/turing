package profiling

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildTags(t *testing.T) {
	t.Run("only router_name when pod env vars are unset", func(t *testing.T) {
		assert.Equal(t, map[string]string{"router_name": "test-router"}, buildTags("test-router", true))
	})

	t.Run("includes pod_name and pod_namespace when set and includePodTags is true", func(t *testing.T) {
		t.Setenv(envPodName, "test-router-abc123")
		t.Setenv(envPodNamespace, "test-namespace")

		assert.Equal(t, map[string]string{
			"router_name":   "test-router",
			"pod_name":      "test-router-abc123",
			"pod_namespace": "test-namespace",
		}, buildTags("test-router", true))
	})

	t.Run("omits pod_name and pod_namespace when includePodTags is false, even if set", func(t *testing.T) {
		t.Setenv(envPodName, "test-router-abc123")
		t.Setenv(envPodNamespace, "test-namespace")

		assert.Equal(t, map[string]string{"router_name": "test-router"}, buildTags("test-router", false))
	})
}
