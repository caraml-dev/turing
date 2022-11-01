package nop

import (
	"encoding/json"

	"github.com/caraml-dev/turing/engines/experiment/manager"
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

func (m *ExperimentManager) GetEngineInfo() (manager.Engine, error) {
	return manager.Engine{
		Name:        "nop",
		DisplayName: m.displayName,
		Type:        manager.StandardExperimentManagerType,
		StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
			ClientSelectionEnabled:     false,
			ExperimentSelectionEnabled: false,
			HomePageURL:                "http://example.com",
		},
	}, nil
}

func (*ExperimentManager) ValidateExperimentConfig(json.RawMessage) error {
	return nil
}

func (*ExperimentManager) GetExperimentRunnerConfig(json.RawMessage) (json.RawMessage, error) {
	return json.RawMessage{}, nil
}
