package plugin

import "encoding/json"

type Configurable interface {
	Configure(cfg json.RawMessage) error
}

type PluginCapabilities interface {
	Capabilities() *Capabilities
}

type Capabilities struct {
	Manager bool
	Runner  bool
}
