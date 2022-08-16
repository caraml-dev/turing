package inproc

import (
	"encoding/json"

	"github.com/caraml-dev/turing/engines/experiment/config"

	"github.com/caraml-dev/turing/engines/experiment/manager"
	managerPlugin "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/manager"
	runnerPlugin "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner"
	"github.com/caraml-dev/turing/engines/experiment/runner"
)

// EngineFactory implements experiment.EngineFactory and creates experiment manager/runner
// backed by compile-time plugins and registered within the application
type EngineFactory struct {
	EngineName   string
	EngineConfig json.RawMessage
}

func (f *EngineFactory) GetExperimentManager() (manager.ExperimentManager, error) {
	return managerPlugin.Get(f.EngineName, f.EngineConfig)
}

func (f *EngineFactory) GetExperimentRunner() (runner.ExperimentRunner, error) {
	return runnerPlugin.Get(f.EngineName, f.EngineConfig)
}

func NewEngineFactory(name string, cfg config.EngineConfig) (*EngineFactory, error) {
	engineCfg, err := cfg.RawEngineConfig()
	if err != nil {
		return nil, err
	}
	return &EngineFactory{
		EngineName:   name,
		EngineConfig: engineCfg,
	}, nil
}
