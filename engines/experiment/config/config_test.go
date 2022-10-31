package config_test

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/experiment/config"
)

func Test_EngineConfigDecoding(t *testing.T) {
	var suite = map[string]struct {
		cfg      interface{}
		expected config.EngineConfig
		err      string
	}{
		"success | plugin": {
			cfg: map[string]interface{}{
				"plugin_binary": "path/to/binary",
			},
			expected: config.EngineConfig{
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
			expected: config.EngineConfig{
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
			expected: config.EngineConfig{
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
			var actual config.EngineConfig
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

func TestEngineConfig_IsPlugin(t *testing.T) {
	var suite = map[string]struct {
		cfg      config.EngineConfig
		expected bool
	}{
		"success | binary path provided": {
			cfg: config.EngineConfig{
				PluginBinary: "/app/plugins/my_plugin",
			},
			expected: true,
		},
		"success | binary path not provided": {
			cfg:      config.EngineConfig{},
			expected: false,
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			actual := tt.cfg.IsPlugin()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
