package plugin

import (
	"github.com/hashicorp/go-plugin"
	"net/rpc"
)

type ExperimentEnginePlugin struct {
	ManagerImpl ConfigurableExperimentManager
	RunnerImpl  ConfigurableExperimentRunner
}

func (p *ExperimentEnginePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &rpcServer{
		ManagerImpl: p.ManagerImpl,
		RunnerImpl:  p.RunnerImpl,
	}, nil
}

func (ExperimentEnginePlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &rpcClient{Client: c}, nil
}
