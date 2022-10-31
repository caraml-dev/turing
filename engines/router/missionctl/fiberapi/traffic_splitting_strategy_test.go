package fiberapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
	fiberHttp "github.com/gojek/fiber/http"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
	"github.com/caraml-dev/turing/engines/router"
	"github.com/caraml-dev/turing/engines/router/missionctl/fiberapi"
	tfu "github.com/caraml-dev/turing/engines/router/missionctl/fiberapi/internal/testutils"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

var (
	// to avoid compiler optimisation in benchmark
	primaryRoute fiber.Component
)

func TestTrafficSplittingStrategyRule_TestRequest(t *testing.T) {
	type testCase struct {
		rule          *fiberapi.TrafficSplittingStrategyRule
		request       fiber.Request
		expected      bool
		expectedError string
	}

	suite := map[string]testCase{
		"success | http": {
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
			request: tfu.NewHTTPFiberRequest(t, http.Header{
				"X-Region": []string{"region-a"},
			}, `{"foo": {"bar": "foobar"}}`),

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
			request: tfu.NewHTTPFiberRequest(t, http.Header{
				"X-Region": []string{"region-a"},
			}, `{"foo": {"bar": "foobar"}, "another": "19"}`),

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
			request: tfu.NewHTTPFiberRequest(t, http.Header{
				"X-Region": []string{"region-c"},
			}, `{"foo": {"bar": "foobazz"}}`),
			expected: false,
		},
		"success | upi": {
			rule: &fiberapi.TrafficSplittingStrategyRule{
				Conditions: []*router.TrafficRuleCondition{
					{
						FieldSource: request.HeaderFieldSource,
						Field:       "X-Region",
						Operator:    router.InConditionOperator,
						Values:      []string{"region-a", "region-b"},
					},
					{
						FieldSource: request.PredictionContextSource,
						Field:       "foo",
						Operator:    router.InConditionOperator,
						Values:      []string{"foobar"},
					},
				},
			},
			request: tfu.NewUPIFiberRequest(t, map[string]string{
				"X-Region": "region-a",
			}, &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "foo",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "foobar",
					},
				},
			}),
			expected: true,
		},
		"success | upi with non string prediction context": {
			rule: &fiberapi.TrafficSplittingStrategyRule{
				Conditions: []*router.TrafficRuleCondition{
					{
						FieldSource: request.HeaderFieldSource,
						Field:       "X-Region",
						Operator:    router.InConditionOperator,
						Values:      []string{"region-a", "region-b"},
					},
					{
						FieldSource: request.PredictionContextSource,
						Field:       "foo",
						Operator:    router.InConditionOperator,
						Values:      []string{"42"},
					},
				},
			},
			request: tfu.NewUPIFiberRequest(t, map[string]string{
				"X-Region": "region-a",
			}, &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:         "foo",
						Type:         upiv1.Type_TYPE_INTEGER,
						IntegerValue: 42,
					},
				},
			}),
			expected: true,
		},
		"failure | upi with one condition false": {
			rule: &fiberapi.TrafficSplittingStrategyRule{
				Conditions: []*router.TrafficRuleCondition{
					{
						FieldSource: request.HeaderFieldSource,
						Field:       "X-Region",
						Operator:    router.InConditionOperator,
						Values:      []string{"region-a", "region-b"},
					},
					{
						FieldSource: request.PredictionContextSource,
						Field:       "another",
						Operator:    router.InConditionOperator,
						Values:      []string{"42"},
					},
					{
						FieldSource: request.PredictionContextSource,
						Field:       "foo.bar",
						Operator:    router.InConditionOperator,
						Values:      []string{"foobar"},
					},
				},
			},
			request: tfu.NewUPIFiberRequest(t, map[string]string{
				"X-Region": "region-a",
			}, &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "foo",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "foobar",
					},
				},
			}),

			expected: false,
		},
		"failure | upi with all conditions false": {
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
			request: tfu.NewUPIFiberRequest(t, map[string]string{
				"X-Region": "region-c",
			}, &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "foo",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "foobar",
					},
				},
			}),
			expected: false,
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			actual, err := tt.rule.TestRequest(tt.request)
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
		"success | upi": {
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
							"field_source": "prediction_context",
							"field": "my-variable",
							"operator": "in",
							"values": ["my-variable-value"]
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
								FieldSource: request.PredictionContextSource,
								Field:       "my-variable",
								Operator:    router.InConditionOperator,
								Values:      []string{"my-variable-value"},
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
		request       fiber.Request
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
				"route-a": tfu.NewFiberCaller(t, "route-a"),
			},
			request: tfu.NewHTTPFiberRequest(t, http.Header{
				"X-Region": []string{"region-b"},
			}, `{"service_type": "service-b"}`),
			expected:  tfu.NewFiberCaller(t, "route-a"),
			fallbacks: []fiber.Component{},
		},
		"success | upi": {
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
								FieldSource: request.PredictionContextSource,
								Field:       "service_type",
								Operator:    router.InConditionOperator,
								Values:      []string{"service-b"},
							},
						},
					},
				},
			},
			routes: map[string]fiber.Component{
				"route-a": tfu.NewFiberCaller(t, "route-a"),
			},
			request: tfu.NewUPIFiberRequest(t, map[string]string{
				"X-Region": "region-a",
			}, &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "service_type",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "service-b",
					},
				},
			}),
			expected:  tfu.NewFiberCaller(t, "route-a"),
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
				"control": tfu.NewFiberCaller(t, "control"),
				"route-a": tfu.NewFiberCaller(t, "route-a"),
				"route-b": tfu.NewFiberCaller(t, "route-b"),
			},
			request:   tfu.NewHTTPFiberRequest(t, http.Header{}, `{"service_type": "service-c"}`),
			expected:  tfu.NewFiberCaller(t, "control"),
			fallbacks: []fiber.Component{},
		},
		"success | upi with default route": {
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
								FieldSource: request.PredictionContextSource,
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
								FieldSource: request.PredictionContextSource,
								Field:       "service_type",
								Operator:    router.InConditionOperator,
								Values:      []string{"service-a", "service-b"},
							},
						},
					},
				},
			},
			routes: map[string]fiber.Component{
				"control": tfu.NewFiberCaller(t, "control"),
				"route-a": tfu.NewFiberCaller(t, "route-a"),
				"route-b": tfu.NewFiberCaller(t, "route-b"),
			},
			request: tfu.NewUPIFiberRequest(t, map[string]string{
				"X-Region": "region-a",
			}, &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "foo",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "foobar",
					},
				},
			}),
			expected:  tfu.NewFiberCaller(t, "control"),
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
				"route-a": tfu.NewFiberCaller(t, "route-a"),
				"route-b": tfu.NewFiberCaller(t, "route-b"),
			},
			request: tfu.NewHTTPFiberRequest(t,
				http.Header{"X-Region": []string{"region-b"}},
				`{"service_type": "service-b"}`),
			expected:  tfu.NewFiberCaller(t, "route-a"),
			fallbacks: []fiber.Component{tfu.NewFiberCaller(t, "route-b")},
		},
		"success | upi with fallbacks": {
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
								FieldSource: request.PredictionContextSource,
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
								FieldSource: request.PredictionContextSource,
								Field:       "service_type",
								Operator:    router.InConditionOperator,
								Values:      []string{"service-a", "service-b"},
							},
						},
					},
				},
			},
			routes: map[string]fiber.Component{
				"route-a": tfu.NewFiberCaller(t, "route-a"),
				"route-b": tfu.NewFiberCaller(t, "route-b"),
			},
			request: tfu.NewUPIFiberRequest(t, map[string]string{
				"X-Region": "region-a",
			}, &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "service_type",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "service-b",
					},
				},
			}),
			expected:  tfu.NewFiberCaller(t, "route-a"),
			fallbacks: []fiber.Component{tfu.NewFiberCaller(t, "route-b")},
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
				"route-a": tfu.NewFiberCaller(t, "route-a"),
			},
			request:       tfu.NewHTTPFiberRequest(t, http.Header{"X-Region": []string{"region-d"}}, ``),
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
				"route-b": tfu.NewFiberCaller(t, "route-b"),
				"route-c": tfu.NewFiberCaller(t, "route-c"),
			},
			request:       tfu.NewHTTPFiberRequest(t, http.Header{"X-Region": []string{"region-a"}}, ``),
			fallbacks:     []fiber.Component{},
			expectedError: `route with id "route-a" doesn't exist in the router`,
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			ctx := turingctx.NewTuringContext(context.Background())
			actual, actualFallbacks, err := tt.strategy.SelectRoute(ctx, tt.request, tt.routes)
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

func BenchmarkTrafficSplittingStrategy_SelectRoute(b *testing.B) {
	strategy := &fiberapi.TrafficSplittingStrategy{
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
	}

	routes := map[string]fiber.Component{
		"route-a": tfu.NewFiberCaller(b, "route-a"),
		"route-b": tfu.NewFiberCaller(b, "route-b"),
	}

	ctx := turingctx.NewTuringContext(context.Background())
	payload := `{"service_type": "service-b"}`
	req, _ := http.NewRequest(http.MethodPost, "/predict", bytes.NewReader([]byte(payload)))
	req.Header = http.Header{
		"X-Region": []string{"region-d"},
	}

	fiberReq, _ := fiberHttp.NewHTTPRequest(req)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		primaryRoute, _, _ = strategy.SelectRoute(ctx, fiberReq, routes)
	}
}
