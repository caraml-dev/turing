package manager

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// ExperimentManager describes the methods to be implemented by an experiment engine's Manager,
// providing access to the information required to set up experiments on Turing.
type ExperimentManager interface {
	// IsCacheEnabled returns whether the experiment engine wants to cache its responses in the Turing API cache
	IsCacheEnabled() bool
	// GetEngineInfo returns the configuration of the experiment engine
	GetEngineInfo() Engine
	// ListClients returns a list of the clients registered on the experiment engine
	ListClients() ([]Client, error)
	// ListExperiments returns a list of the experiments registered on the experiment engine
	ListExperiments() ([]Experiment, error)
	// ListExperimentsForClient returns a list of the experiments registered on the experiment engine,
	// for the given client
	ListExperimentsForClient(Client) ([]Experiment, error)
	// ListVariablesForClient returns a list of the variables registered on the given client
	ListVariablesForClient(Client) ([]Variable, error)
	// ListVariablesForExperiments returns a list of the variables registered on the given experiments
	ListVariablesForExperiments([]Experiment) (map[string][]Variable, error)
	// GetExperimentRunnerConfig converts the given TuringExperimentConfig into a format suitable for the
	// Turing router. TuringExperimentConfig holds the experiment configuration in a format that is suitable
	// for use with the Turing UI and this is the data that is saved to the Turing DB. This interface method
	// will be called at the time of router deployment to convert the data into the format that the router, i.e.,
	// Experiment Runner expects.
	GetExperimentRunnerConfig(TuringExperimentConfig) (json.RawMessage, error)
	// ValidateExperimentConfig validates the given Turing experiment config for the expected data and format,
	// based on the given engine properties
	ValidateExperimentConfig(Engine, TuringExperimentConfig) error
}

var managersLock sync.Mutex
// managers contain all the registered experiment managers by name.
var managers = make(map[string]Factory)

// Factory creates an experiment manager from the provided config.
//
// Config is a raw encoded JSON value. The experiment manager implementation
// for each experiment engine should provide a schema and example
// of the JSON value to explain the usage.
type Factory func(config json.RawMessage) (ExperimentManager, error)

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
func Get(name string, config json.RawMessage) (ExperimentManager, error) {
	managersLock.Lock()
	defer managersLock.Unlock()

	name = strings.ToLower(name)
	m, ok := managers[name]
	if !ok {
		return nil, fmt.Errorf("no experiment manager found for name %s", name)
	}

	return m(config)
}
