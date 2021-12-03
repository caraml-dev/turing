package manager

import (
	"encoding/json"
	"github.com/gojek/turing/engines/experiment/manager"
	"net/rpc"
)

type rpcClient struct {
	*rpc.Client
}

func (c *rpcClient) Configure(cfg json.RawMessage) error {
	return c.Call("Plugin.Configure", cfg, new(interface{}))
}

func (c *rpcClient) GetEngineInfo() manager.Engine {
	info := manager.Engine{}
	err := c.Call("Plugin.GetEngineInfo", new(interface{}), &info)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return info
}
