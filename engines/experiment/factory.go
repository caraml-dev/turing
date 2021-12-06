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

type Factory interface {
	GetExperimentManager() (manager.ExperimentManager, error)
	GetExperimentRunner() (runner.ExperimentRunner, error)
}

func NewFactory(name string, cfg interface{}, logger *zap.Logger) (Factory, error) {
	var engineCfg EngineConfig
	if err := mapstructure.Decode(cfg, &engineCfg); err != nil {
		return nil, err
	}

	engineCfgJSON, err := json.Marshal(engineCfg.EngineConfiguration)
	if err != nil {
		return nil, err
	}

	if engineCfg.Plugin.PluginBinary != "" {
		pluginCfg := engineCfg.Plugin
		pluginCfg.PluginConfig = engineCfgJSON
		return plugin.NewFactory(&pluginCfg, logger)
	}

	return v1.NewFactory(name, engineCfgJSON)
}
