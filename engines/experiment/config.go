package experiment

// EngineConfig is a struct used to decode engine's configuration into
// It consists of an optional PluginBinary (if the experiment engine is implemented
// as net/rpc plugin) and unstructured EngineConfiguration of key/value data, that is
// used to configure experiment manager/runner
type EngineConfig struct {
	PluginBinary        string
	EngineConfiguration map[string]interface{} `mapstructure:",remain"`
}
