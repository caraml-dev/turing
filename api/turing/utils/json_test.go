package utils_test

import (
	"encoding/json"
	"github.com/gojek/turing/api/turing/utils"
	"github.com/stretchr/testify/assert"
	"testing"
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
