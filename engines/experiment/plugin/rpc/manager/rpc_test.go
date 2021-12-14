package manager

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/plugin/rpc/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRpcClient_Configure(t *testing.T) {
	suite := map[string]struct {
		cfg json.RawMessage
		err string
	}{
		"success": {
			cfg: json.RawMessage("{\"my_config\": \"my_value\"}"),
		},
		"failure": {
			cfg: json.RawMessage("{\"my_config\": \"my_value\"}"),
			err: "failed to configure experiment manager plugin",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockClient := &mocks.RPCClient{}
			mockClient.On("Call", "Plugin.Configure", tt.cfg, mock.Anything).Return(
				func() error {
					if tt.err != "" {
						return errors.New(tt.err)
					}
					return nil
				})

			rpcClient := rpcClient{RPCClient: mockClient}

			err := rpcClient.Configure(tt.cfg)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRpcClient_GetEngineInfo(t *testing.T) {
	suite := map[string]struct {
		expected manager.Engine
		err      string
	}{
		"success | get engine info": {
			expected: manager.Engine{
				Name: "engine-1",
				Type: manager.StandardExperimentManagerType,
			},
		},
		"failure | get engine info": {
			err: "something's wrong",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockClient := &mocks.RPCClient{}
			mockClient.
				On("Call", "Plugin.GetEngineInfo", mock.Anything, mock.AnythingOfType("*manager.Engine")).
				Run(func(args mock.Arguments) {
					resp := args.Get(2).(*manager.Engine)
					*resp = tt.expected
				}).
				Return(func() error {
					if tt.err != "" {
						return errors.New(tt.err)
					}
					return nil
				})

			rpcClient := rpcClient{RPCClient: mockClient}

			actual := rpcClient.GetEngineInfo()
			if tt.err != "" {
				empty := manager.Engine{}
				assert.Equal(t, actual, empty)
			} else {
				assert.Equal(t, actual, tt.expected)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestRpcServer_Configure(t *testing.T) {
	suite := map[string]struct {
		cfg json.RawMessage
		err string
	}{
		"success | configure manager": {
			cfg: json.RawMessage("{\"my_config\": \"my_value\"}"),
		},
		"failure | configuration failed": {
			cfg: json.RawMessage("{\"my_config\": \"my_value\"}"),
			err: "failed to initialize experiment manager",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := &mocks.ConfigurableExperimentManager{}
			mockManager.On("Configure", tt.cfg).Return(func() error {
				if tt.err != "" {
					return errors.New(tt.err)
				}
				return nil
			})
			rpcServer := &rpcServer{mockManager}

			err := rpcServer.Configure(tt.cfg, nil)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
			mockManager.AssertExpectations(t)
		})
	}
}

func TestRpcServer_GetEngineInfo(t *testing.T) {
	suite := map[string]struct {
		expected manager.Engine
		err      string
	}{
		"success | get engine info": {
			expected: manager.Engine{
				Name: "engine-1",
				Type: manager.StandardExperimentManagerType,
			},
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := &mocks.ConfigurableExperimentManager{}
			mockManager.On("GetEngineInfo", mock.Anything).
				Return(tt.expected)
			rpcServer := &rpcServer{mockManager}

			var actual manager.Engine
			err := rpcServer.GetEngineInfo(nil, &actual)

			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
			mockManager.AssertExpectations(t)
		})
	}
}
