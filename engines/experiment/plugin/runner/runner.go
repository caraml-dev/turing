package runner

import (
	"context"
	"net/http"

	"github.com/gojek/turing/engines/experiment/plugin/shared"

	"github.com/gojek/turing/engines/experiment/runner"
)

// ConfigurableExperimentRunner interface of an ExperimentRunner, that can be configured
// with an arbitrary configuration passed as a JSON data
type ConfigurableExperimentRunner interface {
	shared.Configurable
	runner.ExperimentRunner
}

// GetTreatmentRequest is a struct, used to pass the data required by
// ExperimentRunner.GetTreatmentForRequest() between RPC client and server
type GetTreatmentRequest struct {
	Context context.Context
	Logger  runner.Logger
	Header  http.Header
	Payload []byte
}
