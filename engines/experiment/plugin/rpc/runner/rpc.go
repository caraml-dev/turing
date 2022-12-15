package runner

import (
	"encoding/json"
	"net/http"
	"net/rpc"
	"time"

	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/shared"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation"
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
	"github.com/hashicorp/go-plugin"
)

// rpcClient implements ConfigurableExperimentRunner interface
type rpcClient struct {
	*plugin.MuxBroker
	shared.RPCClient
}

func (c *rpcClient) Configure(cfg json.RawMessage) error {
	return c.Call("Plugin.Configure", cfg, new(interface{}))
}

func (c *rpcClient) GetTreatmentForRequest(
	header http.Header,
	payload []byte,
	options runner.GetTreatmentOptions,
) (*runner.Treatment, error) {
	req := GetTreatmentRequest{
		Header:  header,
		Payload: payload,
		Options: options,
	}
	var resp runner.Treatment

	err := c.Call("Plugin.GetTreatmentForRequest", &req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *rpcClient) RegisterMetrics(
	_ metrics.Collector,
	metricsRegistrationHelper runner.MetricsRegistrationHelper,
) error {
	collectorBrokerId := c.MuxBroker.NextId()
	go c.MuxBroker.AcceptAndServe(collectorBrokerId, &rpcCollectorServer{})

	metricsRegistrationHelperBrokerId := c.MuxBroker.NextId()
	go c.MuxBroker.AcceptAndServe(metricsRegistrationHelperBrokerId,
		&rpcMetricsRegistrationHelperServer{Impl: metricsRegistrationHelper})

	return c.Call(
		"Plugin.RegisterMetrics",
		[]uint32{collectorBrokerId, metricsRegistrationHelperBrokerId},
		new(interface{}),
	)
}

// rpcCollectorClient is an implementation of Collector used by the plugin to talk to the router over RPC.
type rpcCollectorClient struct {
	*plugin.MuxBroker
	shared.RPCClient
}

func (c *rpcCollectorClient) MeasureDurationMsSince(key metrics.MetricName, starttime time.Time, labels map[string]string) error {
	req := MeasureDurationMsSinceRequest{
		Key:       key,
		Starttime: starttime,
		Labels:    labels,
	}
	return c.Call("Plugin.MeasureDurationMsSince", &req, new(interface{}))
}

func (c *rpcCollectorClient) MeasureDurationMs(key metrics.MetricName, labels map[string]func() string) func() {
	var resp func()

	req := MeasureDurationMsRequest{
		Key:    key,
		Labels: labels,
	}

	err := c.Call("Plugin.MeasureDurationMs", &req, &resp)
	if err != nil {
		return nil
	}

	return resp
}

func (c *rpcCollectorClient) RecordGauge(key metrics.MetricName, value float64, labels map[string]string) error {
	req := RecordGaugeRequest{
		Key:    key,
		Value:  value,
		Labels: labels,
	}
	return c.Call("Plugin.RecordGauge", &req, new(interface{}))
}

func (c *rpcCollectorClient) Inc(key metrics.MetricName, labels map[string]string) error {
	req := IncRequest{
		Key:    key,
		Labels: labels,
	}
	return c.Call("Plugin.Inc", &req, new(interface{}))
}

// rpcMetricsRegistrationHelperClient is an implementation of MetricsRegistrationHelper used by the plugin to talk to
// the router over RPC.
type rpcMetricsRegistrationHelperClient struct {
	*plugin.MuxBroker
	shared.RPCClient
}

func (c *rpcMetricsRegistrationHelperClient) Register(metrics []instrumentation.Metric) error {
	return c.Call("Plugin.Register", metrics, new(interface{}))
}

// rpcServer serves the implementation of a ConfigurableExperimentRunner
type rpcServer struct {
	*plugin.MuxBroker
	Impl ConfigurableExperimentRunner
}

func (s *rpcServer) Configure(cfg json.RawMessage, _ *interface{}) (err error) {
	return s.Impl.Configure(cfg)
}

func (s *rpcServer) GetTreatmentForRequest(req *GetTreatmentRequest, resp *runner.Treatment) error {
	treatment, err := s.Impl.GetTreatmentForRequest(req.Header, req.Payload, req.Options)
	if err != nil {
		return err
	}
	*resp = *treatment
	return nil
}

func (s *rpcServer) RegisterMetrics(brokerIds []uint32, _ *interface{}) (err error) {
	collectorConn, err := s.MuxBroker.Dial(brokerIds[0])
	if err != nil {
		return err
	}

	metricsRegistrationHelperConn, err := s.MuxBroker.Dial(brokerIds[1])
	if err != nil {
		return err
	}

	return s.Impl.RegisterMetrics(
		&rpcCollectorClient{RPCClient: rpc.NewClient(collectorConn)},
		&rpcMetricsRegistrationHelperClient{RPCClient: rpc.NewClient(metricsRegistrationHelperConn)},
	)
}

// rpcCollectorServer is used by the router to talk to the plugin over RPC.
type rpcCollectorServer struct{}

func (s *rpcCollectorServer) MeasureDurationMsSince(req *MeasureDurationMsSinceRequest, _ *interface{}) error {
	return metrics.Glob().MeasureDurationMsSince(req.Key, req.Starttime, req.Labels)
}

func (s *rpcCollectorServer) MeasureDurationMs(req *MeasureDurationMsRequest, resp *func()) error {
	returnFunc := metrics.Glob().MeasureDurationMs(req.Key, req.Labels)

	*resp = returnFunc
	return nil
}

func (s *rpcCollectorServer) RecordGauge(req *RecordGaugeRequest, _ *interface{}) error {
	return metrics.Glob().RecordGauge(req.Key, req.Value, req.Labels)
}

func (s *rpcCollectorServer) Inc(req *IncRequest, _ *interface{}) error {
	return metrics.Glob().Inc(req.Key, req.Labels)
}

// rpcMetricsRegistrationHelperServer is used by the router to talk to the plugin over RPC.
type rpcMetricsRegistrationHelperServer struct {
	Impl runner.MetricsRegistrationHelper
}

func (s *rpcMetricsRegistrationHelperServer) Register(req []instrumentation.Metric, _ *interface{}) error {
	return s.Impl.Register(req)
}
