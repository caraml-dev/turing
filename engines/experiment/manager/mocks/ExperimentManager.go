// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	json "encoding/json"

	manager "github.com/gojek/turing/engines/experiment/manager"
	mock "github.com/stretchr/testify/mock"
)

// ExperimentManager is an autogenerated mock type for the ExperimentManager type
type ExperimentManager struct {
	mock.Mock
}

// GetEngineInfo provides a mock function with given fields:
func (_m *ExperimentManager) GetEngineInfo() manager.Engine {
	ret := _m.Called()

	var r0 manager.Engine
	if rf, ok := ret.Get(0).(func() manager.Engine); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(manager.Engine)
	}

	return r0
}

// GetExperimentRunnerConfig provides a mock function with given fields: cfg
func (_m *ExperimentManager) GetExperimentRunnerConfig(cfg interface{}) (json.RawMessage, error) {
	ret := _m.Called(cfg)

	var r0 json.RawMessage
	if rf, ok := ret.Get(0).(func(interface{}) json.RawMessage); ok {
		r0 = rf(cfg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(json.RawMessage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(interface{}) error); ok {
		r1 = rf(cfg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateExperimentConfig provides a mock function with given fields: cfg
func (_m *ExperimentManager) ValidateExperimentConfig(cfg json.RawMessage) error {
	ret := _m.Called(cfg)

	var r0 error
	if rf, ok := ret.Get(0).(func(json.RawMessage) error); ok {
		r0 = rf(cfg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
