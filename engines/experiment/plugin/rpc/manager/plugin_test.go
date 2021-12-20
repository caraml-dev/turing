package manager_test

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/gojek/turing/engines/experiment/plugin/rpc"
	rpcManager "github.com/gojek/turing/engines/experiment/plugin/rpc/manager"
	"github.com/gojek/turing/engines/experiment/plugin/rpc/mocks"
	"github.com/hashicorp/go-plugin"
)

func configuredMockManager() *mocks.ConfigurableExperimentManager {
	mockManager := &mocks.ConfigurableExperimentManager{}
	mockManager.On("Configure", json.RawMessage(nil)).Return(nil)

	return mockManager
}

func withExperimentManager(
	t *testing.T,
	mock *mocks.ConfigurableExperimentManager,
	testFn func(em manager.ExperimentManager, err error),
) {
	plugins := map[string]plugin.Plugin{
		rpc.ManagerPluginIdentifier: &rpcManager.ExperimentManagerPlugin{
			Impl: mock,
		},
	}

	client, _ := plugin.TestPluginRPCConn(t, plugins, nil)

	factory := &rpc.EngineFactory{
		Client:       client,
		EngineConfig: nil,
	}

	testFn(factory.GetExperimentManager())

	mock.AssertExpectations(t)
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
		})
	}
}

func TestExperimentManagerPlugin_GetEngineInfo(t *testing.T) {
	suite := map[string]struct {
		expected manager.Engine
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
			mockManager := configuredMockManager()
			mockManager.On("GetEngineInfo").Return(tt.expected)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				actual := em.GetEngineInfo()
				assert.Equal(t, tt.expected, actual)
			})
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
			mockManager := configuredMockManager()
			mockManager.On("ValidateExperimentConfig", tt.experimentConfig).Return(tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				err := em.ValidateExperimentConfig(tt.experimentConfig)
				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				} else {
					assert.NoError(t, err)
				}
			})
		})
	}
}

func TestExperimentManagerPlugin_GetExperimentRunnerConfig(t *testing.T) {
	suite := map[string]struct {
		experimentConfig interface{}
		expected         json.RawMessage
		err              error
	}{
		"success | standard experiment config": {
			experimentConfig: &manager.TuringExperimentConfig{
				Client: manager.Client{
					ID: "client-id",
				},
			},
			expected: json.RawMessage(`{}`),
		},
		"success | custom experiment config": {
			experimentConfig: json.RawMessage(`{
				"client": {
					"id": "client-id"
				}
			`),
			expected: json.RawMessage(`{}`),
		},
		"failure": {
			experimentConfig: "unexpected config",
			err:              errors.New("failed to retrieve runner's config"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := configuredMockManager()
			mockManager.On("GetExperimentRunnerConfig", tt.experimentConfig).Return(tt.expected, tt.err)

			withExperimentManager(t, mockManager, func(em manager.ExperimentManager, _ error) {
				gob.Register(tt.experimentConfig)

				actual, err := em.GetExperimentRunnerConfig(tt.experimentConfig)
				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, actual)
				}
			})
		})
	}
}
