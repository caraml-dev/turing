package experiment

import (
	"github.com/gojek/turing/engines/experiment/plugin"
)

type EngineConfig struct {
	Plugin              plugin.Configuration   `mapstructure:",squash"`
	EngineConfiguration map[string]interface{} `mapstructure:",remain"`
}
