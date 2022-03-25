package hardcoded

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/pkg/request"
	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExperimentRunner_GetTreatmentForRequest(t *testing.T) {
	suite := map[string]struct {
		experiments []Experiment
		header      http.Header
		payload     json.RawMessage
		expected    *runner.Treatment
		err         string
	}{
		"success": {
			experiments: []Experiment{
				{
					Experiment: manager.Experiment{
						ID:   "001",
						Name: "exp_1",
						Variants: []manager.Variant{
							{
								Name: "control",
							},
							{
								Name: "treatment-1",
							},
						},
					},
					SegmentationConfig: SegmenterConfig{
						Name:            "customer_id",
						SegmenterSource: request.PayloadFieldSource,
						SegmenterValue:  "customer.id",
					},
					VariantsConfig: map[string]TreatmentConfig{
						"control": {
							Traffic: 0.85,
							Data:    json.RawMessage(`{"foo": "bar"}`),
						},
						"treatment-1": {
							Traffic: 0.15,
							Data:    json.RawMessage(`{"bar": "baz"}`),
						},
					},
				},
			},
			payload: json.RawMessage(`{"customer": {"id": 7}}`),
			expected: &runner.Treatment{
				ExperimentName: "exp_1",
				Name:           "treatment-1",
				Config:         json.RawMessage(`{"bar": "baz"}`),
			},
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			expRunner := ExperimentRunner{experiments: tt.experiments}

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
