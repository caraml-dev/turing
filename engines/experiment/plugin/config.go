package plugin

import (
	"github.com/gojek/turing/engines/experiment/plugin/manager"
	"github.com/gojek/turing/engines/experiment/plugin/runner"
	"github.com/hashicorp/go-plugin"
)

const (
	ManagerPluginIdentified = "experiment_manager"
	RunnerPluginIdentified  = "experiment_runner"
)

var (
	handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "EXPERIMENTS_PLUGIN",
		MagicCookieValue: "turing",
	}

	pluginMap = map[string]plugin.Plugin{
		ManagerPluginIdentified: &manager.ExperimentManagerPlugin{},
		RunnerPluginIdentified:  &runner.ExperimentRunnerPlugin{},
	}
)
