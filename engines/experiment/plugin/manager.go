package plugin

import (
	"github.com/gojek/turing/engines/experiment/manager"
)

type ConfigurableExperimentManager interface {
	Configurable
	manager.ExperimentManager
}
