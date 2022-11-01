package hardcoded

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/engines/experiment/runner"
)

func TestExperimentRunner_GetTreatmentForRequest(t *testing.T) {
	runnerConfig := json.RawMessage(`{
		"experiments": [
			{
				"id": "001",
				"name": "exp_1",
				"variants": [
					{"name": "control"}, 
					{"name": "treatment-1"}
				],
				"segmentation_configuration": {
					"name": "client_id",
					"source": "payload",
					"value": "client.id"
				},
				"variants_configuration": {
					"control": {
						"traffic": 0.85,
						"treatment_configuration": {"foo": "bar"}
					},
					"treatment-1": {
						"traffic": 0.15,
						"treatment_configuration": {"bar": "baz"}
					}
				}
			}
		]
	}`)

	suite := map[string]struct {
		runnerConfig json.RawMessage
		header       http.Header
		payload      json.RawMessage
		expected     *runner.Treatment
		err          string
	}{
		"success | client_id:4": {
			runnerConfig: runnerConfig,
			payload:      json.RawMessage(`{"client": {"id": 4}}`),
			expected: &runner.Treatment{
				ExperimentName: "exp_1",
				Name:           "control",
				Config:         json.RawMessage(`{"foo": "bar"}`),
			},
		},
		"success | client_id:7": {
			runnerConfig: runnerConfig,
			payload:      json.RawMessage(`{"client": {"id": 7}}`),
			expected: &runner.Treatment{
				ExperimentName: "exp_1",
				Name:           "treatment-1",
				Config:         json.RawMessage(`{"bar": "baz"}`),
			},
		},
		"failure": {
			runnerConfig: runnerConfig,
			payload:      json.RawMessage(`{}`),
			err:          "no experiment configured for the unit",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			expRunner := &ExperimentRunner{}
			err := expRunner.Configure(tt.runnerConfig)
			require.NoError(t, err)

			actual, err := expRunner.GetTreatmentForRequest(tt.header, tt.payload, runner.GetTreatmentOptions{})
			if tt.err != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}
