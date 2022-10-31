package router_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	fiberHttp "github.com/gojek/fiber/http"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
	"github.com/caraml-dev/turing/engines/router"
)

type operatorSerializationTestCase struct {
	operator      router.RuleConditionOperator
	serialized    string
	expectedError string
}

func TestRuleConditionOperator_UnmarshalJSON(t *testing.T) {
	suite := map[string]operatorSerializationTestCase{
		"success": {
			operator:   router.InConditionOperator,
			serialized: `"in"`,
		},
		"failure | unknown operator": {
			serialized:    `"some-operator"`,
			expectedError: "unknown operator: some-operator",
		},
		"failure | invalid type": {
			serialized:    "42",
			expectedError: "json: cannot unmarshal number into Go value of type string",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			var actual router.RuleConditionOperator
			err := json.Unmarshal([]byte(tt.serialized), &actual)

			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.operator, actual)
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

func TestRuleConditionOperator_MarshalJSON(t *testing.T) {
	suite := map[string]operatorSerializationTestCase{
		"success": {
			operator:   router.InConditionOperator,
			serialized: `"in"`,
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			actual, err := json.Marshal(tt.operator)

			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, []byte(tt.serialized), actual)
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

type inOperatorTestCase struct {
	left          interface{}
	right         interface{}
	expected      bool
	expectedError string
}

func TestInConditionOperator_Test(t *testing.T) {
	suite := map[string]inOperatorTestCase{
		"success | ok | string": {
			left:     "test-string",
			right:    []string{"first", "another", "test-string"},
			expected: true,
		},
		"success | nok | string": {
			left:     "test-string",
			right:    []string{"foo", "bar", "foobar"},
			expected: false,
		},
		"success | nok | different types": {
			left:     "42",
			right:    []int{42, 19, 84},
			expected: false,
		},
		"success | ok | int": {
			left:     42,
			right:    []int{42, 19, 84},
			expected: true,
		},
		"failure | incompatible type": {
			left:          "test-string",
			right:         "test-string",
			expectedError: "invalid type of right argument: slice is expected",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			actual, err := router.InConditionOperator.Test(tt.left, tt.right)

			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

type mockOperator struct {
	mock.Mock
}

func (o *mockOperator) String() string {
	return "mock-operator"
}

func (o *mockOperator) Test(left interface{}, right interface{}) (bool, error) {
	args := o.Called(left, right)
	return args.Bool(0), args.Error(1)
}

func makeMockOperator(left, right interface{}, result bool, err error) router.RuleConditionOperator {
	operator := new(mockOperator)
	operator.
		On("Test", left, right).
		Return(result, err)

	return router.RuleConditionOperator{Operator: operator}
}

type trafficRuleConditionTestCase struct {
	condition     *router.TrafficRuleCondition
	header        http.Header
	payload       string
	expected      bool
	expectedError string
}

func TestTrafficRuleCondition_TestRequest(t *testing.T) {
	suite := map[string]trafficRuleConditionTestCase{
		"success | header": {
			condition: &router.TrafficRuleCondition{
				FieldSource: request.HeaderFieldSource,
				Field:       "Content-Type",
				Operator: makeMockOperator(
					"application/json", []string{"application/json"}, true, nil),
				Values: []string{"application/json"},
			},
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			expected: true,
		},
		"success | header not in": {
			condition: &router.TrafficRuleCondition{
				FieldSource: request.HeaderFieldSource,
				Field:       "Content-Type",
				Operator: makeMockOperator(
					"application/xml", []string{"application/json", "text/plain"}, false, nil),
				Values: []string{"application/json", "text/plain"},
			},
			header: http.Header{
				"Content-Type": []string{"application/xml"},
			},
			expected: false,
		},
		"success | payload": {
			condition: &router.TrafficRuleCondition{
				FieldSource: request.PayloadFieldSource,
				Field:       "parent_field.nested",
				Operator: makeMockOperator(
					"foo", []string{"foo", "bar"}, true, nil),
				Values: []string{"foo", "bar"},
			},
			payload:  `{"parent_field": {"nested": "foo"}}`,
			expected: true,
		},
		"failure | header not found": {
			condition: &router.TrafficRuleCondition{
				FieldSource: request.HeaderFieldSource,
				Field:       "Session-ID",
				Operator:    makeMockOperator(nil, nil, true, nil),
				Values:      []string{"foo", "bar"},
			},
			header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			expectedError: "Field Session-ID not found in the request header",
		},
		"failure | key not found": {
			condition: &router.TrafficRuleCondition{
				FieldSource: request.PayloadFieldSource,
				Field:       "parent_field.bar",
				Operator:    makeMockOperator(nil, nil, true, nil),
				Values:      []string{"foo", "bar"},
			},
			payload:       `{}`,
			expectedError: "Field parent_field.bar not found in the request payload: Key path not found",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			r, _ := fiberHttp.NewHTTPRequest(&http.Request{
				Header: tt.header,
				Body:   ioutil.NopCloser(strings.NewReader(tt.payload)),
			})

			actual, err := tt.condition.TestRequest(r)
			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

type upiTrafficRuleConditionTestCase struct {
	condition     *router.TrafficRuleCondition
	header        metadata.MD
	payload       *upiv1.PredictValuesRequest
	expected      bool
	expectedError string
}

func TestTrafficRuleCondition_TestUPIRequest(t *testing.T) {
	suite := map[string]upiTrafficRuleConditionTestCase{
		"success | header": {
			condition: &router.TrafficRuleCondition{
				FieldSource: request.HeaderFieldSource,
				Field:       "foo",
				Operator: makeMockOperator(
					"bar", []string{"bar"}, true, nil),
				Values: []string{"bar"},
			},
			header: metadata.MD{
				"foo": []string{"bar"},
			},
			expected: true,
		},
		"success | header not in": {
			condition: &router.TrafficRuleCondition{
				FieldSource: request.HeaderFieldSource,
				Field:       "header",
				Operator: makeMockOperator(
					"actual-value", []string{"exp-value-1", "exp-value-2"}, false, nil),
				Values: []string{"exp-value-1", "exp-value-2"},
			},
			header: metadata.MD{
				"header": []string{"actual-value"},
			},
			expected: false,
		},
		"success | prediction context": {
			condition: &router.TrafficRuleCondition{
				FieldSource: request.PredictionContextSource,
				Field:       "my-variable",
				Operator: makeMockOperator(
					"foo", []string{"foo", "bar"}, true, nil),
				Values: []string{"foo", "bar"},
			},
			payload: &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "my-variable",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "foo",
					},
				},
			},
			expected: true,
		},
		"failure | header not found": {
			condition: &router.TrafficRuleCondition{
				FieldSource: request.HeaderFieldSource,
				Field:       "Session-ID",
				Operator:    makeMockOperator(nil, nil, true, nil),
				Values:      []string{"foo", "bar"},
			},
			header: metadata.MD{
				"Content-Type": []string{"application/json"},
			},
			expectedError: "Field Session-ID not found in the request header",
		},
		"failure | variable not found": {
			condition: &router.TrafficRuleCondition{
				FieldSource: request.PredictionContextSource,
				Field:       "missing-variable",
				Operator:    makeMockOperator(nil, nil, true, nil),
				Values:      []string{"foo", "bar"},
			},
			payload: &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "my-variable",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "foo",
					},
				},
			},
			expectedError: "Variable missing-variable not found in the prediction context",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			actual, err := tt.condition.TestUPIRequest(tt.payload, tt.header)
			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
