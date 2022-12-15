package nop

import (
	"encoding/json"
	"net/http"

	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
)

var nopTreatment = &runner.Treatment{}

// ExperimentRunner is a dummy experiment runner
type ExperimentRunner struct{}

// GetTreatmentForRequest returns a dummy experiment treatment
func (ExperimentRunner) GetTreatmentForRequest(
	http.Header,
	[]byte,
	runner.GetTreatmentOptions,
) (*runner.Treatment, error) {
	return nopTreatment, nil
}

func (ExperimentRunner) RegisterMetrics(
	_ metrics.Collector,
	_ runner.MetricsRegistrationHelper,
) error {
	return nil
}

// NewExperimentRunner is a creator for the experiment runners
func NewExperimentRunner(json.RawMessage) (runner.ExperimentRunner, error) {
	return &ExperimentRunner{}, nil
}
