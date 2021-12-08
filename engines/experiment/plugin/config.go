package plugin

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"runtime"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/hashicorp/go-plugin"
	"github.com/zaffka/zap-to-hclog"
	"go.uber.org/zap"
)

const (
	ExperimentsEnginePluginIdentifier = "experiment_engine"
)

var (
	HandshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "EXPERIMENTS_PLUGIN",
		MagicCookieValue: "turing",
	}
)

var pluginMap = map[string]plugin.Plugin{
	ExperimentsEnginePluginIdentifier: &ExperimentEnginePlugin{},
}

type Services struct {
	Manager ConfigurableExperimentManager
	Runner  ConfigurableExperimentRunner
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

	raw, err := rpcClient.Dispense(ExperimentsEnginePluginIdentifier)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve plugin capabilities: %w", err)
	}

	pluginCapabilities, ok := raw.(PluginCapabilities)
	if !ok {
		return nil, fmt.Errorf("unable to cast %T to plugin.PluginCapabilities", raw)
	}

	capabilities := pluginCapabilities.Capabilities()

	var (
		experimentManager ConfigurableExperimentManager
		experimentRunner  ConfigurableExperimentRunner
	)

	if capabilities.Manager {
		experimentManager, ok = raw.(ConfigurableExperimentManager)
		if !ok {
			return nil, fmt.Errorf("unable to cast %T to %s for plugin \"%s\"",
				raw,
				reflect.TypeOf((*manager.ExperimentManager)(nil)).Elem(),
				ExperimentsEnginePluginIdentifier)
		}

		err = experimentManager.Configure(c.PluginConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to configure experiment manager plugin instance: %w", err)
		}
	}

	if capabilities.Runner {
		experimentRunner, ok = raw.(ConfigurableExperimentRunner)
		if !ok {
			return nil, fmt.Errorf("unable to cast %T to %s for plugin \"%s\"",
				raw,
				reflect.TypeOf((*runner.ExperimentRunner)(nil)).Elem(),
				ExperimentsEnginePluginIdentifier)
		}

		err = experimentRunner.Configure(c.PluginConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to configure experiment runner plugin instance: %w", err)
		}
	}

	return &Services{
		Manager: experimentManager,
		Runner:  experimentRunner,
	}, nil
}
