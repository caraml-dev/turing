package fiberapi

import (
	"encoding/json"

	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/experiment"
)

// ****************************************************************************
// Define configuration keys expected to be passed in for initialization

type routeSelPolicyCfg struct {
	DefRoute           string              `json:"default_route_id,omitempty"`
	ExperimentMappings []experimentMapping `json:"experiment_mappings"`
}

type expPolicyCfg struct {
	ExpEngine      string                 `json:"experiment_engine,omitempty"`
	ExpEngineProps map[string]interface{} `json:"experiment_engine_properties,omitempty"`
}

// ****************************************************************************

// routeSelectionPolicy captures the properties for gathering the final results
type routeSelectionPolicy struct {
	defaultRoute       string
	experimentMappings []experimentMapping
}

// ExperimentMapping specifies the route that should be selected for a particular treatment in an experiment
type experimentMapping struct {
	Experiment string // experiment name
	Treatment  string // treatment name
	Route      string // route id
}

// MarshalJSON is used to marshal the struct type with private fields into valid json,
// for testing and logging.
func (rsP routeSelectionPolicy) MarshalJSON() ([]byte, error) {
	jsonVal, err := json.Marshal(struct {
		DefaultRoute string
	}{
		DefaultRoute: rsP.defaultRoute,
	})
	if err != nil {
		return nil, err
	}
	return jsonVal, nil
}

// newRouteSelectionPolicy is a creator function for routeSelectionPolicy
func newRouteSelectionPolicy(properties json.RawMessage) (*routeSelectionPolicy, error) {
	var routeSelPolicy routeSelPolicyCfg

	// Unmarshal the properties
	err := json.Unmarshal(properties, &routeSelPolicy)
	if err != nil {
		return nil, errors.Newf(errors.BadConfig, "Failed to parse route selection policy")
	}
	return &routeSelectionPolicy{
		defaultRoute:       routeSelPolicy.DefRoute,
		experimentMappings: routeSelPolicy.ExperimentMappings,
	}, nil
}

// experimentationPolicy captures the common properties for Experimenting
type experimentationPolicy struct {
	experimentEngine runner.ExperimentRunner
}

// MarshalJSON is used to marshal the struct type with private fields into valid json,
// for testing and logging.
func (exP experimentationPolicy) MarshalJSON() ([]byte, error) {
	jsonVal, err := json.Marshal(struct {
		ExperimentEngine runner.ExperimentRunner
	}{
		ExperimentEngine: exP.experimentEngine,
	})
	if err != nil {
		return nil, err
	}
	return jsonVal, nil
}

// newExperimentationPolicy is a creator function for experimentationPolicy
func newExperimentationPolicy(properties json.RawMessage) (*experimentationPolicy, error) {
	var expPolicy expPolicyCfg

	// Unmarshal the properties
	err := json.Unmarshal(properties, &expPolicy)
	if err != nil {
		return nil, errors.Newf(errors.BadConfig, "Failed to parse experimentation policy")
	}

	// Initialize experiment policy
	engine, err := experiment.NewExperimentRunner(expPolicy.ExpEngine, expPolicy.ExpEngineProps)
	if err != nil {
		return nil, err
	}
	return &experimentationPolicy{
		experimentEngine: engine,
	}, nil
}
