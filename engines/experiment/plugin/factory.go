package plugin

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/plugin/shared"
	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/hashicorp/go-plugin"
	"go.uber.org/zap"
)

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
		instance, err := f.dispenseAndConfigure(ManagerPluginIdentified)
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
		instance, err := f.dispenseAndConfigure(RunnerPluginIdentified)
		if err != nil {
			return nil, err
		}
		f.runner = instance.(runner.ExperimentRunner)
	}

	return f.runner, nil
}

func NewFactory(pluginBinary string, engineCfg json.RawMessage, logger *zap.SugaredLogger) (*EngineFactory, error) {
	rpcClient, err := Connect(pluginBinary, logger.Desugar())
	if err != nil {
		return nil, err
	}

	return &EngineFactory{
		Client:       rpcClient,
		EngineConfig: engineCfg,
	}, nil
}
