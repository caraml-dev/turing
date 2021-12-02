package manager

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/gojek/turing/engines/experiment/v2/manager"
)

var managersLock sync.Mutex

// managers contain all the registered experiment managers by name.
var managers = make(map[string]Factory)

// Factory creates an experiment manager from the provided config.
//
// Config is a raw encoded JSON value. The experiment manager implementation
// for each experiment engine should provide a schema and example
// of the JSON value to explain the usage.
type Factory func(config json.RawMessage) (manager.ExperimentManager, error)

// Register an experiment manager with the provided name and factory function.
//
// For registration to be properly recorded, Register function should be called in the init
// phase of the Go execution. The init function is usually defined in the package where
// the manager is implemented. The name of the experiment manager should be unique
// across all implementations. Registering multiple experiment managers with the
// same name will return an error.
func Register(name string, factory Factory) error {
	managersLock.Lock()
	defer managersLock.Unlock()

	name = strings.ToLower(name)
	if _, found := managers[name]; found {
		return fmt.Errorf("experiment manager %q was registered twice", name)
	}

	managers[name] = factory
	return nil
}

// Get an experiment manager that has been registered.
//
// The manager will be initialized using the registered factory function with the provided config.
// Retrieving an experiment manager that is not yet registered will return an error.
func Get(name string, config json.RawMessage) (manager.ExperimentManager, error) {
	managersLock.Lock()
	defer managersLock.Unlock()

	name = strings.ToLower(name)
	m, ok := managers[name]
	if !ok {
		return nil, fmt.Errorf("no experiment manager found for name %s", name)
	}

	return m(config)
}
