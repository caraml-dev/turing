package plugin

import (
	"github.com/gojek/turing/engines/experiment/plugin/manager"
	"github.com/gojek/turing/engines/experiment/plugin/runner"
	"github.com/hashicorp/go-plugin"
)

func Serve(services *Services) {
	plugins := plugin.PluginSet{
		ManagerPluginIdentifier: &manager.ExperimentManagerPlugin{
			Impl: services.Manager,
		},
		RunnerPluginIdentifier: &runner.ExperimentRunnerPlugin{
			Impl: services.Runner,
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         plugins,
	})
}
