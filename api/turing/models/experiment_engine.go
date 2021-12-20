package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

const (
	ExperimentEngineTypeNop    string = "nop"
	ExperimentEngineTypeLitmus string = "litmus"
	ExperimentEngineTypeXp     string = "xp"
)

// ExperimentEngine contains the type and configuration for the
// Experiment engine powering the router.
type ExperimentEngine struct {
	// Type of Experiment Engine
	Type string `json:"type"`
	// Config contains the configs for the selected experiment engine (other than "nop").
	// For standard experiment engine managers, the config can be unmarshalled into
	// manager.TuringExperimentConfig type.
	Config json.RawMessage `json:"config,omitempty"`
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
