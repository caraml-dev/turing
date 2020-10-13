package fiberapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"bou.ke/monkey"
	"github.com/gojek/fiber"
	fiberhttp "github.com/gojek/fiber/http"
	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/gojek/turing/engines/router/missionctl/experiment"
	tfu "github.com/gojek/turing/engines/router/missionctl/fiberapi/internal/testutils"
	tu "github.com/gojek/turing/engines/router/missionctl/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestInitializeDefaultRoutingStrategy(t *testing.T) {
	type testSuiteInitStrategy struct {
		properties json.RawMessage
		success    bool
		expected   DefaultTuringRoutingStrategy
	}

	tests := map[string]testSuiteInitStrategy{
		"success": {
			properties: json.RawMessage(`{
				"default_route_id":  "route1",
				"experiment_engine": "Litmus",
				"experiment_mappings": [
					{
						"experiment": "experiment-1",
                        "treatment": "control",
                        "route": "route1"
					}
                ]
			}`),
			success: true,
			expected: DefaultTuringRoutingStrategy{
				routeSelectionPolicy: &routeSelectionPolicy{
					defaultRoute: "route1",
					experimentMappings: []experimentMapping{
						{Experiment: "experiment-1", Treatment: "control", Route: "route1"},
					},
				},
				experimentationPolicy: &experimentationPolicy{
					experimentEngine: nil,
				},
			},
		},
		"missing_route_policy": {
			properties: json.RawMessage(`{
				"experiment_engine": "Litmus"
			}`),
			success: false,
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Set up
			strategy := DefaultTuringRoutingStrategy{}

			// Monkey patch functionality that is external to the current package and run
			monkey.Patch(
				runner.Get,
				func(name string, config json.RawMessage) (runner.ExperimentRunner, error) {
					return nil, nil
				},
			)
			monkey.Patch(
				experiment.NewExperimentRunner,
				func(_ string, _ json.RawMessage) (runner.ExperimentRunner, error) {
					return nil, nil
				},
			)
			err := strategy.Initialize(data.properties)

			monkey.Unpatch(experiment.NewExperimentRunner)
			monkey.Unpatch(runner.Get)

			// Test that there is no error and fanIn is initialised as expected
			assert.Equal(t, data.success, err == nil)
			if data.success {
				assert.Equal(t, data.expected, strategy)
			}
		})
	}

}

func TestDefaultRoutingStrategy(t *testing.T) {
	type testSuiteRouting struct {
		endpoints []string
		// treatment that the experiment runner will return
		treatment tfu.TestTreatment
		// experimentMappings in routeSelectionPolicy to select a route from the treatment and experiment in the treatment
		experimentMappings []experimentMapping
		// if true, experiment runner will return an error when the caller calls GetTreatmentForRequest()
		experimentRunnerWantErr bool
		// defaultRoute in routeSelectionPolicy for fallback, it should match one of the endpoints above to be valid
		defaultRoute      string
		expectedRoute     fiber.Component
		expectedFallbacks []fiber.Component
	}

	tests := map[string]testSuiteRouting{
		"match for treatment and experiment in the mappings should select the correct route": {
			endpoints: []string{"route-A", "route-B"},
			treatment: tfu.TestTreatment{
				Name:      "test_experiment",
				Treatment: "treatment-A",
				Raw:       json.RawMessage(`{"test_config": "placeholder"}`),
			},
			experimentMappings: []experimentMapping{
				{Experiment: "test_experiment", Treatment: "treatment-0", Route: "route-0"},
				{Experiment: "test_experiment_2", Treatment: "treatment-A", Route: "route-0"},
				{Experiment: "test_experiment", Treatment: "treatment-A", Route: "route-A"},
			},
			expectedRoute: fiber.NewProxy(
				fiber.NewBackend("route-A", ""),
				tfu.NewFiberCallerWithHTTPDispatcher(t, "route-A"),
			),
			expectedFallbacks: []fiber.Component{},
		},
		"no match for treatment and experiment in the mappings should select no route and fallback to default route": {
			endpoints: []string{"route-A", "route-B", "control"},
			treatment: tfu.TestTreatment{
				Name:      "test_experiment",
				Treatment: "treatment-A",
				Raw:       json.RawMessage(`{"test_config": "placeholder"}`),
			},
			experimentMappings: []experimentMapping{
				{Experiment: "test_experiment", Treatment: "treatment-0", Route: "route-B"},
			},
			defaultRoute:  "control",
			expectedRoute: nil,
			expectedFallbacks: []fiber.Component{
				fiber.NewProxy(
					fiber.NewBackend("control", ""),
					tfu.NewFiberCallerWithHTTPDispatcher(t, "control"),
				),
			},
		},
		"no experiment mappings configured should select no route and fallback to default route": {
			endpoints:     []string{"treatment-B", "control"},
			expectedRoute: nil,
			defaultRoute:  "control",
			expectedFallbacks: []fiber.Component{
				fiber.NewProxy(
					fiber.NewBackend("control", ""),
					tfu.NewFiberCallerWithHTTPDispatcher(t, "control"),
				),
			},
		},
		"no experiment mappings and no default route should select no route and have no fallback to default route": {
			endpoints:         []string{"treatment-B", "treatment-C"},
			expectedRoute:     nil,
			expectedFallbacks: []fiber.Component{},
		},
		"simulate experiment engine returns error when calling GetTreatmentForRequest()": {
			endpoints:               []string{"treatment-B", "treatment-C"},
			experimentRunnerWantErr: true,
			expectedRoute:           nil,
			expectedFallbacks:       []fiber.Component{},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Create test request
			req := tu.MakeTestRequest(t, tu.NopHTTPRequestModifier)
			fiberReq, err := fiberhttp.NewHTTPRequest(req)
			tu.FailOnError(t, err)

			// Create test routes
			routes := makeTestRoutes(t, data.endpoints...)

			// Test SelectRoute
			strategy := DefaultTuringRoutingStrategy{
				&experimentationPolicy{
					experimentEngine: tfu.MockExperimentRunner{
						TestTreatment: data.treatment,
						WantErr:       data.experimentRunnerWantErr,
					},
				},
				&routeSelectionPolicy{
					defaultRoute:       data.defaultRoute,
					experimentMappings: data.experimentMappings,
				},
			}
			route, fallbacks, err := strategy.SelectRoute(context.Background(), fiberReq, routes)

			if (err != nil) != data.experimentRunnerWantErr {
				t.Errorf("SelectRoute() error = %v, wantErr %v", err, data.experimentRunnerWantErr)
			}
			assert.Equal(t, data.expectedRoute, route)
			assert.Equal(t, data.expectedFallbacks, fallbacks)
		})
	}
}

// For every endpoint name, this method creates a fiber proxy with the given name
// and an empty endpoint and returns a map of the endpoint name and proxy
func makeTestRoutes(t *testing.T, names ...string) map[string]fiber.Component {
	var routes = make(map[string]fiber.Component, len(names))

	for _, n := range names {
		b := fiber.NewBackend(n, "")
		httpDispatcher, err := fiberhttp.NewDispatcher(http.DefaultClient)
		if err != nil {
			t.Fatal(err)
		}
		caller, err := fiber.NewCaller(n, httpDispatcher)
		if err != nil {
			t.Fatal(err)
		}
		routes[n] = fiber.NewProxy(b, caller)
	}
	return routes
}
