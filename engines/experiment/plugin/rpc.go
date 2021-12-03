package plugin

import (
	"github.com/gojek/turing/engines/experiment/plugin/manager"
	"github.com/hashicorp/go-plugin"
)

func Serve(services *Services) {
	plugins := plugin.PluginSet{
		ManagerPluginIdentifier: &manager.ExperimentManagerPlugin{
			Impl: services.Manager,
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         plugins,
	})
}
