package rpc

import (
	"github.com/hashicorp/go-plugin"

	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/manager"
	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/runner"
)

const (
	ManagerPluginIdentifier = "experiment_manager"
	RunnerPluginIdentifier  = "experiment_runner"
)

var (
	handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "EXPERIMENTS_PLUGIN",
		MagicCookieValue: "turing",
	}

	pluginMap = map[string]plugin.Plugin{
		ManagerPluginIdentifier: &manager.ExperimentManagerPlugin{},
		RunnerPluginIdentifier:  &runner.ExperimentRunnerPlugin{},
	}
)
