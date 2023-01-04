package runner

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/mocks"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	runnerMocks "github.com/caraml-dev/turing/engines/experiment/runner/mocks"
)

type mockExperimentRunner struct {
	runnerMocks.ExperimentRunner
}

func (_m *mockExperimentRunner) Configure(cfg json.RawMessage) error {
	ret := _m.Called(cfg)

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(error)
	}

	return r0
}

func TestRpcClient_GetTreatmentForRequest(t *testing.T) {
	suite := map[string]struct {
		expected *runner.Treatment
		err      string
	}{
		"success": {
			expected: &runner.Treatment{
				ExperimentName: "experiment-1",
				Name:           "my-treatment",
				Config:         json.RawMessage("{}"),
			},
		},
		"failure": {
			err: "failed to call experiment engine",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			req := GetTreatmentRequest{}

			mockClient := &mocks.RPCClient{}
			mockClient.
				On(
					"Call",
					"Plugin.GetTreatmentForRequest",
					&req,
					mock.AnythingOfType("*runner.Treatment")).
				Run(func(args mock.Arguments) {
					if tt.err == "" {
						resp := args.Get(2).(*runner.Treatment)
						*resp = *tt.expected
					}
				}).
				Return(func() error {
					if tt.err != "" {
						return errors.New(tt.err)
					}
					return nil
				})

			rpcClient := rpcClient{RPCClient: mockClient}
			actual, err := rpcClient.GetTreatmentForRequest(req.Header, req.Payload, req.Options)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
				assert.Nil(t, actual)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, actual, tt.expected)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRpcServer_GetTreatmentForRequest(t *testing.T) {
	suite := map[string]struct {
		expected *runner.Treatment
		err      error
	}{
		"success | fetch treatment": {
			expected: &runner.Treatment{
				ExperimentName: "experiment-1",
				Name:           "my-treatment",
				Config:         json.RawMessage("{}"),
			},
		},
		"failure | fetch treatment": {
			err: errors.New("unknown"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := &mockExperimentRunner{}
			mockManager.On(
				"GetTreatmentForRequest",
				mock.Anything, mock.Anything, mock.Anything,
			).Return(tt.expected, tt.err)
			rpcServer := &rpcServer{nil, mockManager}

			var actual runner.Treatment
			err := rpcServer.GetTreatmentForRequest(&GetTreatmentRequest{}, &actual)

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, &actual)
			}
			mockManager.AssertExpectations(t)
		})
	}
}

