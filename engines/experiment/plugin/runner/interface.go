package runner

import (
	"context"
	"github.com/gojek/turing/engines/experiment/pkg/types"
	"github.com/gojek/turing/engines/experiment/runner"
	"net/http"
)

type ConfigurableExperimentRunner interface {
	types.Configurable
	runner.ExperimentRunner
}

type getTreatmentRequest struct {
	ctx     context.Context
	log     runner.Logger
	header  http.Header
	payload []byte
}
