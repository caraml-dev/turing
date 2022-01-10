package runner

import "context"

// ContextKey is a type that represents a key in a Golang context
type ContextKey string

const (
	// ExperimentEngineKey represents the key for the experiment engine name, stored in the context
	ExperimentEngineKey ContextKey = "experimentEngineKey"
)

// Interceptor interface is used to define concrete interceptors whose methods will
// be run before and after a single fetch treatment call
type Interceptor interface {
	BeforeDispatch(ctx context.Context) context.Context
	AfterCompletion(ctx context.Context, err error)
}
