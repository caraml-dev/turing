package manager

import (
	"github.com/gojek/turing/engines/experiment/common"
	"github.com/gojek/turing/engines/experiment/manager"
)

type ConfigurableExperimentManager interface {
	common.Configurable
	manager.ExperimentManager
}
