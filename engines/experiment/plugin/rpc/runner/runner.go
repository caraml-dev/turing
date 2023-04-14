package runner

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/metrics"

	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/shared"
	"github.com/caraml-dev/turing/engines/experiment/runner"
)

// ConfigurableExperimentRunner interface of an ExperimentRunner, that can be configured
// with an arbitrary configuration passed as a JSON data
type ConfigurableExperimentRunner interface {
	shared.Configurable
	runner.ExperimentRunner
}

func NewConfigurableExperimentRunner(
	factory func(json.RawMessage) (runner.ExperimentRunner, error),
) ConfigurableExperimentRunner {
	return &configurableExperimentRunner{
		factory: factory,
	}
}

type configurableExperimentRunner struct {
	runner.ExperimentRunner
	factory func(cfg json.RawMessage) (runner.ExperimentRunner, error)
}

func (er *configurableExperimentRunner) Configure(cfg json.RawMessage) (err error) {
	er.ExperimentRunner, err = er.factory(cfg)
	return
}

// GetTreatmentRequest is a struct, used to pass the data required by
// ExperimentRunner.GetTreatmentForRequest() between RPC client and server
type GetTreatmentRequest struct {
	Header  http.Header
	Payload []byte
	Options runner.GetTreatmentOptions
}

// MeasureDurationMsSinceRequest is a struct, used to pass the data required by
// Collector.MeasureDurationMsSince() between RPC client and server
type MeasureDurationMsSinceRequest struct {
	Key       metrics.MetricName
	Starttime time.Time
	Labels    map[string]string
}

// MeasureDurationMsRequest is a struct, used to pass the data required by
// Collector.MeasureDurationMs() between RPC client and server
type MeasureDurationMsRequest struct {
	Key    metrics.MetricName
	Labels map[string]func() string
}

// RecordGaugeRequest is a struct, used to pass the data required by
// Collector.RecordGauge() between RPC client and server
type RecordGaugeRequest struct {
	Key    metrics.MetricName
	Value  float64
	Labels map[string]string
}

// IncRequest is a struct, used to pass the data required by
// Collector.Inc() between RPC client and server
type IncRequest struct {
	Key    metrics.MetricName
	Labels map[string]string
}
