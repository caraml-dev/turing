package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/manager/mocks"
	"github.com/gojek/turing/engines/experiment/pkg/request"
)

var standardExperimentManagerConfig = manager.Engine{Type: manager.StandardExperimentManagerType}
var customExperimentManagerConfig = manager.Engine{Type: manager.CustomExperimentManagerType}

func TestIsStandardExperimentManager(t *testing.T) {
	tests := map[string]struct {
		engineInfo    manager.Engine
		engineInfoErr string
		expected      bool
	}{
		"standard": {
			engineInfo: manager.Engine{
				Name: "standard-engine-1",
				Type: manager.StandardExperimentManagerType,
			},
			expected: true,
		},
		"error": {
			engineInfo: manager.Engine{
				Name: "standard-engine-2",
				Type: manager.StandardExperimentManagerType,
			},
			engineInfoErr: "test error",
		},
		"custom": {
			engineInfo: manager.Engine{
				Name: "custom-engine",
				Type: manager.CustomExperimentManagerType,
			},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Set up mock API call
			expMgr := &mocks.ExperimentManager{}
			var err error
			if data.engineInfoErr != "" {
				err = errors.New(data.engineInfoErr)
			}
			expMgr.On("GetEngineInfo").Return(data.engineInfo, err)
			// Set up experiment service
			svc := &experimentsService{
				experimentManagers: map[string]manager.ExperimentManager{
					data.engineInfo.Name: expMgr,
				},
				cache: cache.New(time.Second, time.Second),
			}
			// Validate
			isStdEngine := svc.IsStandardExperimentManager(data.engineInfo.Name)
			assert.Equal(t, data.expected, isStdEngine)
		})
	}
}

func TestIsClientSelectionEnabled(t *testing.T) {
	// Set up mock experiment managers
	tests := map[string]struct {
		engineInfo    manager.Engine
		engineInfoErr string
		expectedErr   string
		expected      bool
	}{
		"standard | no client selection": {
			engineInfo: manager.Engine{
				Name: "standard-engine-1",
				Type: manager.StandardExperimentManagerType,
				StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
					ClientSelectionEnabled: false,
				},
			},
		},
		"standard | with client selection": {
			engineInfo: manager.Engine{
				Name: "standard-engine-2",
				Type: manager.StandardExperimentManagerType,
				StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
					ClientSelectionEnabled: true,
				},
			},
			expected: true,
		},
		"custom": {
			engineInfo: manager.Engine{
				Name: "custom-engine-1",
				Type: manager.CustomExperimentManagerType,
			},
		},
		"error": {
			engineInfo: manager.Engine{
				Name: "custom-engine-2",
				Type: manager.CustomExperimentManagerType,
			},
			engineInfoErr: "test error",
			expectedErr:   "test error",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Set up mock API call
			expMgr := &mocks.ExperimentManager{}
			var err error
			if data.engineInfoErr != "" {
				err = errors.New(data.engineInfoErr)
			}
			expMgr.On("GetEngineInfo").Return(data.engineInfo, err)
			// Set up experiment service
			svc := &experimentsService{
				experimentManagers: map[string]manager.ExperimentManager{
					data.engineInfo.Name: expMgr,
				},
				cache: cache.New(time.Second, time.Second),
			}
			// Validate
			isClientSelectionEnabled, err := svc.IsClientSelectionEnabled(data.engineInfo.Name)
			if data.expectedErr != "" {
				assert.EqualError(t, err, data.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, isClientSelectionEnabled)
			}
		})
	}
}