func TestRpcCollectorClient_MeasureDurationMsSince(t *testing.T) {
	suite := map[string]struct {
		err string
	}{
		"success": {
			err: "",
		},
		"failure": {
			err: "failed to log metric",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			req := MeasureDurationMsSinceRequest{}

			mockClient := &mocks.RPCClient{}
			mockClient.
				On(
					"Call",
					"Plugin.MeasureDurationMsSince",
					&req,
					mock.Anything).
				Run(func(args mock.Arguments) {
					resp := args.Get(2).(*interface{})
					*resp = nil
				}).
				Return(func() error {
					if tt.err != "" {
						return errors.New(tt.err)
					}
					return nil
				})

			rpcCollectorClient := rpcCollectorClient{RPCClient: mockClient}
			err := rpcCollectorClient.MeasureDurationMsSince(req.Key, req.Starttime, req.Labels)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRpcCollectorClient_MeasureDurationMs(t *testing.T) {
	suite := map[string]struct {
		err      string
		expected func()
	}{
		"success": {
			expected: func() {},
		},
		"failure": {
			err: "some error",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			req := MeasureDurationMsRequest{}

			mockClient := &mocks.RPCClient{}
			mockClient.
				On(
					"Call",
					"Plugin.MeasureDurationMs",
					&req,
					mock.AnythingOfType("*func()")).
				Run(func(args mock.Arguments) {
					if tt.err == "" {
						resp := args.Get(2).(*func())
						*resp = tt.expected
					}
				}).
				Return(func() error {
					if tt.err != "" {
						return errors.New(tt.err)
					}
					return nil
				})

			rpcCollectorClient := rpcCollectorClient{RPCClient: mockClient}
			actual := rpcCollectorClient.MeasureDurationMs(req.Key, req.Labels)
			if tt.expected != nil {
				assert.Equal(t, reflect.ValueOf(tt.expected), reflect.ValueOf(actual))
			} else {
				assert.Nil(t, actual)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRpcCollectorClient_RecordGauge(t *testing.T) {
	suite := map[string]struct {
		err string
	}{
		"success": {
			err: "",
		},
		"failure": {
			err: "failed to log metric",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			req := RecordGaugeRequest{}

			mockClient := &mocks.RPCClient{}
			mockClient.
				On(
					"Call",
					"Plugin.RecordGauge",
					&req,
					mock.Anything).
				Run(func(args mock.Arguments) {
					resp := args.Get(2).(*interface{})
					*resp = nil
				}).
				Return(func() error {
					if tt.err != "" {
						return errors.New(tt.err)
					}
					return nil
				})

			rpcCollectorClient := rpcCollectorClient{RPCClient: mockClient}
			err := rpcCollectorClient.RecordGauge(req.Key, req.Value, req.Labels)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRpcCollectorClient_Inc(t *testing.T) {
	suite := map[string]struct {
		err string
	}{
		"success": {
			err: "",
		},
		"failure": {
			err: "failed to log metric",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			req := IncRequest{}

			mockClient := &mocks.RPCClient{}
			mockClient.
				On(
					"Call",
					"Plugin.Inc",
					&req,
					mock.Anything).
				Run(func(args mock.Arguments) {
					resp := args.Get(2).(*interface{})
					*resp = nil
				}).
				Return(func() error {
					if tt.err != "" {
						return errors.New(tt.err)
					}
					return nil
				})

			rpcCollectorClient := rpcCollectorClient{RPCClient: mockClient}
			err := rpcCollectorClient.Inc(req.Key, req.Labels)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRpcCollectorServer_MeasureDurationMsSince(t *testing.T) {
	suite := map[string]struct {
		err string
	}{
		"success": {
			err: "",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			rpcCollectorServer := &rpcCollectorServer{}
			err := rpcCollectorServer.MeasureDurationMsSince(&MeasureDurationMsSinceRequest{}, nil)

			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRpcCollectorServer_MeasureDurationMs(t *testing.T) {
	suite := map[string]struct {
		expected func()
		err      string
	}{
		"success": {
			err: "",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			rpcCollectorServer := &rpcCollectorServer{}

			var actual func()
			err := rpcCollectorServer.MeasureDurationMs(&MeasureDurationMsRequest{}, &actual)

			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
				// assert that the returned value is not nil since Go does not support function comparisons
				assert.NotNil(t, actual)
			}
		})
	}
}

func TestRpcCollectorServer_RecordGauge(t *testing.T) {
	suite := map[string]struct {
		err string
	}{
		"success": {
			err: "",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			rpcCollectorServer := &rpcCollectorServer{}
			err := rpcCollectorServer.RecordGauge(&RecordGaugeRequest{}, nil)

			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRpcCollectorServer_Inc(t *testing.T) {
	suite := map[string]struct {
		err string
	}{
		"success": {
			err: "",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			rpcCollectorServer := &rpcCollectorServer{}
			err := rpcCollectorServer.Inc(&IncRequest{}, nil)

			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type mockMetricsRegistrationHelper struct {
	runnerMocks.MetricsRegistrationHelper
}

func TestRpcMetricsRegistrationHelperClient_Register(t *testing.T) {
	suite := map[string]struct {
		err string
	}{
		"success": {
			err: "",
		},
		"failure": {
			err: "failed to register metrics",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			req := make([]instrumentation.Metric, 0)

			mockClient := &mocks.RPCClient{}
			mockClient.
				On(
					"Call",
					"Plugin.Register",
					req,
					mock.Anything).
				Run(func(args mock.Arguments) {
					resp := args.Get(2).(*interface{})
					*resp = nil
				}).
				Return(func() error {
					if tt.err != "" {
						return errors.New(tt.err)
					}
					return nil
				})

			rpcMetricsRegistrationHelperClient := rpcMetricsRegistrationHelperClient{RPCClient: mockClient}
			err := rpcMetricsRegistrationHelperClient.Register(req)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRpcMetricsRegistrationHelperServer_Register(t *testing.T) {
	suite := map[string]struct {
		req []instrumentation.Metric
		err error
	}{
		"success": {
			req: make([]instrumentation.Metric, 0),
		},
		"failure": {
			err: errors.New("unknown"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockManager := &mockMetricsRegistrationHelper{}
			mockManager.On(
				"Register",
				mock.Anything,
			).Return(tt.err)
			rpcMetricsRegistrationHelperServer := &rpcMetricsRegistrationHelperServer{mockManager}
			err := rpcMetricsRegistrationHelperServer.Register(tt.req, nil)

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
			mockManager.AssertExpectations(t)
		})
	}
}
