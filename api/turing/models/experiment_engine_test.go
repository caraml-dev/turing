package models

import (
	"testing"

	"github.com/gojek/turing/engines/experiment"
	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/stretchr/testify/assert"
)

func TestExperimentEngineValue(t *testing.T) {
	tests := map[string]struct {
		expEngine ExperimentEngine
		expected  string
	}{
		"nop": {
			expEngine: ExperimentEngine{
				Type: ExperimentEngineTypeNop,
			},
			expected: string(`{
				"type": "nop"
			}`),
		},
		"litmus": {
			expEngine: ExperimentEngine{
				Type: ExperimentEngineTypeLitmus,
				Config: &manager.TuringExperimentConfig{
					Client: manager.Client{
						ID:       "1",
						Username: "c1",
						Passkey:  "encrypted",
					},
					Experiments: []manager.Experiment{},
					Variables: manager.Variables{
						ClientVariables: []manager.Variable{
							{
								Name:     "var-1",
								Required: true,
								Type:     manager.UnsupportedVariableType,
							},
						},
						ExperimentVariables: map[string][]manager.Variable{},
						Config: []manager.VariableConfig{
							{
								Name:        "var-1",
								Required:    true,
								Field:       "field1",
								FieldSource: experiment.HeaderFieldSource,
							},
						},
					},
				},
			},
			expected: string(`{
				"type": "litmus",
				"config": {
					"client": {
						"id": "1",
						"username": "c1",
						"passkey": "encrypted"
					},
					"experiments": [],
					"variables": {
						"client_variables": [
							{
								"name": "var-1",
								"required": true,
								"type": "unsupported"
							}
						],
						"experiment_variables": {},
						"config": [
							{
								"name": "var-1",
								"required": true,
								"field": "field1",
								"field_source": "header"
							}
						]
					}
				}
			}`),
		},
		"xp": {
			expEngine: ExperimentEngine{
				Type:   ExperimentEngineTypeXp,
				Config: []int{1, 2},
			},
			expected: string(`{
				"type": "xp",
				"config": [1,2]
			}`),
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			value, err := data.expEngine.Value()
			// Convert to string for comparison
			byteValue, ok := value.([]byte)
			assert.True(t, ok)
			// Validate
			assert.NoError(t, err)
			assert.JSONEq(t, data.expected, string(byteValue))
		})
	}
}

func TestExperimentEngineScan(t *testing.T) {
	tests := map[string]struct {
		value    interface{}
		success  bool
		expected ExperimentEngine
		err      string
	}{
		"success": {
			value:   []byte(`{"type": "nop"}`),
			success: true,
			expected: ExperimentEngine{
				Type: ExperimentEngineTypeNop,
			},
		},
		"failure | invalid bytes": {
			value:   []byte("test-string"),
			success: false,
			err:     "invalid character 'e' in literal true (expecting 'r')",
		},
		"failure | invalid value": {
			value:   100,
			success: false,
			err:     "type assertion to []byte failed",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			var expEngine ExperimentEngine
			err := expEngine.Scan(data.value)
			if data.success {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, expEngine)
			} else {
				assert.Error(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}
