package config

import "encoding/json"

// EngineConfig is a struct used to decode engine's configuration into
// It consists of an optional PluginBinary (if the experiment engine is implemented
// as net/rpc plugin) and unstructured EngineConfiguration of key/value data, that is
// used to configure experiment manager/runner
type EngineConfig struct {
	PluginBinary        string                 `mapstructure:"plugin_binary"`
	EngineConfiguration map[string]interface{} `mapstructure:",remain"`
}

func (c EngineConfig) IsPlugin() bool {
	return c.PluginBinary != ""
}

func (c EngineConfig) RawEngineConfig() (json.RawMessage, error) {
	return json.Marshal(c.EngineConfiguration)
}
