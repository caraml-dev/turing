package manager

import (
	"encoding/json"
	"fmt"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/plugin/shared"
)

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
		println(fmt.Sprintf("plugin errors: %v", err))
	}

	return resp
}

type rpcServer struct {
	Impl ConfigurableExperimentManager
}

func (s *rpcServer) Configure(cfg json.RawMessage, _ *interface{}) (err error) {
	return s.Impl.Configure(cfg)
}

func (s *rpcServer) GetEngineInfo(_ interface{}, resp *manager.Engine) error {
	*resp = s.Impl.GetEngineInfo()
	return nil
}
