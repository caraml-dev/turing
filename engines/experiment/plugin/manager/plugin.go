package manager

import (
	"github.com/hashicorp/go-plugin"
	"net/rpc"
)

type ExperimentManagerPlugin struct {
	Impl ConfigurableExperimentManager
}

func (p *ExperimentManagerPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &rpcServer{
		Impl: p.Impl,
	}, nil
}

func (ExperimentManagerPlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &rpcClient{Client: c}, nil
}
