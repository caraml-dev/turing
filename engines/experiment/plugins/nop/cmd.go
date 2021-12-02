package main

import (
	"github.com/gojek/turing/engines/experiment"
	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/hashicorp/go-plugin"
)

func main() {
	nopManager := &ExperimentManager{}

	pluginMap := map[string]plugin.Plugin{
		experiment.ManagerPluginIdentifier: &manager.ExperimentManagerPlugin{Impl: nopManager},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: experiment.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
