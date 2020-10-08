package manager

import "encoding/json"

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
