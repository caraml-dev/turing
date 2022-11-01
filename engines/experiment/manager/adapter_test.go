package manager_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/experiment/manager"
	"github.com/caraml-dev/turing/engines/experiment/manager/mocks"
)

func TestStandardExperimentIsCacheEnabled(t *testing.T) {
	// Set up mocks
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType}, nil)
	stdExpMgr.On("IsCacheEnabled").Return(true, nil)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{}, nil)

	// IsCacheEnabled
	assert.True(t, manager.IsCacheEnabled(stdExpMgr))
	assert.False(t, manager.IsCacheEnabled(expMgr))
}

func TestStandardExperimentListClients(t *testing.T) {
	stdExpMgrErr := "Method is only supported by standard experiment managers"

	// Set up mocks
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType}, nil)
	stdExpMgr.On("ListClients").Return([]manager.Client{{}}, nil)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{}, nil)

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
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType}, nil)
	stdExpMgr.On("ListExperiments").Return([]manager.Experiment{{}}, nil)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{}, nil)

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
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType}, nil)
	stdExpMgr.On("ListExperimentsForClient", manager.Client{ID: "1"}).
		Return([]manager.Experiment{{}}, nil)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{}, nil)

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
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType}, nil)
	stdExpMgr.On("ListVariablesForClient", manager.Client{ID: "1"}).
		Return([]manager.Variable{{Name: "var-1"}}, nil)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{}, nil)

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
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{Type: manager.StandardExperimentManagerType}, nil)
	stdExpMgr.On("ListVariablesForExperiments", []manager.Experiment{{}}).
		Return(map[string][]manager.Variable{
			"test-exp": {{Name: "var-1"}},
		}, nil)

	expMgr := &mocks.ExperimentManager{}
	expMgr.On("GetEngineInfo").Return(manager.Engine{}, nil)

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
