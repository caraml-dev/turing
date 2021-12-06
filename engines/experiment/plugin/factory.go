package plugin

import (
	"errors"
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

func NewFactory(builder *Configuration) (*factory, error) {
	services, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return &factory{services}, nil
}
