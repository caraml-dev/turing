package runner

import (
	"encoding/json"
	"errors"
	"testing"

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
			rpcServer := &rpcServer{mockManager}

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
