package service

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/plugin"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/patrickmn/go-cache"
)

func NewExperimentsServiceV2(cfg config.ExperimentsConfig) (ExperimentsService, error) {
	experimentManagers := make(map[string]manager.ExperimentManager)

	for name, engineCfg := range cfg.Engines {
		engineCfgJSON, _ := json.Marshal(engineCfg)

		builder := &plugin.Configuration{
			PluginBinary:   filepath.Join(cfg.PluginsDirectory, name),
			PluginConfig:   engineCfgJSON,
			PluginLogLevel: "info",
		}

		services, err := builder.Build()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load %s experiment manager plugin", name)
		}
		experimentManagers[name] = services.Manager
	}

	// Initialize the experimentsService with cache
	svc := &experimentsService{
		experimentManagers: experimentManagers,
		cache:              cache.New(expCacheExpirySeconds*time.Second, expCacheCleanUpSeconds*time.Second),
	}

	return svc, nil
}
