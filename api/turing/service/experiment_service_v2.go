package service

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"reflect"
	"time"

	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/engines/experiment"
	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/hashicorp/go-plugin"
	"github.com/patrickmn/go-cache"
)

type experimentsServiceV2 struct {
	// map of engine name -> Experiment Manager
	experimentManagers map[string]*experimentManagerPlugin
	cache              *cache.Cache
}

type experimentManagerPlugin struct {
	manager.ExperimentManager
}

func (e experimentsServiceV2) IsStandardExperimentManager(engine string) bool {
	panic("implement me")
}

func (e experimentsServiceV2) GetStandardExperimentConfig(config interface{}) (manager.TuringExperimentConfig, error) {
	panic("implement me")
}

func (e experimentsServiceV2) ListEngines() []manager.Engine {
	var engines []manager.Engine

	for _, expManager := range e.experimentManagers {
		info := expManager.GetEngineInfo()
		engines = append(engines, manager.Engine{
			Name:                            info.Name,
			DisplayName:                     info.DisplayName,
			Type:                            manager.ExperimentManagerType(info.Type),
			StandardExperimentManagerConfig: nil,
			CustomExperimentManagerConfig:   nil,
		})
	}
	return engines
}

func (e experimentsServiceV2) ListClients(engine string) ([]manager.Client, error) {
	panic("implement me")
}

func (e experimentsServiceV2) ListExperiments(engine string, clientID string) ([]manager.Experiment, error) {
	panic("implement me")
}

func (e experimentsServiceV2) ListVariables(engine string, clientID string, experimentIDs []string) (manager.Variables, error) {
	panic("implement me")
}

func (e experimentsServiceV2) ValidateExperimentConfig(engine string, cfg interface{}) error {
	panic("implement me")
}

func (e experimentsServiceV2) GetExperimentRunnerConfig(engine string, cfg interface{}) (json.RawMessage, error) {
	panic("implement me")
}

func NewExperimentsServiceV2(cfg config.ExperimentsConfig) (ExperimentsService, error) {
	experimentManagers := make(map[string]*experimentManagerPlugin)

	for name, engineCfg := range cfg.Engines {
		configJSON, err := json.Marshal(engineCfg)
		if err != nil {
			return nil, err
		}

		m, err := loadPlugin(filepath.Join(cfg.PluginsDirectory, name), configJSON)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load %s experiment manager plugin", name)
		}
		experimentManagers[name] = m
	}

	// Initialize the experimentsService with cache
	svc := &experimentsServiceV2{
		experimentManagers: experimentManagers,
		cache:              cache.New(expCacheExpirySeconds*time.Second, expCacheCleanUpSeconds*time.Second),
	}

	return svc, nil
}

func loadPlugin(path string, _ json.RawMessage) (*experimentManagerPlugin, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: experiment.HandshakeConfig,
		Plugins: plugin.PluginSet{
			experiment.ManagerPluginIdentifier: &manager.ExperimentManagerPlugin{},
		},
		Cmd:     exec.Command(path),
		Managed: true,
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(experiment.ManagerPluginIdentifier)
	if err != nil {
		return nil, err
	}

	experimentManager, ok := raw.(manager.ExperimentManager)
	if !ok {
		emType := reflect.TypeOf((*manager.ExperimentManager)(nil)).Elem()
		return nil, fmt.Errorf("failed to cast plugin to type %v", emType)
	}

	return &experimentManagerPlugin{
		ExperimentManager: experimentManager,
	}, nil
}
