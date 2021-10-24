package manager_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/manager/mocks"
)

func TestGetExperimentRunnerConfig(t *testing.T) {
	var testRawConfig interface{}
	err := json.Unmarshal([]byte(`{
		"client": {"username": "client_name"},
		"experiments": [{"name": "exp_name"}],
		"variables": {"client_variables": [{"name": "var_name"}]}
	}`), &testRawConfig)
	assert.NoError(t, err)
	testSuccessResponse := json.RawMessage([]byte(`{}`))

	// Set up mock experiment managers
	// Standard Exp Manager
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType})
	stdExpMgr.On("GetExperimentRunnerConfig", manager.TuringExperimentConfig{
		Client:      manager.Client{Username: "client_name"},
		Experiments: []manager.Experiment{{Name: "exp_name"}},
		Variables:   manager.Variables{ClientVariables: []manager.Variable{{Name: "var_name"}}},
	}).Return(testSuccessResponse, nil)
	stdExpMgr.On("GetExperimentRunnerConfig", mock.Anything).Return(nil, errors.New("Unexpected Parameters"))
	// Custom Exp Manager
	customExpMgr := &mocks.CustomExperimentManager{}
	customExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.CustomExperimentManagerType})
	customExpMgr.On("GetExperimentRunnerConfig", testRawConfig).Return(testSuccessResponse, nil)
	customExpMgr.On("GetExperimentRunnerConfig", mock.Anything).Return(nil, errors.New("Unexpected Parameters"))

	// Test calls to the correct experiment manager method, based on the type
	resp, err := manager.GetExperimentRunnerConfig(stdExpMgr, testRawConfig)
	assert.NoError(t, err)
	assert.Equal(t, testSuccessResponse, resp)
	resp, err = manager.GetExperimentRunnerConfig(customExpMgr, testRawConfig)
	assert.NoError(t, err)
	assert.Equal(t, testSuccessResponse, resp)
}

func TestValidateExperimentConfig(t *testing.T) {
	var testRawConfig interface{}
	err := json.Unmarshal([]byte(`{
		"client": {"username": "client_name"},
		"experiments": [{"name": "exp_name"}],
		"variables": {"client_variables": [{"name": "var_name"}]}
	}`), &testRawConfig)
	assert.NoError(t, err)
	testStdExpMgrConfig := &manager.StandardExperimentManagerConfig{}

	// Set up mock experiment managers
	// Standard Exp Manager
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{
		Type:                            manager.StandardExperimentManagerType,
		StandardExperimentManagerConfig: testStdExpMgrConfig,
	})
	stdExpMgr.On("ValidateExperimentConfig",
		testStdExpMgrConfig,
		manager.TuringExperimentConfig{
			Client:      manager.Client{Username: "client_name"},
			Experiments: []manager.Experiment{{Name: "exp_name"}},
			Variables:   manager.Variables{ClientVariables: []manager.Variable{{Name: "var_name"}}},
		},
	).Return(nil)
	stdExpMgr.On("ValidateExperimentConfig", mock.Anything).Return(errors.New("Unexpected Parameters"))
	// Custom Exp Manager
	customExpMgr := &mocks.CustomExperimentManager{}
	customExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.CustomExperimentManagerType})
	customExpMgr.On("ValidateExperimentConfig", testRawConfig).Return(nil)
	customExpMgr.On("ValidateExperimentConfig", mock.Anything).Return(errors.New("Unexpected Parameters"))

	// Test calls to the correct experiment manager method, based on the type
	err = manager.ValidateExperimentConfig(stdExpMgr, testRawConfig)
	assert.NoError(t, err)
	err = manager.ValidateExperimentConfig(customExpMgr, testRawConfig)
	assert.NoError(t, err)
}
