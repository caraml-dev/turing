package runner_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc"
	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc/mocks"
	rpcRunner "github.com/caraml-dev/turing/engines/experiment/plugin/rpc/runner"
	"github.com/caraml-dev/turing/engines/experiment/runner"
)

func configuredRunnerMock() *mocks.ConfigurableExperimentRunner {
	mockRunner := &mocks.ConfigurableExperimentRunner{}
	mockRunner.On("Configure", json.RawMessage(nil)).Return(nil)

	return mockRunner
}

func withExperimentRunner(
	t *testing.T,
	impl rpcRunner.ConfigurableExperimentRunner,
	config json.RawMessage,
	testFn func(expRunner runner.ExperimentRunner, err error),
) {
	plugins := map[string]plugin.Plugin{
		rpc.RunnerPluginIdentifier: &rpcRunner.ExperimentRunnerPlugin{
			Impl: impl,
		},
	}

	client, _ := plugin.TestPluginRPCConn(t, plugins, nil)

	factory := &rpc.EngineFactory{
		Client:       client,
		EngineConfig: config,
	}

	testFn(factory.GetExperimentRunner())
}

func TestExperimentRunnerPlugin_Configure(t *testing.T) {
	suite := map[string]struct {
		cfg json.RawMessage
		err error
	}{
		"success": {
			cfg: json.RawMessage(`{"config": [1,2]}`),
		},
		"failure | configuration error": {
			err: errors.New("failed to configure plugin"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockRunner := &mocks.ConfigurableExperimentRunner{}
			mockRunner.On("Configure", tt.cfg).Return(tt.err)

			withExperimentRunner(t, mockRunner, tt.cfg,
				func(expRunner runner.ExperimentRunner, err error) {
					if tt.err != nil {
						assert.Nil(t, expRunner)
						assert.EqualError(t, err,
							fmt.Sprintf(
								`failed to configure "%s" plugin instance: %v`,
								rpc.RunnerPluginIdentifier, tt.err))
					} else {
						assert.NoError(t, err)
						assert.NotNil(t, expRunner)
					}
				})

			mockRunner.AssertExpectations(t)
		})
	}
}

func TestExperimentRunnerPlugin_GetTreatmentForRequest(t *testing.T) {
	suite := map[string]struct {
		header   http.Header
		payload  []byte
		options  runner.GetTreatmentOptions
		expected *runner.Treatment
		err      error
	}{
		"success": {
			header: http.Header{
				"x-country-code": []string{"id"},
			},
			payload: json.RawMessage(`{"customer_id": 1}`),
			options: runner.GetTreatmentOptions{
				TuringRequestID: "req-1",
			},
			expected: &runner.Treatment{
				ExperimentName: "experiment-1",
				Name:           "treatment-0",
				Config:         json.RawMessage(`{"key-1": "val-1"}`),
			},
		},
		"failure": {
			err: errors.New("no experiments configured for given request"),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			mockRunner := configuredRunnerMock()
			mockRunner.
				On("GetTreatmentForRequest", tt.header, tt.payload, tt.options).
				Return(tt.expected, tt.err)

			withExperimentRunner(t, mockRunner, nil,
				func(expRunner runner.ExperimentRunner, _ error) {
					actual, err := expRunner.GetTreatmentForRequest(tt.header, tt.payload, tt.options)

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
