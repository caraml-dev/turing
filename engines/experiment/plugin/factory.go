package plugin

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/gojek/turing/engines/experiment/manager"
	managerPlugin "github.com/gojek/turing/engines/experiment/plugin/manager"
	runnerPlugin "github.com/gojek/turing/engines/experiment/plugin/runner"
	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/hashicorp/go-plugin"
	"go.uber.org/zap"
)

type factory struct {
	sync.Mutex
	rpcClient plugin.ClientProtocol
	config    json.RawMessage
	manager   manager.ExperimentManager
	runner    runner.ExperimentRunner
}

func (f *factory) GetExperimentManager() (manager.ExperimentManager, error) {
	f.Lock()
	defer f.Unlock()

	if f.manager == nil {
		raw, err := f.rpcClient.Dispense(ManagerPluginIdentified)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve experiment manager plugin instance: %w", err)
		}

		experimentManager, ok := raw.(managerPlugin.ConfigurableExperimentManager)
		if !ok {
			return nil, fmt.Errorf("unable to cast %T to %s for plugin \"%s\"",
				raw,
				reflect.TypeOf((*manager.ExperimentManager)(nil)).Elem(),
				ManagerPluginIdentified)
		}

		err = experimentManager.Configure(f.config)
		if err != nil {
			return nil, fmt.Errorf("failed to configure experiment manager plugin instance: %w", err)
		}

		f.manager = experimentManager
	}

	return f.manager, nil
}

func (f *factory) GetExperimentRunner() (runner.ExperimentRunner, error) {
	f.Lock()
	defer f.Unlock()

	if f.runner == nil {
		raw, err := f.rpcClient.Dispense(RunnerPluginIdentified)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve experiment runner plugin instance: %w", err)
		}

		experimentRunner, ok := raw.(runnerPlugin.ConfigurableExperimentRunner)
		if !ok {
			return nil, fmt.Errorf("unable to cast %T to %s for plugin \"%s\"",
				raw,
				reflect.TypeOf((*runner.ExperimentRunner)(nil)).Elem(),
				RunnerPluginIdentified)
		}

		err = experimentRunner.Configure(f.config)
		if err != nil {
			return nil, fmt.Errorf("failed to configure experiment runner plugin instance: %w", err)
		}

		f.runner = experimentRunner
	}

	return f.runner, nil
}

func NewFactory(cfg *Configuration, logger *zap.SugaredLogger) (*factory, error) {
	rpcClient, err := Connect(cfg, logger.Desugar())
	if err != nil {
		return nil, err
	}

	return &factory{
		rpcClient: rpcClient,
		config:    cfg.PluginConfig,
	}, nil
}
