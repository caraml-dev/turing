package plugin

import (
	"errors"
	"go.uber.org/zap"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/runner"
)

type factory struct {
	*Services
}

func (f *factory) GetExperimentManager() (manager.ExperimentManager, error) {
	return f.Manager, nil
}

func (f *factory) GetExperimentRunner() (runner.ExperimentRunner, error) {
	return nil, errors.New("not implemented")
}

func NewFactory(builder *Configuration, logger *zap.Logger) (*factory, error) {
	services, err := builder.Build(logger)
	if err != nil {
		return nil, err
	}

	return &factory{services}, nil
}
