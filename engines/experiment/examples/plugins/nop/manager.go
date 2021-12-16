package nop

import (
	"encoding/json"
	"github.com/gojek/turing/engines/experiment/manager"
)

type ExperimentManager struct {
	displayName string
}

func (m *ExperimentManager) Configure(rawCfg json.RawMessage) error {
	cfg := struct {
		DisplayName string `json:"display_name"`
	}{}

	if err := json.Unmarshal(rawCfg, &cfg); err != nil {
		return err
	}
	m.displayName = cfg.DisplayName
	return nil
}

func (m *ExperimentManager) GetEngineInfo() manager.Engine {
	return manager.Engine{
		Name:        "nop",
		DisplayName: m.displayName,
		Type:        manager.StandardExperimentManagerType,
		StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
			ClientSelectionEnabled:     false,
			ExperimentSelectionEnabled: false,
			HomePageURL:                "http://example.com",
		},
	}
}

func (*ExperimentManager) ValidateExperimentConfig(json.RawMessage) error {
	return nil
}
