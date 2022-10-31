package fiberapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/caraml-dev/turing/engines/experiment/runner/nop"
	"github.com/caraml-dev/turing/engines/router/missionctl/experiment"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"

	_ "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner/nop"
)

func TestNewRouteSelectionPolicy(t *testing.T) {
	tests := map[string]struct {
		props          json.RawMessage
		expectedPolicy routeSelectionPolicy
		success        bool
		err            string
	}{
		"success": {
			props: json.RawMessage(`{
				"default_route_id":  "route1",
				"experiment_engine": "Test"
			}`),
			expectedPolicy: routeSelectionPolicy{
				defaultRoute: "route1",
			},
			success: true,
		},
		"success | with experiment mappings": {
			props: json.RawMessage(`{
				"default_route_id":  "route-1",
				"experiment_engine": "Test",
				"experiment_mappings": [
					{
					  "experiment": "experiment-1",
					  "treatment": "treatment-1",
					  "route": "route-1"
					},
					{
					  "experiment": "experiment-1",
					  "treatment": "treatment-2",
					  "route": "route-2"
					}
				]
			}`),
			expectedPolicy: routeSelectionPolicy{
				defaultRoute: "route-1",
				experimentMappings: []experimentMapping{
					{Experiment: "experiment-1", Treatment: "treatment-1", Route: "route-1"},
					{Experiment: "experiment-1", Treatment: "treatment-2", Route: "route-2"},
				},
			},
			success: true,
		},
		"success | empty route": {
			props:   json.RawMessage(`{}`),
			success: true,
		},
		"success | with route name path": {
			props: json.RawMessage(`{
				"default_route_id":  "route-1",
				"experiment_engine": "Test",
				"route_name_path": "policy.route_name"			
			}`),
			expectedPolicy: routeSelectionPolicy{
				defaultRoute:  "route-1",
				routeNamePath: "policy.route_name",
			},
			success: true,
		},
		"failure | with experiment mappings and route name path": {
			props: json.RawMessage(`{
				"default_route_id":  "route-1",
				"experiment_engine": "Test",
				"route_name_path": "policy.route_name",	
				"experiment_mappings": [
					{
					  "experiment": "experiment-1",
					  "treatment": "treatment-1",
					  "route": "route-1"
					},
					{
					  "experiment": "experiment-1",
					  "treatment": "treatment-2",
					  "route": "route-2"
					}
				]
			}`),
			success: false,
			err:     "Experiment mappings and route name path cannot both be configured together",
		},
		"failure | invalid data": {
			props:   json.RawMessage(`invalid_data`),
			success: false,
			err:     "Failed to parse route selection policy",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := newRouteSelectionPolicy(data.props)
			assert.Equal(t, data.success, err == nil)
			if data.success {
				tu.FailOnError(t, tu.CompareObjects(*actual, data.expectedPolicy))
			} else {
				tu.FailOnNil(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}

func TestNewExperimentationPolicy(t *testing.T) {
	tests := map[string]struct {
		props          json.RawMessage
		expectedPolicy experimentationPolicy
		success        bool
		err            string
	}{
		"success": {
			props: json.RawMessage(`{
				"default_route_id":  "route1",
				"experiment_engine": "Nop"
			}`),
			expectedPolicy: experimentationPolicy{
				experimentEngine: runner.NewInterceptRunner(
					"Nop",
					&nop.ExperimentRunner{},
					&experiment.MetricsInterceptor{}),
			},
			success: true,
		},
		"failure | invalid experiment engine": {
			props: json.RawMessage(`{
				"experiment_engine": "Test"
			}`),
			success: false,
			err:     `no experiment runner found for name "test"`,
		},
		"failure | invalid data": {
			props:   json.RawMessage(`invalid_data`),
			success: false,
			err:     "Failed to parse experimentation policy",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := newExperimentationPolicy(data.props)
			assert.Equal(t, data.success, err == nil)
			if data.success {
				assert.Equal(t, data.expectedPolicy, *actual)
			} else {
				tu.FailOnNil(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}

func TestCustomRoutingPolicyMarshal(t *testing.T) {
	expectedJSON := `{"DefaultRoute": "route1"}`
	policy := routeSelectionPolicy{
		defaultRoute: "route1",
	}

	actualJSON, err := json.Marshal(policy)

	tu.FailOnError(t, err)
	assert.JSONEq(t, expectedJSON, string(actualJSON))
}

func TestCustomExperimentationPolicyMarshal(t *testing.T) {
	expectedJSON := `{
		"ExperimentEngine":  null
	}`
	policy := experimentationPolicy{
		experimentEngine: nil,
	}

	actualJSON, err := json.Marshal(policy)

	tu.FailOnError(t, err)
	assert.JSONEq(t, expectedJSON, string(actualJSON))
}
