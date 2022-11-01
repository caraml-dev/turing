package rpc

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/hashicorp/go-plugin"
	"github.com/mitchellh/hashstructure/v2"
	"go.uber.org/zap"

	"github.com/caraml-dev/turing/engines/experiment/config"
	"github.com/caraml-dev/turing/engines/experiment/manager"
	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/shared"
	"github.com/caraml-dev/turing/engines/experiment/runner"
)

var factoriesmu sync.Mutex
var factories = make(map[string]*EngineFactory)

// EngineFactory implements experiment.EngineFactory and creates experiment manager/runner
// backed by net/rpc plugin implementations
type EngineFactory struct {
	sync.Mutex
	manager manager.ExperimentManager
	runner  runner.ExperimentRunner

	Client       plugin.ClientProtocol
	EngineConfig json.RawMessage
}

func (f *EngineFactory) dispenseAndConfigure(id string) (interface{}, error) {
	raw, err := f.Client.Dispense(id)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve \"%s\" plugin instance: %w", id, err)
	}

	configurable, ok := raw.(shared.Configurable)
	if !ok {
		return nil, fmt.Errorf(
			"unable to cast %T to %s for plugin \"%s\"", raw,
			reflect.TypeOf((*shared.Configurable)(nil)).Elem(), id)
	}

	err = configurable.Configure(f.EngineConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure \"%s\" plugin instance: %w", id, err)
	}
	return configurable, nil
}

func (f *EngineFactory) GetExperimentManager() (manager.ExperimentManager, error) {
	f.Lock()
	defer f.Unlock()

	if f.manager == nil {
		instance, err := f.dispenseAndConfigure(ManagerPluginIdentifier)
		if err != nil {
			return nil, err
		}
		f.manager = instance.(manager.ExperimentManager)
	}

	return f.manager, nil
}

func (f *EngineFactory) GetExperimentRunner() (runner.ExperimentRunner, error) {
	f.Lock()
	defer f.Unlock()

	if f.runner == nil {
		instance, err := f.dispenseAndConfigure(RunnerPluginIdentifier)
		if err != nil {
			return nil, err
		}
		f.runner = instance.(runner.ExperimentRunner)
	}

	return f.runner, nil
}

func NewFactory(name string, cfg config.EngineConfig, logger *zap.SugaredLogger) (*EngineFactory, error) {
	factoriesmu.Lock()
	defer factoriesmu.Unlock()

	// get a hash of the engine's configuration and use it as a configuration's fingerprint
	cfgHash, err := hashstructure.Hash(cfg, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, err
	}
	factoryKey := fmt.Sprintf("%s-%d", name, cfgHash)

	if engineFactory, ok := factories[factoryKey]; ok {
		return engineFactory, nil
	}

	engineCfg, err := cfg.RawEngineConfig()
	if err != nil {
		return nil, err
	}
	if cfg.PluginBinary != "" {
		factories[factoryKey], err = NewFactoryFromBinary(cfg.PluginBinary, engineCfg, logger)
	} else {
		err = fmt.Errorf("`plugin_binary` must be specified")
	}

	if err != nil {
		return nil, err
	}
	return factories[factoryKey], nil
}

func NewFactoryFromBinary(
	pluginBinary string,
	engineCfg json.RawMessage,
	logger *zap.SugaredLogger,
) (*EngineFactory, error) {
	rpcClient, err := Connect(pluginBinary, logger.Desugar())
	if err != nil {
		return nil, err
	}

	return &EngineFactory{
		Client:       rpcClient,
		EngineConfig: engineCfg,
	}, nil
}
