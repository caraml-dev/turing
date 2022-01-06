package runner

import (
	"context"
	"encoding/json"
)

// ContextKey is a type that represents a key in a Golang context
type ContextKey string

const (
	// ExperimentEngineKey represents the key for the experiment engine name, stored in the context
	ExperimentEngineKey ContextKey = "experimentEngineKey"
)

// Interceptor interface is used to define concrete interceptors whose methods will
// be run before and after a single run experiment call.
type Interceptor interface {
	BeforeDispatch(ctx context.Context) context.Context
	AfterCompletion(ctx context.Context, err error)
}

type Treatment struct {
	ExperimentName string
	Name           string
	Config         json.RawMessage
}
