package nop

import (
	"encoding/json"
	"net/http"

	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
)

type ExperimentRunner struct{}

func (r *ExperimentRunner) Configure(json.RawMessage) error {
	return nil
}

func (r *ExperimentRunner) GetTreatmentForRequest(
	http.Header,
	[]byte,
	runner.GetTreatmentOptions,
) (*runner.Treatment, error) {
	return &runner.Treatment{
		Name:   "my treatment",
		Config: json.RawMessage(`{"config-1": "value-a"}`),
	}, nil
}

func (r *ExperimentRunner) RegisterMetricsCollector(_ metrics.Collector, _ runner.MetricsRegistrationHelper) error {
	return nil
}
