package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// EngineType is used to capture the types of the supported experimentation engines
// Need to expose this value in turing-router
// engines/router/missionctl/experiment/experiment.go
type ExperimentEngineType string

const (
	ExperimentEngineTypeNop    ExperimentEngineType = "nop"
	ExperimentEngineTypeLitmus ExperimentEngineType = "litmus"
	ExperimentEngineTypeXp     ExperimentEngineType = "xp"
)

// ExperimentEngine contains the type and configuration for the
// Experiment engine powering the router.
type ExperimentEngine struct {
	// Type of Experiment Engine. Currently supports "litmus", "xp" and "nop".
	Type ExperimentEngineType `json:"type"`
	// Config contains the configs for the selected experiment engine (other than "nop").
	// For standard experiment engine managers, the config can be cast into TuringExperimentConfig type.
	Config interface{} `json:"config,omitempty"`
}

func (eec ExperimentEngine) Value() (driver.Value, error) {
	return json.Marshal(eec)
}

func (eec *ExperimentEngine) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &eec)
}
