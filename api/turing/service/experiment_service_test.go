package service

import (
	"encoding/json"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"

	tu "github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/gojek/turing/engines/experiment/common"
	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/manager/mocks"
)

var standardExperimentManagerConfig = manager.Engine{Type: manager.StandardExperimentManagerType}
var customExperimentManagerConfig = manager.Engine{Type: manager.CustomExperimentManagerType}

func TestListEngines(t *testing.T) {
	// Set up mock Experiment Managers
	litmusEngineInfo := manager.Engine{
		Name: "Litmus",
		Type: manager.StandardExperimentManagerType,
		StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
			ClientSelectionEnabled:     true,
			ExperimentSelectionEnabled: true,
		},
	}
	xpEngineInfo := manager.Engine{
		Name: "XP",
		Type: manager.StandardExperimentManagerType,
		StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
			ClientSelectionEnabled:     false,
			ExperimentSelectionEnabled: false,
		},
	}
	expMgr1 := &mocks.ExperimentManager{}
	expMgr1.On("GetEngineInfo").Return(litmusEngineInfo)
	expMgr2 := &mocks.ExperimentManager{}
	expMgr2.On("GetEngineInfo").Return(xpEngineInfo)

	// Create the experiment managers map and the experiment service
	experimentManagers := make(map[string]manager.ExperimentManager)
	experimentManagers["litmus"] = expMgr1
	experimentManagers["xp"] = expMgr2
	svc := &experimentsService{
		experimentManagers: experimentManagers,
		cache:              cache.New(time.Second, time.Second),
	}

	// Run test and validate
	response := svc.ListEngines()
	// Sort items
	sort.SliceStable(response, func(i, j int) bool {
		return response[i].Name < response[j].Name
	})
	assert.Equal(t, []manager.Engine{litmusEngineInfo, xpEngineInfo}, response)
}

func TestListClients(t *testing.T) {
	clients := []manager.Client{
		{
			ID:       "1",
			Username: "test-client",
		},
	}
	// Set up mock Experiment Managers
	expMgrSuccess := &mocks.StandardExperimentManager{}
	expMgrSuccess.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig)
	expMgrSuccess.On("IsCacheEnabled").Return(true)
	expMgrSuccess.On("ListClients").Return(clients, nil)

	expMgrFailure := &mocks.StandardExperimentManager{}
	expMgrFailure.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig)
	expMgrFailure.On("IsCacheEnabled").Return(true)
	expMgrFailure.On("ListClients").Return([]manager.Client{}, errors.New("List clients error"))

	// Define tests
	tests := map[string]struct {
		expMgr      manager.ExperimentManager
		useBadCache bool
		expected    []manager.Client
		err         string
	}{
		"success": {
			expMgr:      expMgrSuccess,
			useBadCache: false,
			expected:    clients,
		},
		"failure | list clients": {
			expMgr:      expMgrFailure,
			useBadCache: false,
			expected:    []manager.Client{},
			err:         "List clients error",
		},
		"failure | bad cache": {
			expMgr:      expMgrSuccess,
			useBadCache: true,
			expected:    []manager.Client{},
			err:         "Malformed clients info found in the cache for engine litmus",
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Create experiment service
			cacheObj := cache.New(time.Second*2, time.Second*2)
			if data.useBadCache {
				cacheObj.Set("engine:litmus:clients", "test", cache.DefaultExpiration)
			}
			svc := &experimentsService{
				experimentManagers: map[string]manager.ExperimentManager{
					"litmus": data.expMgr,
				},
				cache: cacheObj,
			}

			// Run and Validate
			response, err := svc.ListClients("litmus")
			assert.Equal(t, data.expected, response)
			if data.err != "" {
				tu.FailOnNil(t, err)
				assert.Equal(t, data.err, err.Error())
			}

			// Access cache
			if data.err == "" {
				response, err := svc.ListClients("litmus")
				assert.Equal(t, data.expected, response)
				assert.NoError(t, err)
			}
		})
	}
}

