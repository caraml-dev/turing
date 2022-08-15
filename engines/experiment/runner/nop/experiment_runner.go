package nop

import (
	"encoding/json"
	"net/http"

	"github.com/caraml-dev/turing/engines/experiment/runner"
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

// NewExperimentRunner is a creator for the experiment runners
func NewExperimentRunner(json.RawMessage) (runner.ExperimentRunner, error) {
	return &ExperimentRunner{}, nil
}
