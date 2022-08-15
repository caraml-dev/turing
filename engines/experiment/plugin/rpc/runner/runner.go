package runner

import (
	"encoding/json"
	"net/http"

	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/shared"
	"github.com/caraml-dev/turing/engines/experiment/runner"
)

// ConfigurableExperimentRunner interface of an ExperimentRunner, that can be configured
// with an arbitrary configuration passed as a JSON data
type ConfigurableExperimentRunner interface {
	shared.Configurable
	runner.ExperimentRunner
}

func NewConfigurableExperimentRunner(
	factory func(json.RawMessage) (runner.ExperimentRunner, error),
) ConfigurableExperimentRunner {
	return &configurableExperimentRunner{
		factory: factory,
	}
}

type configurableExperimentRunner struct {
	runner.ExperimentRunner
	factory func(cfg json.RawMessage) (runner.ExperimentRunner, error)
}

func (er *configurableExperimentRunner) Configure(cfg json.RawMessage) (err error) {
	er.ExperimentRunner, err = er.factory(cfg)
	return
}

// GetTreatmentRequest is a struct, used to pass the data required by
// ExperimentRunner.GetTreatmentForRequest() between RPC client and server
type GetTreatmentRequest struct {
	Header  http.Header
	Payload []byte
	Options runner.GetTreatmentOptions
}
