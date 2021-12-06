package v1

import (
	"encoding/json"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/runner"
	managerV1 "github.com/gojek/turing/engines/experiment/v1/manager"
	runnerV1 "github.com/gojek/turing/engines/experiment/v1/runner"
)

type factory struct {
	engineName string
	cfg        json.RawMessage
}

func (f *factory) GetExperimentManager() (manager.ExperimentManager, error) {
	return managerV1.Get(f.engineName, f.cfg)
}

func (f *factory) GetExperimentRunner() (runner.ExperimentRunner, error) {
	return runnerV1.Get(f.engineName, f.cfg)
}

func NewFactory(name string, cfg json.RawMessage) (*factory, error) {
	return &factory{
		engineName: name,
		cfg:        cfg,
	}, nil
}
