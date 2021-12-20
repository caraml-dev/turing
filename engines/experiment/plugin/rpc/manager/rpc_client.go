package manager

import (
	"encoding/json"
	"fmt"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/plugin/rpc/shared"
)

// rpcClient implements ConfigurableExperimentManager interface
type rpcClient struct {
	shared.RPCClient
}

func (c *rpcClient) Configure(cfg json.RawMessage) error {
	return c.Call("Plugin.Configure", cfg, new(interface{}))
}

func (c *rpcClient) GetEngineInfo() manager.Engine {
	resp := manager.Engine{}
	err := c.Call("Plugin.GetEngineInfo", new(interface{}), &resp)
	if err != nil {
		// err should be propagated upstream, but it's currently not
		// possible as GetEngineInfo() on the manager.ExperimentManager
		// interface doesn't return errors
		fmt.Printf("plugin errors: %v", err)
	}

	return resp
}

func (c *rpcClient) ValidateExperimentConfig(cfg json.RawMessage) error {
	return c.Call("Plugin.ValidateExperimentConfig", cfg, new(interface{}))
}

func (c *rpcClient) GetExperimentRunnerConfig(cfg json.RawMessage) (json.RawMessage, error) {
	var resp json.RawMessage
	err := c.Call("Plugin.GetExperimentRunnerConfig", cfg, &resp)
	return resp, err
}