func TestListEngines(t *testing.T) {
	// Set up mock Experiment Managers
	standardEngineInfo := manager.Engine{
		Name: "standard-engine",
		Type: manager.StandardExperimentManagerType,
		StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
			ClientSelectionEnabled:     true,
			ExperimentSelectionEnabled: true,
		},
	}
	customEngineInfo := manager.Engine{
		Name: "custom-engine",
		Type: manager.StandardExperimentManagerType,
		StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
			ClientSelectionEnabled:     false,
			ExperimentSelectionEnabled: false,
		},
	}
	expMgr1 := &mocks.ExperimentManager{}
	expMgr1.On("GetEngineInfo").Return(standardEngineInfo, nil)
	expMgr2 := &mocks.ExperimentManager{}
	expMgr2.On("GetEngineInfo").Return(customEngineInfo, nil)

	// Create the experiment managers map and the experiment service
	experimentManagers := make(map[string]manager.ExperimentManager)
	experimentManagers["standard-engine"] = expMgr1
	experimentManagers["custom-engine"] = expMgr2
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
	assert.Equal(t, []manager.Engine{customEngineInfo, standardEngineInfo}, response)
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
	expMgrSuccess.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig, nil)
	expMgrSuccess.On("IsCacheEnabled").Return(true, nil)
	expMgrSuccess.On("ListClients").Return(clients, nil)

	expMgrFailure := &mocks.StandardExperimentManager{}
	expMgrFailure.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig, nil)
	expMgrFailure.On("IsCacheEnabled").Return(true, nil)
	expMgrFailure.On("ListClients").Return([]manager.Client{}, errors.New("List clients error"))

	expManagerName := "exp-engine-000"
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
			err:         fmt.Sprintf("Malformed clients info found in the cache for engine %s", expManagerName),
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Create experiment service
			cacheObj := cache.New(time.Second*2, time.Second*2)
			if data.useBadCache {
				cacheObj.Set(fmt.Sprintf("engine:%s:clients", expManagerName), "test", cache.DefaultExpiration)
			}
			svc := &experimentsService{
				experimentManagers: map[string]manager.ExperimentManager{
					expManagerName: data.expMgr,
				},
				cache: cacheObj,
			}

			// Run and Validate
			response, err := svc.ListClients(expManagerName)
			assert.Equal(t, data.expected, response)
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				response, err := svc.ListClients(expManagerName)
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
	expMgrSuccess.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig, nil)
	expMgrSuccess.On("IsCacheEnabled").Return(true, nil)
	expMgrSuccess.On("ListClients").Return(clients, nil)
	expMgrSuccess.On("ListExperiments").Return(experiments, nil)
	expMgrSuccess.On("ListExperimentsForClient", client).Return(experiments, nil)

	expMgrFailure1 := &mocks.StandardExperimentManager{}
	expMgrFailure1.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig, nil)
	expMgrFailure1.On("IsCacheEnabled").Return(true, nil)
	expMgrFailure1.On("ListClients").Return([]manager.Client{}, errors.New("List clients error"))

	expMgrFailure2 := &mocks.StandardExperimentManager{}
	expMgrFailure2.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig, nil)
	expMgrFailure2.On("IsCacheEnabled").Return(true, nil)
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

	expManagerName := "exp-engine-001"
	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Create experiment service
			cacheObj := cache.New(time.Second*2, time.Second*2)
			if data.useBadCache {
				cacheObj.Set(
					fmt.Sprintf("engine:%s:clients:1:experiments", expManagerName),
					"test",
					cache.DefaultExpiration,
				)
			}
			svc := &experimentsService{
				experimentManagers: map[string]manager.ExperimentManager{
					expManagerName: data.expMgr,
				},
				cache: cacheObj,
			}

			// Run and Validate
			response, err := svc.ListExperiments(expManagerName, data.clientID)
			assert.Equal(t, data.expected, response)
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				response, err := svc.ListExperiments(expManagerName, "1")
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
	expMgrSuccess.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig, nil)
	expMgrSuccess.On("IsCacheEnabled").Return(true, nil)
	expMgrSuccess.On("ListClients").Return(clients, nil)
	expMgrSuccess.On("ListVariablesForClient", client).Return(clientVariables, nil)
	expMgrSuccess.On("ListExperimentsForClient", client).Return(experiments, nil)
	expMgrSuccess.On("ListVariablesForExperiments", experiments).Return(experimentVariables, nil)
	expMgrSuccess.On("ListVariablesForExperiments", []manager.Experiment{}).
		Return(map[string][]manager.Variable{}, nil)

	expMgrFailure1 := &mocks.StandardExperimentManager{}
	expMgrFailure1.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig, nil)
	expMgrFailure1.On("IsCacheEnabled").Return(true, nil)
	expMgrFailure1.On("ListClients").Return([]manager.Client{}, errors.New("List clients error"))

	expMgrFailure2 := &mocks.StandardExperimentManager{}
	expMgrFailure2.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig, nil)
	expMgrFailure2.On("IsCacheEnabled").Return(true, nil)
	expMgrFailure2.On("ListClients").Return(clients, nil)
	expMgrFailure2.On("ListVariablesForClient", client).
		Return([]manager.Variable{}, errors.New("List client vars error"))

	expMgrFailure3 := &mocks.StandardExperimentManager{}
	expMgrFailure3.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig, nil)
	expMgrFailure3.On("IsCacheEnabled").Return(true, nil)
	expMgrFailure3.On("ListClients").Return(clients, nil)
	expMgrFailure3.On("ListVariablesForClient", client).Return(clientVariables, nil)
	expMgrFailure3.On("ListExperimentsForClient", client).
		Return([]manager.Experiment{}, errors.New("List experiments error"))

	expMgrFailure4 := &mocks.StandardExperimentManager{}
	expMgrFailure4.On("GetEngineInfo", mock.Anything).Return(standardExperimentManagerConfig, nil)
	expMgrFailure4.On("IsCacheEnabled").Return(true, nil)
	expMgrFailure4.On("ListClients").Return(clients, nil)
	expMgrFailure4.On("ListVariablesForClient", client).Return(clientVariables, nil)
	expMgrFailure4.On("ListExperimentsForClient", client).Return(experiments, nil)
	expMgrFailure4.On("ListVariablesForExperiments", experiments).
		Return(map[string][]manager.Variable{}, errors.New("List experiment vars error"))

	expManagerName := "exp-engine-002"
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
						FieldSource: request.HeaderFieldSource,
					},
					{
						Name:        "var-2",
						Required:    false,
						FieldSource: request.HeaderFieldSource,
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
			badCacheKey:   fmt.Sprintf("engine:%s:clients:1:variables", expManagerName),
			err:           "Malformed variables info found in the cache for client 1",
		},
		"failure | bad experiment vars cache": {
			expMgr:        expMgrSuccess,
			clientID:      "1",
			experimentIDs: []string{"2"},
			badCacheKey:   fmt.Sprintf("engine:%s:experiments:2:variables", expManagerName),
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
					expManagerName: data.expMgr,
				},
				cache: cacheObj,
			}

			// Run and Validate
			response, err := svc.ListVariables(expManagerName, data.clientID, data.experimentIDs)
			// Sort items
			sort.SliceStable(response.Config, func(i, j int) bool {
				return response.Config[i].Name < response.Config[j].Name
			})
			assert.Equal(t, data.expected, response)
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				response, err := svc.ListVariables(expManagerName, data.clientID, data.experimentIDs)
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
	tests := map[string]struct {
		engine   string
		inputCfg json.RawMessage
		managers func(engine string, cfg json.RawMessage, err error) map[string]manager.ExperimentManager
		err      error
	}{
		"failure | std exp mgr": {
			engine: "test-engine",
			managers: func(engine string, cfg json.RawMessage, err error) map[string]manager.ExperimentManager {
				return map[string]manager.ExperimentManager{}
			},
			err: errors.New("Unknown experiment engine test-engine"),
		},
		"success | std exp mgr": {
			engine: "standard",
			managers: func(engine string, cfg json.RawMessage, err error) map[string]manager.ExperimentManager {
				stdExpMgr := &mocks.StandardExperimentManager{}
				stdExpMgr.
					On("ValidateExperimentConfig", cfg).
					Return(err)

				return map[string]manager.ExperimentManager{
					engine: stdExpMgr,
				}
			},
			inputCfg: json.RawMessage(`{
				"client": {},
				"experiments": [],
				"variables": {}
			}`),
		},
		"success | custom exp mgr": {
			engine: "custom",
			managers: func(engine string, cfg json.RawMessage, err error) map[string]manager.ExperimentManager {
				customExpMgr := &mocks.ExperimentManager{}
				customExpMgr.On("ValidateExperimentConfig", cfg).Return(err)

				return map[string]manager.ExperimentManager{
					engine: customExpMgr,
				}
			},
			inputCfg: json.RawMessage("[1, 2]"),
		},
		"failure | validation errors": {
			managers: func(engine string, cfg json.RawMessage, err error) map[string]manager.ExperimentManager {
				customExpMgr := &mocks.ExperimentManager{}
				customExpMgr.
					On("ValidateExperimentConfig", cfg).
					Return(err)

				return map[string]manager.ExperimentManager{
					engine: customExpMgr,
				}
			},
			inputCfg: json.RawMessage("[3, 4]"),
			err:      errors.New("validation error"),
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			expSvc := &experimentsService{
				experimentManagers: data.managers(data.engine, data.inputCfg, data.err),
			}

			err := expSvc.ValidateExperimentConfig(data.engine, data.inputCfg)
			if data.err == nil {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, data.err.Error())
			}
		})
	}
}

