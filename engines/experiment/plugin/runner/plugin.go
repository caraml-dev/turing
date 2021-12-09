package runner

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
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
