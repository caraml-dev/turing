package fiberapi

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"bou.ke/monkey"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
	fiberGrpc "github.com/gojek/fiber/grpc"
	fiberHttp "github.com/gojek/fiber/http"
	"github.com/stretchr/testify/assert"

	runnerV1 "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/caraml-dev/turing/engines/router/missionctl/experiment"
	testutils2 "github.com/caraml-dev/turing/engines/router/missionctl/fiberapi/internal/testutils"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
)

func TestInitializeDefaultRoutingStrategy(t *testing.T) {
	type testSuiteInitStrategy struct {
		properties    json.RawMessage
		success       bool
		expected      DefaultTuringRoutingStrategy
		expectedError string
	}

	tests := map[string]testSuiteInitStrategy{
		"success | with experiment mappings": {
			properties: json.RawMessage(`{
				"default_route_id":  "route1",
				"experiment_engine": "Test",
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
		"success | with route name path": {
			properties: json.RawMessage(`{
				"default_route_id":  "route1",
				"experiment_engine": "Test",
				"route_name_path": "policy.route_name"	
			}`),
			success: true,
			expected: DefaultTuringRoutingStrategy{
				routeSelectionPolicy: &routeSelectionPolicy{
					defaultRoute:  "route1",
					routeNamePath: "policy.route_name",
				},
				experimentationPolicy: &experimentationPolicy{
					experimentEngine: nil,
				},
			},
		},
		"failure | with route name path": {
			properties: json.RawMessage(`{
				"default_route_id":  "route1",
				"experiment_engine": "Test",
				"route_name_path": "policy.route_name",
				"experiment_mappings": [
					{
						"experiment": "experiment-1",
                        "treatment": "control",
                        "route": "route1"
					}
                ]
			}`),
			success: false,
			expectedError: "Failed initializing route selection policy on routing strategy: Experiment mappings and " +
				"route name path cannot both be configured together",
		},
		"failure | missing_route_policy": {
			properties: json.RawMessage(`{
				"experiment_engine": "Test"
			}`),
			success:       false,
			expectedError: "No default route defined",
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Set up
			strategy := DefaultTuringRoutingStrategy{}

			// Monkey patch functionality that is external to the current package and run
			monkey.Patch(
				runnerV1.Get,
				func(name string, config json.RawMessage) (runner.ExperimentRunner, error) {
					return nil, nil
				},
			)
			monkey.Patch(
				experiment.NewExperimentRunner,
				func(_ string, _ map[string]interface{}) (runner.ExperimentRunner, error) {
					return nil, nil
				},
			)
			err := strategy.Initialize(data.properties)

			monkey.Unpatch(experiment.NewExperimentRunner)
			monkey.Unpatch(runnerV1.Get)

			// Test that there is no error and the routing strategy is initialised as expected
			if data.success {
				assert.Nil(t, err)
				assert.Equal(t, data.expected, strategy)
			} else {
				assert.EqualError(t, err, data.expectedError)
			}
		})
	}

}

func TestDefaultRoutingStrategy(t *testing.T) {
	type testSuiteRouting struct {
		endpoints []string
		// treatment that the experiment runner will return
		treatment runner.Treatment
		// experimentMappings in routeSelectionPolicy to select a route from the treatment and experiment in the treatment
		experimentMappings []experimentMapping
		// routeNamePath in routeSelectionPolicy that contains the treatment config path that has the name of the route
		// to be used as the final response
		routeNamePath string
		// if true, experiment runner will return an error when the caller calls GetTreatmentForRequest()
		experimentRunnerWantErr bool
		// defaultRoute in routeSelectionPolicy for fallback, it should match one of the endpoints above to be valid
		defaultRoute      string
		expectedRoute     fiber.Component
		expectedFallbacks []fiber.Component
		expectedError     error
		request           fiber.Request
	}

	// Create test request
	HTTPReq := tu.MakeTestRequest(t, tu.NopHTTPRequestModifier)
	fiberHTTPReq, err := fiberHttp.NewHTTPRequest(HTTPReq)
	tu.FailOnError(t, err)

	tests := map[string]testSuiteRouting{
		"match for treatment and experiment in the mappings should select the correct route": {
			endpoints: []string{"route-A", "route-B"},
			treatment: runner.Treatment{
				ExperimentName: "test_experiment",
				Name:           "treatment-A",
				Config:         json.RawMessage(`{"test_config": "placeholder"}`),
			},
			experimentMappings: []experimentMapping{
				{Experiment: "test_experiment", Treatment: "treatment-0", Route: "route-0"},
				{Experiment: "test_experiment_2", Treatment: "treatment-A", Route: "route-0"},
				{Experiment: "test_experiment", Treatment: "treatment-A", Route: "route-A"},
			},
			expectedRoute: fiber.NewProxy(
				fiber.NewBackend("route-A", ""),
				testutils2.NewFiberCaller(t, "route-A"),
			),
			expectedFallbacks: []fiber.Component{},
		},
		"match for treatment and experiment in the mappings should select the correct route - upi": {
			endpoints: []string{"route-A", "route-B"},
			treatment: runner.Treatment{
				ExperimentName: "test_experiment",
				Name:           "treatment-A",
				Config:         json.RawMessage(`{"test_config": "placeholder"}`),
			},
			experimentMappings: []experimentMapping{
				{Experiment: "test_experiment", Treatment: "treatment-0", Route: "route-0"},
				{Experiment: "test_experiment_2", Treatment: "treatment-A", Route: "route-0"},
				{Experiment: "test_experiment", Treatment: "treatment-A", Route: "route-A"},
			},
			expectedRoute: fiber.NewProxy(
				fiber.NewBackend("route-A", ""),
				testutils2.NewFiberCaller(t, "route-A"),
			),
			expectedFallbacks: []fiber.Component{},
			request: &fiberGrpc.Request{
				Proto: &upiv1.PredictValuesRequest{},
			},
		},
		"match for route name in route name path in treatment config and route names should select the correct route": {
			endpoints: []string{"route-A", "route-B"},
			treatment: runner.Treatment{
				ExperimentName: "test_experiment",
				Name:           "treatment-A",
				Config:         json.RawMessage(`{"test_config": "placeholder", "route_name": "route-A"}`),
			},
			routeNamePath: "route_name",
			expectedRoute: fiber.NewProxy(
				fiber.NewBackend("route-A", ""),
				testutils2.NewFiberCaller(t, "route-A"),
			),
			expectedFallbacks: []fiber.Component{},
		},
		"no match for treatment and experiment in the mappings should select no route and fallback to default route": {
			endpoints: []string{"route-A", "route-B", "control"},
			treatment: runner.Treatment{
				ExperimentName: "test_experiment",
				Name:           "treatment-A",
				Config:         json.RawMessage(`{"test_config": "placeholder"}`),
			},
			experimentMappings: []experimentMapping{
				{Experiment: "test_experiment", Treatment: "treatment-0", Route: "route-B"},
			},
			defaultRoute:  "control",
			expectedRoute: nil,
			expectedFallbacks: []fiber.Component{
				fiber.NewProxy(
					fiber.NewBackend("control", ""),
					testutils2.NewFiberCaller(t, "control"),
				),
			},
		},
		"no match for route name in treatment config and route names should select no route and fallback to default " +
			"route": {
			endpoints: []string{"route-A", "route-B", "control"},
			treatment: runner.Treatment{
				ExperimentName: "test_experiment",
				Name:           "treatment-A",
				Config:         json.RawMessage(`{"test_config": "placeholder", "route_name": "route-C"}`),
			},
			routeNamePath: "route_name",
			defaultRoute:  "control",
			expectedRoute: nil,
			expectedFallbacks: []fiber.Component{
				fiber.NewProxy(
					fiber.NewBackend("control", ""),
					testutils2.NewFiberCaller(t, "control"),
				),
			},
		},
		"route name path not be found in treatment config should select no route and fallback to default route": {
			endpoints: []string{"route-A", "route-B", "control"},
			treatment: runner.Treatment{
				ExperimentName: "test_experiment",
				Name:           "treatment-A",
				Config:         json.RawMessage(`{"test_config": "placeholder", "route_name": "route-A"}`),
			},
			routeNamePath: "policy.route_name",
			defaultRoute:  "control",
			expectedRoute: nil,
			expectedFallbacks: []fiber.Component{
				fiber.NewProxy(
					fiber.NewBackend("control", ""),
					testutils2.NewFiberCaller(t, "control"),
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
					testutils2.NewFiberCaller(t, "control"),
				),
			},
		},
		"no route name path configured should select no route and fallback to default route": {
			endpoints: []string{"route-A", "route-B"},
			treatment: runner.Treatment{
				ExperimentName: "test_experiment",
				Name:           "treatment-A",
				Config:         json.RawMessage(`{"test_config": "placeholder", "route_name": "route-A"}`),
			},
			expectedRoute: nil,
			defaultRoute:  "route-B",
			expectedFallbacks: []fiber.Component{
				fiber.NewProxy(
					fiber.NewBackend("route-B", ""),
					testutils2.NewFiberCaller(t, "route-B"),
				),
			},
		},
		"no route name path, no experiment mappings, and no default route should select no route and have no fallback " +
			"to default route": {
			endpoints:         []string{"treatment-B", "treatment-C"},
			expectedRoute:     nil,
			expectedFallbacks: []fiber.Component{},
		},
		"error when calling GetTreatmentForRequest() should fallback to default route": {
			endpoints:               []string{"treatment-B", "control"},
			experimentRunnerWantErr: true,
			expectedRoute:           nil,
			defaultRoute:            "control",
			expectedFallbacks: []fiber.Component{
				fiber.NewProxy(
					fiber.NewBackend("control", ""),
					testutils2.NewFiberCaller(t, "control"),
				),
			},
		},
		"grpc non upi request": {
			endpoints:     []string{"control", "route-A"},
			defaultRoute:  "control",
			request:       &fiberGrpc.Request{},
			expectedError: errors.New("failed to convert into UPI request"),
		},
		"upi non upi request": {
			endpoints:    []string{"control", "route-A"},
			defaultRoute: "control",
			request: &fiberGrpc.Request{
				Proto: &upiv1.PredictValuesRequest{
					PredictionContext: []*upiv1.Variable{
						{
							Name: "unknown",
							Type: 0,
						},
					},
				},
			},
			expectedError: errors.New("Unknown value type TYPE_UNSPECIFIED"),
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Create test routes
			routes := makeTestRoutes(t, data.endpoints...)

			// Test SelectRoute
			strategy := DefaultTuringRoutingStrategy{
				&experimentationPolicy{
					experimentEngine: testutils2.MockExperimentRunner{
						Treatment: &data.treatment,
						WantErr:   data.experimentRunnerWantErr,
					},
				},
				&routeSelectionPolicy{
					defaultRoute:       data.defaultRoute,
					experimentMappings: data.experimentMappings,
					routeNamePath:      data.routeNamePath,
				},
			}
			var fiberReq fiber.Request
			fiberReq = fiberHTTPReq
			if data.request != nil {
				fiberReq = data.request
			}
			route, fallbacks, err := strategy.SelectRoute(context.Background(), fiberReq, routes)

			assert.Equal(t, data.expectedError, err)
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
		routes[n] = fiber.NewProxy(b, testutils2.NewFiberCaller(t, n))
	}
	return routes
}