func TestGetExperimentRunnerConfig(t *testing.T) {
	expectedResult1 := json.RawMessage(`{"key": "value1"}`)
	expectedResult2 := json.RawMessage(`{"key": "value2"}`)
	// Create mock experiment managers
	stdExpMgr := &mocks.StandardExperimentManager{}
	stdExpMgr.On("GetEngineInfo").Return(standardExperimentManagerConfig)
	stdExpMgr.
		On("GetExperimentRunnerConfig", json.RawMessage(`{"client": {"id": "1"}}`)).
		Return(expectedResult1, nil)

	customExpMgr := &mocks.ExperimentManager{}
	customExpMgr.On("GetEngineInfo").Return(customExperimentManagerConfig)
	customExpMgr.On("GetExperimentRunnerConfig", json.RawMessage(`[1,2]`)).Return(expectedResult2, nil)

	// Create test experiment service
	expSvc := &experimentsService{
		experimentManagers: map[string]manager.ExperimentManager{
			"standard": stdExpMgr,
			"custom":   customExpMgr,
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
			engine:         "standard",
			inputCfg:       json.RawMessage(`{"client": {"id": "1"}}`),
			expectedResult: expectedResult1,
		},
		"success | custom exp mgr": {
			engine:         "custom",
			inputCfg:       json.RawMessage(`[1,2]`),
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
