package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/api/turing/utils"
)

func TestMergeJSON(t *testing.T) {
	tests := map[string]struct {
		original  json.RawMessage
		overrides map[string]interface{}
		expected  json.RawMessage
		err       string
	}{
		"success": {
			original: json.RawMessage(`{"key-1": "value-1"}`),
			overrides: map[string]interface{}{
				"key-2": "value-2",
			},
			expected: json.RawMessage(`{"key-1": "value-1", "key-2": "value-2"}`),
		},
		"success | override": {
			original: json.RawMessage(`{"key-1": "value-1"}`),
			overrides: map[string]interface{}{
				"key-1": "value-3",
				"key-2": "value-2",
			},
			expected: json.RawMessage(`{"key-1": "value-3", "key-2": "value-2"}`),
		},
		"success | nested": {
			original: json.RawMessage(`{"key-1": "value-1"}`),
			overrides: map[string]interface{}{
				"key-1": json.RawMessage(`{"key-1-1": "value-1", "key-1-2": "value-2"}`),
			},
			expected: json.RawMessage(`{"key-1": {"key-1-1": "value-1", "key-1-2": "value-2"}}`),
		},
		"success | original json empty": {
			original: json.RawMessage{},
			overrides: map[string]interface{}{
				"key-1": 1,
			},
			expected: json.RawMessage(`{"key-1": 1}`),
		},
		"success | overrides empty": {
			original: json.RawMessage(`{"key-1": "value-1"}`),
			expected: json.RawMessage(`{"key-1": "value-1"}`),
		},
		"failure": {
			original: json.RawMessage(`1`),
			err:      "json: cannot unmarshal number into Go value of type map[string]interface {}",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := utils.MergeJSON(tt.original, tt.overrides)
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.JSONEq(t, string(tt.expected), string(actual))
			}
		})
	}
}
