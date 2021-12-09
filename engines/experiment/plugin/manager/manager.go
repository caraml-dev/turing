package manager

import (
	"encoding/json"

	"github.com/gojek/turing/engines/experiment/manager"
)

type ConfigurableExperimentManager interface {
	Configure(cfg json.RawMessage) error
	manager.ExperimentManager
}
