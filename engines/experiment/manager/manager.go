package manager

import (
	"encoding/json"
)

// ExperimentManager describes the minimal set of methods to be implemented by an
// experiment engine's Manager, providing access to the information required to set up
// experiments on Turing.
type ExperimentManager interface {
	// GetEngineInfo returns the configuration of the experiment engine
	GetEngineInfo() Engine

	// ValidateExperimentConfig validates the given Turing experiment config for the expected data and format
	ValidateExperimentConfig(cfg json.RawMessage) error

	// GetExperimentRunnerConfig converts the given config (as retrieved from the DB) into a format suitable
	// for the Turing router (i.e., to be passed to the Experiment Runner). This interface method will be
	// called at the time of router deployment.
	//
	// cfg holds the experiment configuration in a format that is suitable for use with the Turing UI and
	// this is the data that is saved to the Turing DB.
	//
	// In case of StandardExperimentManager, cfg is expected to be unmarshalled into TuringExperimentConfig
	GetExperimentRunnerConfig(cfg interface{}) (json.RawMessage, error)
}

type StandardExperimentManager interface {
	ExperimentManager

	//// GetExperimentRunnerConfig converts the given TuringExperimentConfig into a format suitable for the
	//// Turing router. TuringExperimentConfig holds the experiment configuration in a format that is suitable
	//// for use with the Turing UI and this is the data that is saved to the Turing DB. This interface method
	//// will be called at the time of router deployment to convert the data into the format that the router, i.e.,
	//// Experiment Runner expects.
	//GetExperimentRunnerConfig(TuringExperimentConfig) (json.RawMessage, error)

	// BaseStandardExperimentManager provides default implementations for the following methods
	// that may be composed into the experiment engine.

	// IsCacheEnabled returns whether the experiment engine wants to cache its responses in the Turing API cache
	IsCacheEnabled() bool
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
	// ValidateExperimentConfig validates the given Turing experiment config for the expected data and format,
	// based on the given engine's properties.
	//ValidateExperimentConfig(*StandardExperimentManagerConfig, TuringExperimentConfig) error
}

type CustomExperimentManager interface {
	ExperimentManager
}

func GetStandardExperimentConfig(cfg interface{}) (TuringExperimentConfig, error) {
	var stdExpCfg TuringExperimentConfig

	// Marshal to json
	bytes, err := json.Marshal(cfg)
	if err != nil {
		return stdExpCfg, err
	}

	// Unmarshal using the TuringExperimentConfig type
	err = json.Unmarshal(bytes, &stdExpCfg)
	return stdExpCfg, err
}
