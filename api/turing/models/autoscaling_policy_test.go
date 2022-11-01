package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/api/turing/models"
)

func TestAutoscalingPolicyValue(t *testing.T) {
	autoscalingPolicy := models.AutoscalingPolicy{
		Metric: models.AutoscalingMetricRPS,
		Target: "100",
	}

	// Validate
	value, err := autoscalingPolicy.Value()
	// Convert to string for comparison
	byteValue, ok := value.([]byte)
	assert.True(t, ok)
	// Validate
	assert.NoError(t, err)
	assert.JSONEq(t, `
		{
			"metric": "rps",
			"target": "100"
		}
	`, string(byteValue))
}

func TestAutoscalingPolicyScan(t *testing.T) {
	tests := map[string]struct {
		value    interface{}
		success  bool
		expected models.AutoscalingPolicy
		err      string
	}{
		"success": {
			value: []byte(`{
				"metric": "cpu",
				"target": "90"
			}`),
			success: true,
			expected: models.AutoscalingPolicy{
				Metric: models.AutoscalingMetricCPU,
				Target: "90",
			},
		},
		"failure | invalid value": {
			value:   100,
			success: false,
			err:     "type assertion to []byte failed",
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			var autoscalingPolicy models.AutoscalingPolicy
			err := autoscalingPolicy.Scan(data.value)
			if data.success {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, autoscalingPolicy)
			} else {
				assert.Error(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}
