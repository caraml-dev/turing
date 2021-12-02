package manager_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/manager/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStandardExperimentIsCacheEnabled(t *testing.T) {
	// Set up mocks
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType})
	stdExpMgr.On("IsCacheEnabled").Return(true)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{})

	// Validate
	// IsCacheEnabled
	assert.Equal(t, true, manager.IsCacheEnabled(stdExpMgr))
	assert.Equal(t, false, manager.IsCacheEnabled(expMgr))
}

func TestStandardExperimentListClients(t *testing.T) {
	stdExpMgrErr := "Method is only supported by standard experiment managers"

	// Set up mocks
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType})
	stdExpMgr.On("ListClients").Return([]manager.Client{{}}, nil)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{})

	// Validate
	// ListClients
	clients, err := manager.ListClients(stdExpMgr)
	assert.Equal(t, []manager.Client{{}}, clients)
	assert.NoError(t, err)
	clients, err = manager.ListClients(expMgr)
	assert.Equal(t, []manager.Client{}, clients)
	assert.EqualError(t, err, stdExpMgrErr)
}

func TestStandardExperimentListExperiments(t *testing.T) {
	stdExpMgrErr := "Method is only supported by standard experiment managers"

	// Set up mocks
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType})
	stdExpMgr.On("ListExperiments").Return([]manager.Experiment{{}}, nil)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{})

	// Validate
	// ListExperiments
	experiments, err := manager.ListExperiments(stdExpMgr)
	assert.Equal(t, []manager.Experiment{{}}, experiments)
	assert.NoError(t, err)
	experiments, err = manager.ListExperiments(expMgr)
	assert.Equal(t, []manager.Experiment{}, experiments)
	assert.EqualError(t, err, stdExpMgrErr)
}

func TestStandardExperimentListExperimentsForClient(t *testing.T) {
	stdExpMgrErr := "Method is only supported by standard experiment managers"

	// Set up mocks
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType})
	stdExpMgr.On("ListExperimentsForClient", manager.Client{ID: "1"}).
		Return([]manager.Experiment{{}}, nil)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{})

	// Validate
	// ListExperimentsForClient
	experiments, err := manager.ListExperimentsForClient(stdExpMgr, manager.Client{ID: "1"})
	assert.Equal(t, []manager.Experiment{{}}, experiments)
	assert.NoError(t, err)
	experiments, err = manager.ListExperimentsForClient(expMgr, manager.Client{ID: "1"})
	assert.Equal(t, []manager.Experiment{}, experiments)
	assert.EqualError(t, err, stdExpMgrErr)
}

func TestStandardExperimentListVariablesForClient(t *testing.T) {
	stdExpMgrErr := "Method is only supported by standard experiment managers"

	// Set up mocks
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType})
	stdExpMgr.On("ListVariablesForClient", manager.Client{ID: "1"}).
		Return([]manager.Variable{{Name: "var-1"}}, nil)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{})

	// Validate
	// ListVariablesForClient
	variables, err := manager.ListVariablesForClient(stdExpMgr, manager.Client{ID: "1"})
	assert.Equal(t, []manager.Variable{{Name: "var-1"}}, variables)
	assert.NoError(t, err)
	variables, err = manager.ListVariablesForClient(expMgr, manager.Client{ID: "1"})
	assert.Equal(t, []manager.Variable{}, variables)
	assert.EqualError(t, err, stdExpMgrErr)
}

func TestStandardExperimentListVariablesForExperiments(t *testing.T) {
	stdExpMgrErr := "Method is only supported by standard experiment managers"

	// Set up mocks
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType})
	stdExpMgr.On("ListVariablesForExperiments", []manager.Experiment{{}}).
		Return(map[string][]manager.Variable{
			"test-exp": {{Name: "var-1"}},
		}, nil)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{})

	// Validate
	// ListVariablesForExperiments
	variablesMap, err := manager.ListVariablesForExperiments(stdExpMgr, []manager.Experiment{{}})
	assert.Equal(t, map[string][]manager.Variable{
		"test-exp": {{Name: "var-1"}},
	}, variablesMap)
	assert.NoError(t, err)
	variablesMap, err = manager.ListVariablesForExperiments(expMgr, []manager.Experiment{{}})
	assert.Equal(t, map[string][]manager.Variable{}, variablesMap)
	assert.EqualError(t, err, stdExpMgrErr)
}

