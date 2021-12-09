package plugin

import (
	"fmt"
	"os/exec"
	"runtime"

	managerPlugin "github.com/gojek/turing/engines/experiment/plugin/manager"
	runnerPlugin "github.com/gojek/turing/engines/experiment/plugin/runner"
	"github.com/hashicorp/go-plugin"
	wrapper "github.com/zaffka/zap-to-hclog"
	"go.uber.org/zap"
)

type ClientServices struct {
	Manager managerPlugin.ConfigurableExperimentManager
	Runner  runnerPlugin.ConfigurableExperimentRunner
}

func Connect(pluginBinary string, logger *zap.Logger) (plugin.ClientProtocol, error) {
	hcLogger := wrapper.Wrap(logger)

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Cmd:              exec.Command(pluginBinary),
		Plugins:          pluginMap,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
		Managed:          true,
		Logger:           hcLogger,
	})

	runtime.SetFinalizer(client, func(c *plugin.Client) {
		c.Kill()
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("error attempting to connect to plugin rpc client: %w", err)
	}

	return rpcClient, nil
}

func Serve(services *ClientServices) {
	plugins := plugin.PluginSet{
		ManagerPluginIdentified: &managerPlugin.ExperimentManagerPlugin{
			Impl: services.Manager,
		},
		RunnerPluginIdentified: &runnerPlugin.ExperimentRunnerPlugin{
			Impl: services.Runner,
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         plugins,
	})
}
