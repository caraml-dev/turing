package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