func TestGetExperimentRunnerConfig(t *testing.T) {
	testStdExpConfig := manager.TuringExperimentConfig{
		Client:      manager.Client{Username: "client_name"},
		Experiments: []manager.Experiment{{Name: "exp_name"}},
		Variables:   manager.Variables{ClientVariables: []manager.Variable{{Name: "var_name"}}},
	}

	// Get test data
	testData, err := ioutil.ReadFile(filepath.Join("testdata", "experiment_runner_config.json"))
	require.NoError(t, err)
	var testRawConfig interface{}
	err = json.Unmarshal(testData, &testRawConfig)
	assert.NoError(t, err)
	testSuccessResponse := json.RawMessage([]byte(`{}`))

	// Set up mock experiment managers
	// Standard Exp Manager
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType})
	stdExpMgr.On("GetExperimentRunnerConfig", testStdExpConfig).Return(testSuccessResponse, nil)
	stdExpMgr.On("GetExperimentRunnerConfig", mock.Anything).Return(nil, errors.New("Unexpected Parameters"))
	// Custom Exp Manager
	customExpMgr := &mocks.CustomExperimentManager{}
	customExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.CustomExperimentManagerType})
	customExpMgr.On("GetExperimentRunnerConfig", testRawConfig).Return(testSuccessResponse, nil)
	customExpMgr.On("GetExperimentRunnerConfig", mock.Anything).Return(nil, errors.New("Unexpected Parameters"))
	// Bad Experiment Managers
	badStdExpMgr := &mocks.ExperimentManager{}
	badStdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType})
	badCustomExpMgr := &mocks.ExperimentManager{}
	badCustomExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.CustomExperimentManagerType})
	badExpMgr := &mocks.ExperimentManager{}
	badExpMgr.On("GetEngineInfo").Return(manager.Engine{})

	// Set up tests
	tests := map[string]struct {
		mgr      manager.ExperimentManager
		rawCfg   interface{}
		expected json.RawMessage
		err      string
	}{
		"success | standard": {
			mgr:      stdExpMgr,
			rawCfg:   testStdExpConfig,
			expected: testSuccessResponse,
		},
		"success | custom": {
			mgr:      customExpMgr,
			rawCfg:   testRawConfig,
			expected: testSuccessResponse,
		},
		"failure | std mismatched type": {
			mgr: badStdExpMgr,
			err: "Error casting  to standard experiment manager",
		},
		"failure | std config error": {
			mgr:    stdExpMgr,
			rawCfg: []int{},
			err: strings.Join([]string{
				"Unable to parse standard experiment config: ",
				"json: cannot unmarshal array into Go value of type manager.TuringExperimentConfig"}, ""),
		},
		"failure | custom mismatched type": {
			mgr: badCustomExpMgr,
			err: "Error casting  to custom experiment manager",
		},
		"failure | unknown exp manager type": {
			mgr: badExpMgr,
			err: "Experiment Manager type  is not recognized",
		},
	}

	// Test calls to the correct experiment manager method, based on the type
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := manager.GetExperimentRunnerConfig(data.mgr, data.rawCfg)
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, resp)
			}
		})
	}
}

func TestAdapterValidateExperimentConfig(t *testing.T) {
	testStdExpConfig := manager.TuringExperimentConfig{
		Client:      manager.Client{Username: "client_name"},
		Experiments: []manager.Experiment{{Name: "exp_name"}},
		Variables:   manager.Variables{ClientVariables: []manager.Variable{{Name: "var_name"}}},
	}

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
	stdExpMgr.On("ValidateExperimentConfig", testStdExpMgrConfig, testStdExpConfig).Return(nil)
	stdExpMgr.On("ValidateExperimentConfig", mock.Anything).Return(errors.New("Unexpected Parameters"))
	// Custom Exp Manager
	customExpMgr := &mocks.CustomExperimentManager{}
	customExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.CustomExperimentManagerType})
	customExpMgr.On("ValidateExperimentConfig", testRawConfig).Return(nil)
	customExpMgr.On("ValidateExperimentConfig", mock.Anything).Return(errors.New("Unexpected Parameters"))

	// Bad Experiment Managers
	badStdExpMgr := &mocks.ExperimentManager{}
	badStdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType})
	badCustomExpMgr := &mocks.ExperimentManager{}
	badCustomExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.CustomExperimentManagerType})
	badExpMgr := &mocks.ExperimentManager{}
	badExpMgr.On("GetEngineInfo").Return(manager.Engine{})

	// Set up tests
	tests := map[string]struct {
		mgr    manager.ExperimentManager
		rawCfg interface{}
		err    string
	}{
		"success | standard": {
			mgr:    stdExpMgr,
			rawCfg: testStdExpConfig,
		},
		"success | custom": {
			mgr:    customExpMgr,
			rawCfg: testRawConfig,
		},
		"failure | std mismatched type": {
			mgr: badStdExpMgr,
			err: "Error casting  to standard experiment manager",
		},
		"failure | std config error": {
			mgr:    stdExpMgr,
			rawCfg: []int{},
			err: strings.Join([]string{
				"Unable to parse standard experiment config: ",
				"json: cannot unmarshal array into Go value of type manager.TuringExperimentConfig"}, ""),
		},
		"failure | custom mismatched type": {
			mgr: badCustomExpMgr,
			err: "Error casting  to custom experiment manager",
		},
		"failure | unknown exp manager type": {
			mgr: badExpMgr,
			err: "Experiment Manager type  is not recognized",
		},
	}

	// Test calls to the correct experiment manager method, based on the type
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			err := manager.ValidateExperimentConfig(data.mgr, data.rawCfg)
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
