package experiment

import (
	"github.com/gojek/turing/engines/experiment/config"
	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/plugin/inproc"
	"github.com/gojek/turing/engines/experiment/plugin/rpc"
	"github.com/gojek/turing/engines/experiment/runner"
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
// 	- experiment/plugin/inproc/factory (for experiment engines implemented as compile-time plugins)
//  - experiment/plugin/rpc/factory (for experiment engines implemented as external net/rpc plugins)
// The actual implementation is determined based on provided engine configuration (passed via `cfg`)
func NewEngineFactory(name string, cfg map[string]interface{}, logger *zap.SugaredLogger) (EngineFactory, error) {
	var engineCfg config.EngineConfig
	if err := mapstructure.Decode(cfg, &engineCfg); err != nil {
		return nil, err
	}

	// plugin-based implementation of the experiment engine factory
	if engineCfg.IsPlugin() {
		return rpc.NewFactory(name, engineCfg, logger)
	}

	// compile-time implementation of the experiment engine factory
	return inproc.NewEngineFactory(name, engineCfg)
}