func TestListExperiments(t *testing.T) {
	client := manager.Client{
		ID:       "1",
		Username: "test-client",
	}
	clients := []manager.Client{client}
	experiments := []manager.Experiment{
		{
			ID:       "2",
			ClientID: "1",
			Name:     "test-experiment",
		},
	}
	// Set up mock Experiment Managers
	expMgrSuccess := &mocks.StandardExperimentManager{}
	expMgrSuccess.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig)
	expMgrSuccess.On("IsCacheEnabled").Return(true)
	expMgrSuccess.On("ListClients").Return(clients, nil)
	expMgrSuccess.On("ListExperiments").Return(experiments, nil)
	expMgrSuccess.On("ListExperimentsForClient", client).Return(experiments, nil)

	expMgrFailure1 := &mocks.StandardExperimentManager{}
	expMgrFailure1.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig)
	expMgrFailure1.On("IsCacheEnabled").Return(true)
	expMgrFailure1.On("ListClients").Return([]manager.Client{}, errors.New("List clients error"))

	expMgrFailure2 := &mocks.StandardExperimentManager{}
	expMgrFailure2.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig)
	expMgrFailure2.On("IsCacheEnabled").Return(true)
	expMgrFailure2.On("ListClients").Return(clients, nil)
	expMgrFailure2.On("ListExperimentsForClient", client).
		Return([]manager.Experiment{}, errors.New("List experiments error"))

	// Define tests
	tests := map[string]struct {
		expMgr      manager.ExperimentManager
		clientID    string
		useBadCache bool
		expected    []manager.Experiment
		err         string
	}{
		"success | all experiments": {
			expMgr:      expMgrSuccess,
			useBadCache: false,
			expected:    experiments,
		},
		"success | experiments for client": {
			expMgr:      expMgrSuccess,
			clientID:    "1",
			useBadCache: false,
			expected:    experiments,
		},
		"failure | list clients": {
			expMgr:      expMgrFailure1,
			clientID:    "1",
			useBadCache: false,
			expected:    []manager.Experiment{},
			err:         "List clients error",
		},
		"failure | list experiments": {
			expMgr:      expMgrFailure2,
			clientID:    "1",
			useBadCache: false,
			expected:    []manager.Experiment{},
			err:         "List experiments error",
		},
		"failure | bad cache": {
			expMgr:      expMgrSuccess,
			clientID:    "1",
			useBadCache: true,
			expected:    []manager.Experiment{},
			err:         "Malformed experiments info found in the cache",
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Create experiment service
			cacheObj := cache.New(time.Second*2, time.Second*2)
			if data.useBadCache {
				cacheObj.Set("engine:litmus:clients:1:experiments", "test", cache.DefaultExpiration)
			}
			svc := &experimentsService{
				experimentManagers: map[string]manager.ExperimentManager{
					"litmus": data.expMgr,
				},
				cache: cacheObj,
			}

			// Run and Validate
			response, err := svc.ListExperiments("litmus", data.clientID)
			assert.Equal(t, data.expected, response)
			if data.err != "" {
				tu.FailOnNil(t, err)
				assert.Equal(t, data.err, err.Error())
			}

			// Access cache
			if data.err == "" {
				response, err := svc.ListExperiments("litmus", "1")
				assert.Equal(t, data.expected, response)
				assert.NoError(t, err)
			}
		})
	}
}

