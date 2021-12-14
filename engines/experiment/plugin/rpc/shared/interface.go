package shared

import "encoding/json"

// RPCClient is a minimal interface to describe *rpc.Client implementation
// Used to simplify mocking of the *rpc.Client in unit tests
type RPCClient interface {
	Call(serviceMethod string, args interface{}, reply interface{}) error
}

// Configurable interface can be implemented by plugins that require
// to be configured with a JSON config data
type Configurable interface {
	Configure(cfg json.RawMessage) error
}
