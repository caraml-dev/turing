package nop

import (
	"context"
	"encoding/json"
	"github.com/gojek/turing/engines/experiment/runner"
	"net/http"
)

type ExperimentRunner struct{}

func (r *ExperimentRunner) Configure(json.RawMessage) error {
	return nil
}

func (r *ExperimentRunner) GetTreatmentForRequest(
	context.Context,
	runner.Logger,
	http.Header,
	[]byte,
) (*runner.Treatment, error) {
	return &runner.Treatment{Name: "my treatment"}, nil
}
