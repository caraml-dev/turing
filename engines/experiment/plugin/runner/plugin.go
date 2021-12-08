package runner

import (
	"github.com/hashicorp/go-plugin"
	"net/rpc"
)

type ExperimentRunnerPlugin struct {
	Impl ConfigurableExperimentRunner
}

func (p *ExperimentRunnerPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &rpcServer{
		Impl: p.Impl,
	}, nil
}

func (ExperimentRunnerPlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &rpcClient{Client: c}, nil
}
