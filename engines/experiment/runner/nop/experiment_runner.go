package nop

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gojek/turing/engines/experiment/runner"
)

// init ensures this runner is registered when the package is imported.
func init() {
	err := runner.Register("nop", NewExperimentRunner)
	if err != nil {
		log.Fatal(err)
	}
}

// treatment captures a dummy experiment treatment created by the ExperimentRunner
// and implements the runner.ExperimentRunner
type treatment struct{}

// GetExperimentName returns an empty string
func (treatment) GetExperimentName() string {
	return ""
}

// GetTreatmentName returns an empty string
func (treatment) GetName() string {
	return ""
}

// GetConfig returns an empty json object
func (treatment) GetConfig() json.RawMessage {
	return nil
}

// ExperimentRunner is a dummy experiment runner
type ExperimentRunner struct{}

// GetTreatmentForRequest returns a dummy experiment treatment
func (ExperimentRunner) GetTreatmentForRequest(
	context.Context,
	runner.Logger,
	http.Header,
	[]byte,
) (runner.Treatment, error) {
	return treatment{}, nil
}

// NewExperimentRunner is a creator for the experiment runners
func NewExperimentRunner(json.RawMessage) (runner.ExperimentRunner, error) {
	return &ExperimentRunner{}, nil
}
