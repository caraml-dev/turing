package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/caraml-dev/turing/api/turing/config"
)

const (
	ExperimentEngineTypeNop string = "nop"
)

// ExperimentEngine contains the type and configuration for the
// Experiment engine powering the router.
type ExperimentEngine struct {
	// Type of Experiment Engine
	Type string `json:"type"`
	// PluginConfig (Optional) contains necessary configuration, if experiment engine
	// is implemented as RPC plugin
	PluginConfig *config.ExperimentEnginePluginConfig `json:"plugin_config,omitempty"`
	// ServiceAccount contains the service account file path that points to the service account
	// key that the experiment engine may require
	ServiceAccountKeyFilePath *string `json:"-"`
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
