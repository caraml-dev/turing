package nop

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gojek/turing/engines/experiment/runner"
)

var nopTreatment = &runner.Treatment{}

// ExperimentRunner is a dummy experiment runner
type ExperimentRunner struct{}

// GetTreatmentForRequest returns a dummy experiment treatment
func (ExperimentRunner) GetTreatmentForRequest(
	context.Context,
	runner.Logger,
	http.Header,
	[]byte,
) (*runner.Treatment, error) {
	return nopTreatment, nil
}

// NewExperimentRunner is a creator for the experiment runners
func NewExperimentRunner(json.RawMessage) (runner.ExperimentRunner, error) {
	return &ExperimentRunner{}, nil
}
