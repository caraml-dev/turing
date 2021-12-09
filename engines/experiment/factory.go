package experiment

import (
	"encoding/json"
	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/plugin"
	"github.com/gojek/turing/engines/experiment/runner"
	v1 "github.com/gojek/turing/engines/experiment/v1"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

// EngineFactory interface defines methods for accessing manager/runner
// of a given experiment engine
type EngineFactory interface {
	GetExperimentManager() (manager.ExperimentManager, error)
	GetExperimentRunner() (runner.ExperimentRunner, error)
}

// NewEngineFactory is a constructor method that creates a new instance of EngineFactory
// The concrete implementation of EngineFactory can be either:
// 	- experiment/v1/factory (for experiment engines implemented as compile-time plugins)
//  - experiment/plugin/factory (for experiment engines implemented as external net/rpc plugins)
// The actual implementation is determined based on provided engine configuration (passed via `cfg`)
func NewEngineFactory(name string, cfg map[string]interface{}, logger *zap.SugaredLogger) (EngineFactory, error) {
	var engineCfg EngineConfig
	if err := mapstructure.Decode(cfg, &engineCfg); err != nil {
		return nil, err
	}

	engineCfgJSON, err := json.Marshal(engineCfg.EngineConfiguration)
	if err != nil {
		return nil, err
	}

	// plugin-based implementation of the experiment engine factory
	if engineCfg.PluginBinary != "" {
		return plugin.NewFactory(engineCfg.PluginBinary, engineCfgJSON, logger)
	}

	// compile-time implementation of the experiment engine factory
	return v1.NewFactory(name, engineCfgJSON)
}
