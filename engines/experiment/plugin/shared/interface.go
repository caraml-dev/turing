package shared

// RPCClient is a minimal interface to describe *rpc.Client implementation
// Used to simplify mocking of the *rpc.Client in unit tests
type RPCClient interface {
	Call(serviceMethod string, args interface{}, reply interface{}) error
}
