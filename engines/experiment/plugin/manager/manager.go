package manager

import (
	"github.com/gojek/turing/engines/experiment/plugin/shared"

	"github.com/gojek/turing/engines/experiment/manager"
)

// ConfigurableExperimentManager interface of an ExperimentManager, that can be configured
// with an arbitrary configuration passed as a JSON data
type ConfigurableExperimentManager interface {
	shared.Configurable
	manager.ExperimentManager
}
