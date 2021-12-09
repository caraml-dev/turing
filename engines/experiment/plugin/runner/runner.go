package runner

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gojek/turing/engines/experiment/runner"
)

type ConfigurableExperimentRunner interface {
	Configure(cfg json.RawMessage) error
	runner.ExperimentRunner
}

type GetTreatmentRequest struct {
	Context context.Context
	Logger  runner.Logger
	Header  http.Header
	Payload []byte
}
