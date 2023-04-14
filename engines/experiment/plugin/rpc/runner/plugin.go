package runner

import (
	"net/rpc"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/metrics"
	"github.com/hashicorp/go-plugin"

	"github.com/caraml-dev/turing/engines/experiment/runner"
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

// MetricsRegistrationHelperPlugin implements hashicorp/go-plugin's Plugin interface
// for runner.MetricsRegistrationHelper
type MetricsRegistrationHelperPlugin struct {
	Impl runner.MetricsRegistrationHelper
}

func (p *MetricsRegistrationHelperPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &rpcMetricsRegistrationHelperServer{}, nil
}

func (MetricsRegistrationHelperPlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &rpcMetricsRegistrationHelperClient{RPCClient: c}, nil
}
