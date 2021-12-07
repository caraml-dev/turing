package manager

import (
	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/pkg/types"
)

type ConfigurableExperimentManager interface {
	types.Configurable
	manager.ExperimentManager
}
