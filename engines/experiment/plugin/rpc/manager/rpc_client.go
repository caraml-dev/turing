package manager

import (
	"encoding/json"

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

func (c *rpcClient) GetEngineInfo() (resp manager.Engine, err error) {
	err = c.Call("Plugin.GetEngineInfo", new(interface{}), &resp)
	return
}

func (c *rpcClient) ValidateExperimentConfig(cfg json.RawMessage) error {
	return c.Call("Plugin.ValidateExperimentConfig", cfg, new(interface{}))
}

func (c *rpcClient) GetExperimentRunnerConfig(cfg json.RawMessage) (resp json.RawMessage, err error) {
	err = c.Call("Plugin.GetExperimentRunnerConfig", cfg, &resp)
	return
}

func (c *rpcClient) IsCacheEnabled() (resp bool, err error) {
	err = c.Call("Plugin.IsCacheEnabled", new(interface{}), &resp)
	return
}

func (c *rpcClient) ListClients() (resp []manager.Client, err error) {
	resp = []manager.Client{}
	err = c.Call("Plugin.ListClients", new(interface{}), &resp)
	return
}

func (c *rpcClient) ListExperiments() (resp []manager.Experiment, err error) {
	resp = []manager.Experiment{}
	err = c.Call("Plugin.ListExperiments", new(interface{}), &resp)
	return
}

func (c *rpcClient) ListExperimentsForClient(client manager.Client) (resp []manager.Experiment, err error) {
	resp = []manager.Experiment{}
	err = c.Call("Plugin.ListExperimentsForClient", client, &resp)
	return
}

func (c *rpcClient) ListVariablesForClient(client manager.Client) (resp []manager.Variable, err error) {
	resp = []manager.Variable{}
	err = c.Call("Plugin.ListVariablesForClient", client, &resp)
	return
}

func (c *rpcClient) ListVariablesForExperiments(
	experiments []manager.Experiment,
) (resp map[string][]manager.Variable, err error) {
	err = c.Call("Plugin.ListVariablesForExperiments", experiments, &resp)
	return
}
