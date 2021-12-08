package plugin

import (
	"context"
	"github.com/gojek/turing/engines/experiment/runner"
	"net/http"
)

type ConfigurableExperimentRunner interface {
	Configurable
	runner.ExperimentRunner
}

type GetTreatmentRequest struct {
	Context context.Context
	Logger  runner.Logger
	Header  http.Header
	Payload []byte
}
