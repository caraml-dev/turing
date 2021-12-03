package plugin

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"runtime"

	"github.com/gojek/turing/engines/experiment/manager"
	managerPlugin "github.com/gojek/turing/engines/experiment/plugin/manager"
	"github.com/hashicorp/go-plugin"
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

var PluginMap = map[string]plugin.Plugin{
	ManagerPluginIdentifier: &managerPlugin.ExperimentManagerPlugin{},
}

type Services struct {
	Manager managerPlugin.ConfigurableExperimentManager
}

type Configuration struct {
	PluginBinary   string
	PluginConfig   json.RawMessage
	PluginLogLevel string
}

func (c *Configuration) Build() (*Services, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Cmd:              exec.Command(c.PluginBinary),
		Plugins:          PluginMap,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
		Managed:          true,
		Logger:           nil,
	})

	runtime.SetFinalizer(client, func(c *plugin.Client) {
		c.Kill()
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("error attempting to connect to plugin rpc client: %w", err)
	}

	raw, err := rpcClient.Dispense(ManagerPluginIdentifier)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve experiment manager plugin instance: %w", err)
	}

	experimentManager, ok := raw.(managerPlugin.ConfigurableExperimentManager)
	if !ok {
		return nil, fmt.Errorf("unable to cast %T to %s for plugin \"%s\"",
			raw,
			reflect.TypeOf((*manager.ExperimentManager)(nil)).Elem(),
			ManagerPluginIdentifier)
	}

	err = experimentManager.Configure(c.PluginConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure experiment manager plugin instance: %w", err)
	}

	return &Services{Manager: experimentManager}, nil
}
