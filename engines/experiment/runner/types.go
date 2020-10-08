package runner

import "context"

// ContextKey is a type that represents a key in a Golang context
type ContextKey string

const (
	// ExperimentNameKey represents the key for the name of the experiment, stored in the context
	ExperimentNameKey ContextKey = "experimentNameKey"
)

// Logger interface defines a minimal set of methods expected to be implemented by
// a concrete logger that is passed as a parameter, for use within the library
type Logger interface {
	Errorf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Panicf(template string, args ...interface{})
	Sync() error
}

// Interceptor interface is used to define concrete interceptors whose methods will
// be run before and after a single run experiment call.
type Interceptor interface {
	BeforeDispatch(ctx context.Context) context.Context
	AfterCompletion(ctx context.Context, err error)
}
