package manager

import (
	"encoding/json"
	"net/rpc"

	"github.com/gojek/turing/engines/experiment/manager"
)

type rpcClient struct {
	*rpc.Client
}

func (c *rpcClient) Configure(cfg json.RawMessage) error {
	return c.Call("Plugin.Configure", cfg, new(interface{}))
}

func (c *rpcClient) GetEngineInfo() manager.Engine {
	resp := manager.Engine{}
	err := c.Call("Plugin.GetEngineInfo", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
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
