package manager

import (
	"github.com/gojek/turing/engines/experiment"
	"github.com/gojek/turing/engines/experiment/manager"
)

type ConfigurableExperimentManager interface {
	experiment.Configurable
	manager.ExperimentManager
}
