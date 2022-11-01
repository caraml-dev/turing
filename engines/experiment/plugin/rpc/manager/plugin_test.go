package manager_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/experiment/manager"

	"github.com/hashicorp/go-plugin"

	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc"
	rpcManager "github.com/caraml-dev/turing/engines/experiment/plugin/rpc/manager"
	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/mocks"
)

func configuredManagerMock() *mocks.ConfigurableExperimentManager {
	mockManager := &mocks.ConfigurableExperimentManager{}
	mockManager.On("Configure", json.RawMessage(nil)).Return(nil)

	return mockManager
}

func configuredStandardManagerMock() *mocks.ConfigurableStandardExperimentManager {
	mockManager := &mocks.ConfigurableStandardExperimentManager{}
	mockManager.On("Configure", json.RawMessage(nil)).Return(nil)

	return mockManager
}

func withExperimentManager(
	t *testing.T,
	impl rpcManager.ConfigurableExperimentManager,
	testFn func(em manager.ExperimentManager, err error),
) {
	plugins := map[string]plugin.Plugin{
		rpc.ManagerPluginIdentifier: &rpcManager.ExperimentManagerPlugin{
			Impl: impl,
		},
	}

	client, _ := plugin.TestPluginRPCConn(t, plugins, nil)

	factory := &rpc.EngineFactory{
		Client:       client,
		EngineConfig: nil,
	}

	testFn(factory.GetExperimentManager())
}

func TestExperimentManagerPlugin_Configure(t *testing.T) {
	suite := map[string]struct {
		err error
	}{
		"success": {},
		"failure | configuration error": {
			err: errors.New("failed to configure plugin"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := &mocks.ConfigurableExperimentManager{}
			mockManager.On("Configure", json.RawMessage(nil)).Return(tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, err error) {
				if tt.err != nil {
					assert.Nil(t, em)
					assert.EqualError(t, err,
						fmt.Sprintf("failed to configure \"experiment_manager\" plugin instance: %v", tt.err))
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, em)
				}
			})

			mockManager.AssertExpectations(t)
		})
	}
}

func TestExperimentManagerPlugin_GetEngineInfo(t *testing.T) {
	suite := map[string]struct {
		expected manager.Engine
		err      error
	}{
		"success": {
			expected: manager.Engine{
				Name:        "standard",
				DisplayName: "Standard Manager",
				Type:        manager.StandardExperimentManagerType,
				StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
					ClientSelectionEnabled:     true,
					ExperimentSelectionEnabled: true,
					HomePageURL:                "http://example.com",
				},
			},
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := configuredManagerMock()
			mockManager.On("GetEngineInfo").Return(tt.expected, tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				actual, err := em.GetEngineInfo()
				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, actual)
				}
			})

			mockManager.AssertExpectations(t)
		})
	}
}

func TestExperimentManagerPlugin_ValidateExperimentConfig(t *testing.T) {
	suite := map[string]struct {
		experimentConfig json.RawMessage
		err              error
	}{
		"success | validation passed": {
			experimentConfig: json.RawMessage(`{
				"my_config": "my_value"
			`),
		},
		"failure | validation failed": {
			experimentConfig: json.RawMessage(`[1, 2]`),
			err:              errors.New("validation failed"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := configuredManagerMock()
			mockManager.On("ValidateExperimentConfig", tt.experimentConfig).Return(tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				err := em.ValidateExperimentConfig(tt.experimentConfig)
				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				} else {
					assert.NoError(t, err)
				}
			})

			mockManager.AssertExpectations(t)
		})
	}
}

func TestExperimentManagerPlugin_GetExperimentRunnerConfig(t *testing.T) {
	suite := map[string]struct {
		experimentConfig json.RawMessage
		expected         json.RawMessage
		err              error
	}{
		"success | custom experiment config": {
			experimentConfig: json.RawMessage(`{
				"client": {
					"id": "client-id"
				}
			`),
			expected: json.RawMessage(`{}`),
		},
		"failure": {
			experimentConfig: json.RawMessage("unexpected config"),
			err:              errors.New("failed to retrieve runner's config"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := configuredManagerMock()
			mockManager.On("GetExperimentRunnerConfig", tt.experimentConfig).Return(tt.expected, tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				actual, err := em.GetExperimentRunnerConfig(tt.experimentConfig)
				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, actual)
				}
			})

			mockManager.AssertExpectations(t)
		})
	}
}

