package mocks

import (
	"encoding/json"

	"github.com/caraml-dev/turing/engines/experiment/manager/mocks"
)

type ConfigurableExperimentManager struct {
	mocks.ExperimentManager
}

func (_m *ConfigurableExperimentManager) Configure(cfg json.RawMessage) error {
	ret := _m.Called(cfg)

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type ConfigurableStandardExperimentManager struct {
	mocks.StandardExperimentManager
}

func (_m *ConfigurableStandardExperimentManager) Configure(cfg json.RawMessage) error {
	ret := _m.Called(cfg)

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
