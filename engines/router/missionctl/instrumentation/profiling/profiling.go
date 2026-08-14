package profiling

import (
	"fmt"

	"github.com/grafana/pyroscope-go"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

const applicationName = "turing-router"

// Start starts continuous profiling via pyroscope-go if enabled in cfg. All router
// deployments report under the same Pyroscope application name and are differentiated
// by the router_name tag. Returns a nil profiler and nil error when profiling is
// disabled or cfg is nil. pyroscope.Start does not itself error on an empty
// ServerAddress -- it happily constructs a client that fails silently on every upload --
// so an empty address is rejected explicitly here instead.
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
		Tags:            map[string]string{"router_name": routerName},
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