func TestListVariables(t *testing.T) {
	client := manager.Client{
		ID:       "1",
		Username: "test-client",
	}
	clients := []manager.Client{client}
	clientVariables := []manager.Variable{
		{
			Name:     "var-1",
			Required: true,
			Type:     manager.UnitVariableType,
		},
	}
	experiments := []manager.Experiment{
		{
			ID:       "2",
			ClientID: "1",
			Name:     "test-experiment",
		},
	}
	experimentVariables := map[string][]manager.Variable{
		"2": {
			{
				Name:     "var-1",
				Required: false,
				Type:     manager.FilterVariableType,
			},
			{
				Name:     "var-2",
				Required: false,
				Type:     manager.FilterVariableType,
			},
		},
	}

	// Set up mock Experiment Managers
	expMgrSuccess := &mocks.StandardExperimentManager{}
	expMgrSuccess.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig)
	expMgrSuccess.On("IsCacheEnabled").Return(true)
	expMgrSuccess.On("ListClients").Return(clients, nil)
	expMgrSuccess.On("ListVariablesForClient", client).Return(clientVariables, nil)
	expMgrSuccess.On("ListExperimentsForClient", client).Return(experiments, nil)
	expMgrSuccess.On("ListVariablesForExperiments", experiments).Return(experimentVariables, nil)
	expMgrSuccess.On("ListVariablesForExperiments", []manager.Experiment{}).
		Return(map[string][]manager.Variable{}, nil)

	expMgrFailure1 := &mocks.StandardExperimentManager{}
	expMgrFailure1.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig)
	expMgrFailure1.On("IsCacheEnabled").Return(true)
	expMgrFailure1.On("ListClients").Return([]manager.Client{}, errors.New("List clients error"))

	expMgrFailure2 := &mocks.StandardExperimentManager{}
	expMgrFailure2.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig)
	expMgrFailure2.On("IsCacheEnabled").Return(true)
	expMgrFailure2.On("ListClients").Return(clients, nil)
	expMgrFailure2.On("ListVariablesForClient", client).
		Return([]manager.Variable{}, errors.New("List client vars error"))

	expMgrFailure3 := &mocks.StandardExperimentManager{}
	expMgrFailure3.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig)
	expMgrFailure3.On("IsCacheEnabled").Return(true)
	expMgrFailure3.On("ListClients").Return(clients, nil)
	expMgrFailure3.On("ListVariablesForClient", client).Return(clientVariables, nil)
	expMgrFailure3.On("ListExperimentsForClient", client).
		Return([]manager.Experiment{}, errors.New("List experiments error"))

	expMgrFailure4 := &mocks.StandardExperimentManager{}
	expMgrFailure4.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig)
	expMgrFailure4.On("IsCacheEnabled").Return(true)
	expMgrFailure4.On("ListClients").Return(clients, nil)
	expMgrFailure4.On("ListVariablesForClient", client).Return(clientVariables, nil)
	expMgrFailure4.On("ListExperimentsForClient", client).Return(experiments, nil)
	expMgrFailure4.On("ListVariablesForExperiments", experiments).
		Return(map[string][]manager.Variable{}, errors.New("List experiment vars error"))

	// Define tests
	tests := map[string]struct {
		expMgr        manager.ExperimentManager
		clientID      string
		experimentIDs []string
		badCacheKey   string
		expected      manager.Variables
		err           string
	}{
		"success": {
			expMgr:        expMgrSuccess,
			clientID:      "1",
			experimentIDs: []string{"2"},
			expected: manager.Variables{
				ClientVariables:     clientVariables,
				ExperimentVariables: experimentVariables,
				Config: []manager.VariableConfig{
					{
						Name:        "var-1",
						Required:    true,
						FieldSource: common.HeaderFieldSource,
					},
					{
						Name:        "var-2",
						Required:    false,
						FieldSource: common.HeaderFieldSource,
					},
				},
			},
		},
		"failure | list clients": {
			expMgr:   expMgrFailure1,
			clientID: "1",
			err:      "List clients error",
		},
		"failure | list client vars": {
			expMgr:   expMgrFailure2,
			clientID: "1",
			err:      "List client vars error",
		},
		"failure | list experiments": {
			expMgr:        expMgrFailure3,
			clientID:      "1",
			experimentIDs: []string{"2"},
			err:           "List experiments error",
		},
		"failure | list experiment vars": {
			expMgr:        expMgrFailure4,
			clientID:      "1",
			experimentIDs: []string{"2"},
			err:           "List experiment vars error",
		},
		"failure | bad client vars cache": {
			expMgr:        expMgrSuccess,
			clientID:      "1",
			experimentIDs: []string{"2"},
			badCacheKey:   "engine:litmus:clients:1:variables",
			err:           "Malformed variables info found in the cache for client 1",
		},
		"failure | bad experiment vars cache": {
			expMgr:        expMgrSuccess,
			clientID:      "1",
			experimentIDs: []string{"2"},
			badCacheKey:   "engine:litmus:experiments:2:variables",
			err:           "Malformed variables info found in the cache for experiment 2",
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Create experiment service
			cacheObj := cache.New(time.Second*2, time.Second*2)
			if data.badCacheKey != "" {
				cacheObj.Set(data.badCacheKey, "test", cache.DefaultExpiration)
			}
			svc := &experimentsService{
				experimentManagers: map[string]manager.ExperimentManager{
					"litmus": data.expMgr,
				},
				cache: cacheObj,
			}

			// Run and Validate
			response, err := svc.ListVariables("litmus", data.clientID, data.experimentIDs)
			// Sort items
			sort.SliceStable(response.Config, func(i, j int) bool {
				return response.Config[i].Name < response.Config[j].Name
			})
			assert.Equal(t, data.expected, response)
			if data.err != "" {
				tu.FailOnNil(t, err)
				assert.Equal(t, data.err, err.Error())
			}

			// Access cache
			if data.err == "" {
				response, err := svc.ListVariables("litmus", data.clientID, data.experimentIDs)
				// Sort items
				sort.SliceStable(response.Config, func(i, j int) bool {
					return response.Config[i].Name < response.Config[j].Name
				})
				assert.Equal(t, data.expected, response)
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateExperimentConfig(t *testing.T) {
	// Create mock experiment managers
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(manager.Engine{
		Name: "Litmus",
		Type: manager.StandardExperimentManagerType,
		StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
			ClientSelectionEnabled: true,
		},
	})
	stdExpMgr.On("ValidateExperimentConfig",
		&manager.StandardExperimentManagerConfig{
			ClientSelectionEnabled: true,
		},
		manager.TuringExperimentConfig{},
	).Return(nil)

	customExpMgr := &mocks.CustomExperimentManager{}
	customExpMgr.On("GetEngineInfo").Return(customExperimentManagerConfig)
	customExpMgr.On("ValidateExperimentConfig", []int{1, 2}).Return(nil)

	// Create test experiment service
	expSvc := &experimentsService{
		experimentManagers: map[string]manager.ExperimentManager{
			"litmus": stdExpMgr,
			"xp":     customExpMgr,
		},
	}

	// Define tests
	tests := map[string]struct {
		engine   string
		inputCfg interface{}
		err      string
	}{
		"failure | std exp mgr": {
			engine: "test-engine",
			err:    "Unknown experiment engine test-engine",
		},
		"success | std exp mgr": {
			engine:   "litmus",
			inputCfg: manager.TuringExperimentConfig{},
		},
		"success | custom exp mgr": {
			engine:   "xp",
			inputCfg: []int{1, 2},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			err := expSvc.ValidateExperimentConfig(data.engine, data.inputCfg)
			if data.err == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if err != nil {
					assert.Equal(t, data.err, err.Error())
				}
			}
		})
	}
}

