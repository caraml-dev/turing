package profiling

import (
	"fmt"
	"os"

	"github.com/grafana/pyroscope-go"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

const (
	applicationName = "turing-router"
	// envPodName and envPodNamespace are populated via the Kubernetes downward API by the
	// Turing API's servicebuilder. They are absent when running outside a pod (e.g. local dev).
	envPodName      = "POD_NAME"
	envPodNamespace = "POD_NAMESPACE"
)

// Start starts continuous profiling via pyroscope-go if enabled in cfg. All router
// deployments report under the same Pyroscope application name and are differentiated
// by the router_name tag, plus -- when cfg.IncludePodTags is true -- pod_name/pod_namespace
// tags populated from the POD_NAME and POD_NAMESPACE downward API env vars, to distinguish
// individual pods within a multi-replica router deployment. The pod tags are also omitted
// when those env vars are unset, so as not to report noisy empty-string tags outside a real
// pod. cfg.CustomTags are merged in on top of those, though router_name/pod_name/
// pod_namespace always win on key collision. Returns a nil profiler and nil error when
// profiling is disabled or cfg is nil.
// pyroscope.Start does not itself error on an empty ServerAddress -- it happily constructs a
// client that fails silently on every upload -- so an empty address is rejected explicitly
// here instead.
func Start(routerName string, cfg *config.PyroscopeConfig) (*pyroscope.Profiler, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, nil
	}
	if cfg.ServerAddress == "" {
		return nil, fmt.Errorf("pyroscope profiling is enabled but ServerAddress is empty")
	}

	return pyroscope.Start(pyroscope.Config{
		ApplicationName: applicationName,
		ServerAddress:   cfg.ServerAddress,
		HTTPHeaders:     cfg.HTTPHeaders,
		Tags:            buildTags(routerName, cfg.CustomTags, cfg.IncludePodTags),
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
		},
	})
}

// buildTags returns the static Pyroscope tags for this process: customTags first, then
// router_name always, plus pod_name/pod_namespace -- when includePodTags is true -- for
// whichever of the corresponding downward API env vars are set. router_name/pod_name/
// pod_namespace are applied last, so they always win over a colliding customTags key.
func buildTags(routerName string, customTags map[string]string, includePodTags bool) map[string]string {
	tags := make(map[string]string, len(customTags)+3)
	for k, v := range customTags {
		tags[k] = v
	}
	tags["router_name"] = routerName
	if includePodTags {
		if podName := os.Getenv(envPodName); podName != "" {
			tags["pod_name"] = podName
		}
		if podNamespace := os.Getenv(envPodNamespace); podNamespace != "" {
			tags["pod_namespace"] = podNamespace
		}
	}
	return tags
}