func TestExperimentManagerPlugin_IsCacheEnabled(t *testing.T) {
	suite := map[string]struct {
		cacheEnabled bool
		err          error
	}{
		"success | enabled": {
			cacheEnabled: true,
		},
		"failure | error": {
			err: errors.New("failure"),
		},
		"success | disabled": {
			cacheEnabled: false,
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := configuredStandardManagerMock()
			mockManager.On("IsCacheEnabled").Return(tt.cacheEnabled, tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				actual, err := em.(manager.StandardExperimentManager).IsCacheEnabled()
				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
					assert.False(t, actual)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.cacheEnabled, actual)
				}
			})

			mockManager.AssertExpectations(t)
		})
	}
}

func TestExperimentManagerPlugin_ListClients(t *testing.T) {
	suite := map[string]struct {
		expected []manager.Client
		err      error
	}{
		"success": {
			expected: []manager.Client{
				{
					ID:       "client-1",
					Username: "username-1",
				},
			},
		},
		"failure": {
			err: errors.New("failed to fetch clients"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := configuredStandardManagerMock()
			mockManager.On("ListClients").Return(tt.expected, tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				actual, err := em.(manager.StandardExperimentManager).ListClients()

				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, actual)
				}
			})

			mockManager.AssertExpectations(t)
		})
	}
}

func TestExperimentManagerPlugin_ListExperiments(t *testing.T) {
	suite := map[string]struct {
		expected []manager.Experiment
		err      error
	}{
		"success": {
			expected: []manager.Experiment{
				{
					ID:       "123-456-789",
					Name:     "experiment-01",
					ClientID: "client-01",
				},
			},
		},
		"failure": {
			err: errors.New("failed to fetch experiments"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := configuredStandardManagerMock()
			mockManager.On("ListExperiments").Return(tt.expected, tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				actual, err := em.(manager.StandardExperimentManager).ListExperiments()

				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, actual)
				}
			})

			mockManager.AssertExpectations(t)
		})
	}
}

func TestExperimentManagerPlugin_ListExperimentsForClient(t *testing.T) {
	suite := map[string]struct {
		client   manager.Client
		expected []manager.Experiment
		err      error
	}{
		"success": {
			client: manager.Client{
				ID: "client-02",
			},
			expected: []manager.Experiment{
				{
					Name:     "experiment-02",
					ClientID: "client-02",
				},
			},
		},
		"failure": {
			err: errors.New("failed to fetch experiments for client"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := configuredStandardManagerMock()
			mockManager.On("ListExperimentsForClient", tt.client).Return(tt.expected, tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				actual, err := em.(manager.StandardExperimentManager).ListExperimentsForClient(tt.client)

				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, actual)
				}
			})

			mockManager.AssertExpectations(t)
		})
	}
}

func TestExperimentManagerPlugin_ListVariablesForClient(t *testing.T) {
	suite := map[string]struct {
		client   manager.Client
		expected []manager.Variable
		err      error
	}{
		"success": {
			client: manager.Client{
				ID: "client-03",
			},
			expected: []manager.Variable{
				{
					Name:     "sessionId",
					Required: false,
					Type:     manager.FilterVariableType,
				},
				{
					Name:     "customerType",
					Required: true,
					Type:     manager.UnitVariableType,
				},
			},
		},
		"failure": {
			client: manager.Client{
				ID: "unknown",
			},
			err: errors.New("failed to fetch variables for client"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := configuredStandardManagerMock()
			mockManager.On("ListVariablesForClient", tt.client).Return(tt.expected, tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				actual, err := em.(manager.StandardExperimentManager).ListVariablesForClient(tt.client)

				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, actual)
				}
			})

			mockManager.AssertExpectations(t)
		})
	}
}

func TestExperimentManagerPlugin_ListVariablesForExperiments(t *testing.T) {
	suite := map[string]struct {
		experiments []manager.Experiment
		expected    map[string][]manager.Variable
		err         error
	}{
		"success": {
			experiments: []manager.Experiment{
				{
					Name: "experiment-04",
				},
			},
			expected: map[string][]manager.Variable{
				"experiment-04": {
					{
						Name:     "customerType",
						Required: true,
						Type:     manager.UnitVariableType,
					},
				},
			},
		},
		"failure": {
			experiments: []manager.Experiment(nil),
			err:         errors.New("failed to fetch variables for client"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := configuredStandardManagerMock()
			mockManager.On("ListVariablesForExperiments", tt.experiments).Return(tt.expected, tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				actual, err := em.(manager.StandardExperimentManager).ListVariablesForExperiments(tt.experiments)

				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, actual)
				}
			})

			mockManager.AssertExpectations(t)
		})
	}
}
