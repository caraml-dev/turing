package manager

import (
	"encoding/json"

	"github.com/caraml-dev/turing/engines/experiment/manager"
	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/shared"
)

// ConfigurableExperimentManager interface of an ExperimentManager, that can be configured
// with an arbitrary configuration passed as a JSON data
type ConfigurableExperimentManager interface {
	shared.Configurable
	manager.ExperimentManager
}

func NewConfigurableStandardExperimentManager(
	factory func(cfg json.RawMessage) (manager.StandardExperimentManager, error),
) ConfigurableExperimentManager {
	return &configurableStandardExperimentManager{factory: factory}
}

type configurableStandardExperimentManager struct {
	manager.StandardExperimentManager
	factory func(cfg json.RawMessage) (manager.StandardExperimentManager, error)
}

func (em *configurableStandardExperimentManager) Configure(cfg json.RawMessage) (err error) {
	em.StandardExperimentManager, err = em.factory(cfg)
	return
}

func NewConfigurableCustomExperimentManager(
	factory func(cfg json.RawMessage) (manager.CustomExperimentManager, error),
) ConfigurableExperimentManager {
	return &configurableCustomExperimentManager{factory: factory}
}

type configurableCustomExperimentManager struct {
	manager.CustomExperimentManager
	factory func(cfg json.RawMessage) (manager.CustomExperimentManager, error)
}

func (em *configurableCustomExperimentManager) Configure(cfg json.RawMessage) (err error) {
	em.CustomExperimentManager, err = em.factory(cfg)
	return
}
