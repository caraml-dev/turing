package v1

import (
	"encoding/json"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/runner"
	managerV1 "github.com/gojek/turing/engines/experiment/v1/manager"
	runnerV1 "github.com/gojek/turing/engines/experiment/v1/runner"
)

type EngineFactory struct {
	EngineName   string
	EngineConfig json.RawMessage
}

func (f *EngineFactory) GetExperimentManager() (manager.ExperimentManager, error) {
	return managerV1.Get(f.EngineName, f.EngineConfig)
}

func (f *EngineFactory) GetExperimentRunner() (runner.ExperimentRunner, error) {
	return runnerV1.Get(f.EngineName, f.EngineConfig)
}

func NewEngineFactory(name string, cfg json.RawMessage) (*EngineFactory, error) {
	return &EngineFactory{
		EngineName:   name,
		EngineConfig: cfg,
	}, nil
}
