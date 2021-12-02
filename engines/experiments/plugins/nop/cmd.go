package main

import (
	experiments "github.com/gojek/turing/engines/experiment/v2"
	"github.com/gojek/turing/engines/experiment/v2/manager"
	"github.com/hashicorp/go-plugin"
)

func main() {
	nopManager := &ExperimentManager{}

	pluginMap := map[string]plugin.Plugin{
		experiments.ManagerPluginIdentifier: &manager.ExperimentManagerPlugin{Impl: nopManager},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: experiments.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
