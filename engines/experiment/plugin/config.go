package plugin

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"runtime"

	"github.com/gojek/turing/engines/experiment/manager"
	managerPlugin "github.com/gojek/turing/engines/experiment/plugin/manager"
	runnerPlugin "github.com/gojek/turing/engines/experiment/plugin/runner"
	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/hashicorp/go-plugin"
	"github.com/zaffka/zap-to-hclog"
	"go.uber.org/zap"
)

const (
	ManagerPluginIdentifier = "experiments_manager"
	RunnerPluginIdentifier  = "experiments_runner"
)

var (
	HandshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "EXPERIMENTS_PLUGIN",
		MagicCookieValue: "turing",
	}
)

var pluginMap = map[string]plugin.Plugin{
	ManagerPluginIdentifier: &managerPlugin.ExperimentManagerPlugin{},
	RunnerPluginIdentifier:  &runnerPlugin.ExperimentRunnerPlugin{},
}

type Services struct {
	Manager managerPlugin.ConfigurableExperimentManager
	Runner  runnerPlugin.ConfigurableExperimentRunner
}

type Configuration struct {
	PluginBinary string
	PluginConfig json.RawMessage
}

func (c *Configuration) Build(logger *zap.Logger) (*Services, error) {
	hcLogger := wrapper.Wrap(logger)

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Cmd:              exec.Command(c.PluginBinary),
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

	rawManager, err := rpcClient.Dispense(ManagerPluginIdentifier)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve experiment manager plugin instance: %w", err)
	}

	experimentManager, ok := rawManager.(managerPlugin.ConfigurableExperimentManager)
	if !ok {
		return nil, fmt.Errorf("unable to cast %T to %s for plugin \"%s\"",
			rawManager,
			reflect.TypeOf((*manager.ExperimentManager)(nil)).Elem(),
			ManagerPluginIdentifier)
	}

	err = experimentManager.Configure(c.PluginConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure experiment manager plugin instance: %w", err)
	}

	rawRunner, err := rpcClient.Dispense(RunnerPluginIdentifier)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve experiment manager plugin instance: %w", err)
	}

	experimentRunner, ok := rawRunner.(runnerPlugin.ConfigurableExperimentRunner)
	if !ok {
		return nil, fmt.Errorf("unable to cast %T to %s for plugin \"%s\"",
			rawManager,
			reflect.TypeOf((*runner.ExperimentRunner)(nil)).Elem(),
			RunnerPluginIdentifier)
	}

	err = experimentRunner.Configure(c.PluginConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure experiment runner plugin instance: %w", err)
	}

	return &Services{
		Manager: experimentManager,
		Runner:  experimentRunner,
	}, nil
}