func TestGetExperimentRunnerConfig(t *testing.T) {
	expectedResult1 := json.RawMessage([]byte(`{"key": "value1"}`))
	expectedResult2 := json.RawMessage([]byte(`{"key": "value2"}`))
	// Create mock experiment managers
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(standardExperimentManagerConfig)
	stdExpMgr.On("GetExperimentRunnerConfig", manager.TuringExperimentConfig{
		Client: manager.Client{
			ID: "1",
		},
	}).Return(expectedResult1, nil)

	customExpMgr := &mocks.CustomExperimentManager{}
	customExpMgr.On("GetEngineInfo").Return(customExperimentManagerConfig)
	customExpMgr.On("GetExperimentRunnerConfig", json.RawMessage([]byte(`[1,2]`))).Return(expectedResult2, nil)

	// Create test experiment service
	expSvc := &experimentsService{
		experimentManagers: map[string]manager.ExperimentManager{
			"litmus": stdExpMgr,
			"xp":     customExpMgr,
		},
	}

	// Define tests
	tests := map[string]struct {
		engine         string
		inputCfg       json.RawMessage
		expectedResult json.RawMessage
		err            string
	}{
		"failure | std exp mgr": {
			engine:         "test-engine",
			expectedResult: json.RawMessage{},
			err:            "Unknown experiment engine test-engine",
		},
		"success | std exp mgr": {
			engine:         "litmus",
			inputCfg:       json.RawMessage([]byte(`{"client": {"id": "1"}}`)),
			expectedResult: expectedResult1,
		},
		"success | custom exp mgr": {
			engine:         "xp",
			inputCfg:       json.RawMessage([]byte(`[1,2]`)),
			expectedResult: expectedResult2,
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			jsonCfg, err := expSvc.GetExperimentRunnerConfig(data.engine, data.inputCfg)
			assert.Equal(t, data.expectedResult, jsonCfg)
			if data.err == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if err != nil {
					assert.Equal(t, data.err, err.Error())
				}
			}
		})
	}
}
