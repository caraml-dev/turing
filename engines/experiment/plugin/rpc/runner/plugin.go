package runner

import (
	"net/rpc"

	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
	"github.com/hashicorp/go-plugin"
)

// ExperimentRunnerPlugin implements hashicorp/go-plugin's Plugin interface
// for runner.ExperimentRunner
type ExperimentRunnerPlugin struct {
	Impl ConfigurableExperimentRunner
}

func (p *ExperimentRunnerPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &rpcServer{Impl: p.Impl, MuxBroker: b}, nil
}

func (ExperimentRunnerPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &rpcClient{RPCClient: c, MuxBroker: b}, nil
}

// CollectorPlugin implements hashicorp/go-plugin's Plugin interface
// for metrics.Collector
type CollectorPlugin struct {
	Impl metrics.Collector
}

func (p *CollectorPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &rpcCollectorServer{}, nil
}

func (CollectorPlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &rpcCollectorClient{RPCClient: c}, nil
}
