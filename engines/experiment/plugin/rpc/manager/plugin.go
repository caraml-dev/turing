package manager

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// ExperimentManagerPlugin implements hashicorp/go-plugin's Plugin interface
// for manager.ExperimentManager
type ExperimentManagerPlugin struct {
	Impl ConfigurableExperimentManager
}

func (p *ExperimentManagerPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &rpcServer{Impl: p.Impl}, nil
}

func (ExperimentManagerPlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &rpcClient{RPCClient: c}, nil
}
