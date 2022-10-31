package fiberapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gojek/fiber"
	fiberhttp "github.com/gojek/fiber/http"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
	"github.com/caraml-dev/turing/engines/router"
	"github.com/caraml-dev/turing/engines/router/missionctl/fiberapi"
	tfu "github.com/caraml-dev/turing/engines/router/missionctl/fiberapi/internal/testutils"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

func TestTrafficSplittingStrategyRule_TestRequest(t *testing.T) {
	type testCase struct {
		rule          *fiberapi.TrafficSplittingStrategyRule
		header        http.Header
		payload       string
		expected      bool
		expectedError string
	}

	suite := map[string]testCase{
		"success": {
			rule: &fiberapi.TrafficSplittingStrategyRule{
				Conditions: []*router.TrafficRuleCondition{
					{
						FieldSource: request.HeaderFieldSource,
						Field:       "X-Region",
						Operator:    router.InConditionOperator,
						Values:      []string{"region-a", "region-b"},
					},
					{
						FieldSource: request.PayloadFieldSource,
						Field:       "foo.bar",
						Operator:    router.InConditionOperator,
						Values:      []string{"foobar"},
					},
				},
			},
			header: http.Header{
				"X-Region": []string{"region-a"},
			},
			payload:  `{"foo": {"bar": "foobar"}}`,
			expected: true,
		},
		"failure | one condition false": {
			rule: &fiberapi.TrafficSplittingStrategyRule{
				Conditions: []*router.TrafficRuleCondition{
					{
						FieldSource: request.HeaderFieldSource,
						Field:       "X-Region",
						Operator:    router.InConditionOperator,
						Values:      []string{"region-a", "region-b"},
					},
					{
						FieldSource: request.PayloadFieldSource,
						Field:       "another",
						Operator:    router.InConditionOperator,
						Values:      []string{"42"},
					},
					{
						FieldSource: request.PayloadFieldSource,
						Field:       "foo.bar",
						Operator:    router.InConditionOperator,
						Values:      []string{"foobar"},
					},
				},
			},
			header: http.Header{
				"X-Region": []string{"region-a"},
			},
			payload:  `{"foo": {"bar": "foobar"}, "another": "19"}`,
			expected: false,
		},
		"failure | all conditions false": {
			rule: &fiberapi.TrafficSplittingStrategyRule{
				Conditions: []*router.TrafficRuleCondition{
					{
						FieldSource: request.HeaderFieldSource,
						Field:       "X-Region",
						Operator:    router.InConditionOperator,
						Values:      []string{"region-a", "region-b"},
					},
					{
						FieldSource: request.PayloadFieldSource,
						Field:       "another",
						Operator:    router.InConditionOperator,
						Values:      []string{"42"},
					},
					{
						FieldSource: request.PayloadFieldSource,
						Field:       "foo.bar",
						Operator:    router.InConditionOperator,
						Values:      []string{"foobar"},
					},
				},
			},
			header: http.Header{
				"X-Region": []string{"region-c"},
			},
			payload:  `{"foo": {"bar": "foobazz"}}`,
			expected: false,
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {

			actual, err := tt.rule.TestRequest(tt.header, []byte(tt.payload))
			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

func TestTuringTrafficSplittingStrategy_Initialize(t *testing.T) {
	type testCase struct {
		properties    json.RawMessage
		strategy      *fiberapi.TrafficSplittingStrategy
		expectedError string
	}

	suite := map[string]testCase{
		"success": {
			properties: json.RawMessage(`{
				"rules": [
					{
 						"route_id": "ROUTE-A",
						"conditions": [{
							"field_source": "header",
							"field": "Content-Type",
							"operator": "in",
							"values": ["application/json"]
						}]
					},
					{
 						"route_id": "ROUTE-B",
						"conditions": [{
							"field_source": "header",
							"field": "Content-Type",
							"operator": "in",
							"values": ["application/xml"]
						}]
					}
				]
			}`),
			strategy: &fiberapi.TrafficSplittingStrategy{
				Rules: []*fiberapi.TrafficSplittingStrategyRule{
					{
						RouteID: "ROUTE-A",
						Conditions: []*router.TrafficRuleCondition{
							{
								FieldSource: request.HeaderFieldSource,
								Field:       "Content-Type",
								Operator:    router.InConditionOperator,
								Values:      []string{"application/json"},
							},
						},
					},
					{
						RouteID: "ROUTE-B",
						Conditions: []*router.TrafficRuleCondition{
							{
								FieldSource: request.HeaderFieldSource,
								Field:       "Content-Type",
								Operator:    router.InConditionOperator,
								Values:      []string{"application/xml"},
							},
						},
					},
				},
			},
		},
		"failure | invalid config": {
			properties: json.RawMessage(`{}`),
			expectedError: "Failed initializing traffic splitting strategy: " +
				"Key: 'TrafficSplittingStrategy.Rules' " +
				"Error:Field validation for 'Rules' failed on the 'required' tag",
		},
		"failure | invalid type": {
			properties: json.RawMessage(`42`),
			expectedError: "Failed initializing traffic splitting strategy: " +
				"json: cannot unmarshal number into Go value of type fiberapi.TrafficSplittingStrategy",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			strategy := new(fiberapi.TrafficSplittingStrategy)
			err := strategy.Initialize(tt.properties)
			if tt.expectedError == "" {
				require.NoError(t, err)
				tu.FailOnError(t, tu.CompareObjects(strategy, tt.strategy))
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

func TestTrafficSplittingStrategy_SelectRoute(t *testing.T) {
	type testCase struct {
		strategy      *fiberapi.TrafficSplittingStrategy
		routes        map[string]fiber.Component
		header        http.Header
		payload       string
		expected      fiber.Component
		fallbacks     []fiber.Component
		expectedError string
	}

	suite := map[string]testCase{
		"success": {
			strategy: &fiberapi.TrafficSplittingStrategy{
				Rules: []*fiberapi.TrafficSplittingStrategyRule{
					{
						RouteID: "route-a",
						Conditions: []*router.TrafficRuleCondition{
							{
								FieldSource: request.HeaderFieldSource,
								Field:       "X-Region",
								Operator:    router.InConditionOperator,
								Values:      []string{"region-a", "region-b"},
							},
							{
								FieldSource: request.PayloadFieldSource,
								Field:       "service_type",
								Operator:    router.InConditionOperator,
								Values:      []string{"service-b"},
							},
						},
					},
				},
			},
			routes: map[string]fiber.Component{
				"route-a": tfu.NewFiberCallerWithHTTPDispatcher(t, "route-a"),
			},
			header: http.Header{
				"X-Region": []string{"region-b"},
			},
			payload:   `{"service_type": "service-b"}`,
			expected:  tfu.NewFiberCallerWithHTTPDispatcher(t, "route-a"),
			fallbacks: []fiber.Component{},
		},
		"success | with default route": {
			strategy: &fiberapi.TrafficSplittingStrategy{
				DefaultRouteID: "control",
				Rules: []*fiberapi.TrafficSplittingStrategyRule{
					{
						RouteID: "route-a",
						Conditions: []*router.TrafficRuleCondition{
							{
								FieldSource: request.HeaderFieldSource,
								Field:       "X-Region",
								Operator:    router.InConditionOperator,
								Values:      []string{"region-a", "region-b"},
							},
							{
								FieldSource: request.PayloadFieldSource,
								Field:       "service_type",
								Operator:    router.InConditionOperator,
								Values:      []string{"service-b"},
							},
						},
					},
					{
						RouteID: "route-b",
						Conditions: []*router.TrafficRuleCondition{
							{
								FieldSource: request.PayloadFieldSource,
								Field:       "service_type",
								Operator:    router.InConditionOperator,
								Values:      []string{"service-a", "service-b"},
							},
						},
					},
				},
			},
			routes: map[string]fiber.Component{
				"control": tfu.NewFiberCallerWithHTTPDispatcher(t, "control"),
				"route-a": tfu.NewFiberCallerWithHTTPDispatcher(t, "route-a"),
				"route-b": tfu.NewFiberCallerWithHTTPDispatcher(t, "route-b"),
			},
			payload:   `{"service_type": "service-c"}`,
			expected:  tfu.NewFiberCallerWithHTTPDispatcher(t, "control"),
			fallbacks: []fiber.Component{},
		},
		"success | with fallbacks": {
			strategy: &fiberapi.TrafficSplittingStrategy{
				Rules: []*fiberapi.TrafficSplittingStrategyRule{
					{
						RouteID: "route-a",
						Conditions: []*router.TrafficRuleCondition{
							{
								FieldSource: request.HeaderFieldSource,
								Field:       "X-Region",
								Operator:    router.InConditionOperator,
								Values:      []string{"region-a", "region-b"},
							},
							{
								FieldSource: request.PayloadFieldSource,
								Field:       "service_type",
								Operator:    router.InConditionOperator,
								Values:      []string{"service-b"},
							},
						},
					},
					{
						RouteID: "route-b",
						Conditions: []*router.TrafficRuleCondition{
							{
								FieldSource: request.PayloadFieldSource,
								Field:       "service_type",
								Operator:    router.InConditionOperator,
								Values:      []string{"service-a", "service-b"},
							},
						},
					},
				},
			},
			routes: map[string]fiber.Component{
				"route-a": tfu.NewFiberCallerWithHTTPDispatcher(t, "route-a"),
				"route-b": tfu.NewFiberCallerWithHTTPDispatcher(t, "route-b"),
			},
			header: http.Header{
				"X-Region": []string{"region-b"},
			},
			payload:   `{"service_type": "service-b"}`,
			expected:  tfu.NewFiberCallerWithHTTPDispatcher(t, "route-a"),
			fallbacks: []fiber.Component{tfu.NewFiberCallerWithHTTPDispatcher(t, "route-b")},
		},
		"failure | request doesn't match any rule": {
			strategy: &fiberapi.TrafficSplittingStrategy{
				Rules: []*fiberapi.TrafficSplittingStrategyRule{
					{
						RouteID: "route-a",
						Conditions: []*router.TrafficRuleCondition{
							{
								FieldSource: request.HeaderFieldSource,
								Field:       "X-Region",
								Operator:    router.InConditionOperator,
								Values:      []string{"region-a"},
							},
						},
					},
				},
			},
			routes: map[string]fiber.Component{
				"route-a": tfu.NewFiberCallerWithHTTPDispatcher(t, "route-a"),
			},
			header: http.Header{
				"X-Region": []string{"region-d"},
			},
			payload:       "",
			fallbacks:     []fiber.Component{},
			expectedError: "http request didn't match any traffic rule",
		},
		"failure | bad configuration": {
			strategy: &fiberapi.TrafficSplittingStrategy{
				Rules: []*fiberapi.TrafficSplittingStrategyRule{
					{
						RouteID: "route-a",
						Conditions: []*router.TrafficRuleCondition{
							{
								FieldSource: request.HeaderFieldSource,
								Field:       "X-Region",
								Operator:    router.InConditionOperator,
								Values:      []string{"region-a"},
							},
						},
					},
				},
			},
			routes: map[string]fiber.Component{
				"route-b": tfu.NewFiberCallerWithHTTPDispatcher(t, "route-b"),
				"route-c": tfu.NewFiberCallerWithHTTPDispatcher(t, "route-c"),
			},
			header: http.Header{
				"X-Region": []string{"region-a"},
			},
			payload:       "",
			fallbacks:     []fiber.Component{},
			expectedError: `route with id "route-a" doesn't exist in the router`,
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			ctx := turingctx.NewTuringContext(context.Background())
			req, _ := http.NewRequest(http.MethodPost, "/predict", bytes.NewReader([]byte(tt.payload)))
			req.Header = tt.header
			fiberReq, _ := fiberhttp.NewHTTPRequest(req)

			actual, actualFallbacks, err := tt.strategy.SelectRoute(ctx, fiberReq, tt.routes)
			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
				require.Equal(t, tt.fallbacks, actualFallbacks)
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
