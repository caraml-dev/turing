package experiment_test

import (
	"bou.ke/monkey"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gojek/turing/engines/experiment"
	"github.com/gojek/turing/engines/experiment/plugin"
	v1 "github.com/gojek/turing/engines/experiment/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

const serializationError = "unable to serialize struct"

type unserializableStruct struct{}

func (qty *unserializableStruct) MarshalJSON() ([]byte, error) {
	return []byte{}, errors.New(serializationError)
}

func Test_NewEngineFactory(t *testing.T) {
	var suite = map[string]struct {
		engine    string
		cfg       map[string]interface{}
		expected  experiment.EngineFactory
		withPatch func(expected experiment.EngineFactory, fn func())
		err       string
	}{
		"success | plugin": {
			engine: "plugin-engine",
			cfg: map[string]interface{}{
				"PluginBinary": "path/to/binary",
				"key_1":        "value_1",
			},
			expected: &plugin.EngineFactory{
				EngineConfig: json.RawMessage("{\"key_1\":\"value_1\"}"),
			},
			withPatch: func(expected experiment.EngineFactory, fn func()) {
				monkey.Patch(plugin.NewFactory,
					func(string, json.RawMessage, *zap.SugaredLogger) (*plugin.EngineFactory, error) {
						return expected.(*plugin.EngineFactory), nil
					},
				)
				defer monkey.Unpatch(plugin.NewFactory)
				fn()
			},
		},
		"success | compile-time": {
			engine: "compiled-engine",
			cfg: map[string]interface{}{
				"key_1": "value_1",
			},
			expected: &v1.EngineFactory{
				EngineName:   "compiled-engine",
				EngineConfig: json.RawMessage("{\"key_1\":\"value_1\"}"),
			},
		},
		"failure | invalid configuration": {
			cfg: map[string]interface{}{
				"PluginBinary": 123,
			},
			err: "1 error(s) decoding:\n\n* 'PluginBinary' expected type 'string', got unconvertible type 'int', value: '123'",
		},
		"failure | marshalling error": {
			cfg: map[string]interface{}{
				"unserializable_key": &unserializableStruct{},
			},
			err: fmt.Sprintf(
				"json: error calling MarshalJSON for type *experiment_test.unserializableStruct: %s",
				serializationError),
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()

			assertFunction := func() {
				actual, err := experiment.NewEngineFactory(tt.engine, tt.cfg, logger.Sugar())
				if tt.err != "" {
					assert.EqualError(t, err, tt.err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, actual)
				}
			}

			if tt.withPatch != nil {
				tt.withPatch(tt.expected, assertFunction)
			} else {
				assertFunction()
			}

		})
	}
}
