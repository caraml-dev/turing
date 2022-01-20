package experiment_test

import (
	"testing"

	"github.com/gojek/turing/engines/experiment"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func Test_EngineConfigDecoding(t *testing.T) {
	var suite = map[string]struct {
		cfg      interface{}
		expected experiment.EngineConfig
		err      string
	}{
		"success | plugin": {
			cfg: map[string]interface{}{
				"plugin_binary": "path/to/binary",
			},
			expected: experiment.EngineConfig{
				PluginBinary:        "path/to/binary",
				EngineConfiguration: nil,
			},
		},
		"success | plugin with extra config": {
			cfg: map[string]interface{}{
				"plugin_binary": "path/to/binary",
				"Key1":          "Value1",
				"Key2": map[string]interface{}{
					"Key2-1": "Value2-1",
				},
			},
			expected: experiment.EngineConfig{
				PluginBinary: "path/to/binary",
				EngineConfiguration: map[string]interface{}{
					"Key1": "Value1",
					"Key2": map[string]interface{}{
						"Key2-1": "Value2-1",
					},
				},
			},
		},
		"success | only engine config": {
			cfg: map[string]interface{}{
				"Key1": "Value1",
			},
			expected: experiment.EngineConfig{
				EngineConfiguration: map[string]interface{}{
					"Key1": "Value1",
				},
			},
		},
		"failure | invalid type": {
			cfg: map[string]interface{}{
				"plugin_binary": map[string]interface{}{
					"Key2-1": "Value2-1",
				},
			},
			err: "1 error(s) decoding:\n\n* 'plugin_binary' expected type 'string', " +
				"got unconvertible type 'map[string]interface {}', value: 'map[Key2-1:Value2-1]'",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			var actual experiment.EngineConfig
			err := mapstructure.Decode(tt.cfg, &actual)

			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}
