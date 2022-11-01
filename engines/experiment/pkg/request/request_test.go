package request_test

import (
	"net/http"
	"testing"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
)

func TestGetFieldSource(t *testing.T) {
	// header
	fieldSrc, err := request.GetFieldSource("header")
	assert.Equal(t, request.HeaderFieldSource, fieldSrc)
	assert.NoError(t, err)
	// payload
	fieldSrc, err = request.GetFieldSource("payload")
	assert.Equal(t, request.PayloadFieldSource, fieldSrc)
	assert.NoError(t, err)
	// unknown
	_, err = request.GetFieldSource("test")
	assert.Error(t, err)
}

func TestGetValueFromHTTPRequest(t *testing.T) {
	tests := map[string]struct {
		field    string
		fieldSrc request.FieldSource
		header   http.Header
		body     []byte
		expected string
		err      string
	}{
		"success | header": {
			field:    "CustomerID",
			fieldSrc: request.HeaderFieldSource,
			header: func() http.Header {
				header := http.Header{}
				header.Set("CustomerID", "123")
				return header
			}(),
			expected: "123",
		},
		"success | nested payload": {
			field:    "customer.id",
			fieldSrc: request.PayloadFieldSource,
			body:     []byte(`{"customer": {"id": "test_customer"}}`),
			expected: "test_customer",
		},
		"success | payload integer field": {
			field:    "customer.id",
			fieldSrc: request.PayloadFieldSource,
			body:     []byte(`{"customer": {"id": 42}}`),
			expected: "42",
		},
		"success | payload bool field": {
			field:    "is_premium_customer",
			fieldSrc: request.PayloadFieldSource,
			body:     []byte(`{"is_premium_customer": true}`),
			expected: "true",
		},
		"success | payload null field": {
			field:    "session_id",
			fieldSrc: request.PayloadFieldSource,
			body:     []byte(`{"session_id": null}`),
			expected: "",
		},
		"failure | header": {
			field:    "CustomerID",
			fieldSrc: request.HeaderFieldSource,
			header: func() http.Header {
				header := http.Header{}
				header.Set("SessionID", "123")
				return header
			}(),
			err: "Field CustomerID not found in the request header",
		},
		"failure | payload": {
			field:    "customer_id",
			fieldSrc: request.PayloadFieldSource,
			body:     []byte(`{"customer": {"id": "test_customer"}}`),
			err:      "Field customer_id not found in the request payload: Key path not found",
		},
		"failure | payload unsupported type": {
			field:    "customer",
			fieldSrc: request.PayloadFieldSource,
			body:     []byte(`{"customer": {"id": 42, "email": "test@test.com"}`),
			err:      "Field customer can not be parsed as string value, unsupported type: object",
		},
		"failure | unknown source": {
			field:    "CustomerID",
			fieldSrc: request.FieldSource("unknown"),
			err:      "Unrecognized field source unknown",
		},
		"failure | malformed JSON": {
			field:    "customer.id",
			fieldSrc: request.PayloadFieldSource,
			body:     []byte(`{"customer: {}"id"`),
			err:      "Field customer.id not found in the request payload: Key path not found",
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			val, err := request.GetValueFromHTTPRequest(data.header, data.body, data.fieldSrc, data.field)
			assert.Equal(t, data.expected, val)
			// Check error
			if data.err != "" {
				if err == nil {
					t.Errorf("Expected error but got nil")
					t.FailNow()
				}
				assert.Equal(t, data.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetValueFromUPIRequest(t *testing.T) {
	tests := map[string]struct {
		field    string
		fieldSrc request.FieldSource
		header   metadata.MD
		body     *upiv1.PredictValuesRequest
		expected string
		err      string
	}{
		"success | header": {
			field:    "customer-id",
			fieldSrc: request.HeaderFieldSource,
			header: metadata.MD{
				"customer-id": []string{"123"},
			},
			expected: "123",
		},
		"success | nested payload": {
			field:    "customer-id",
			fieldSrc: request.PredictionContextSource,
			body: &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "foo",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "bar",
					},
					{
						Name:        "customer-id",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "test_customer",
					},
				},
			},
			expected: "test_customer",
		},
		"success | payload integer field": {
			field:    "customer-id",
			fieldSrc: request.PredictionContextSource,
			body: &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "foo",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "bar",
					},
					{
						Name:         "customer-id",
						Type:         upiv1.Type_TYPE_INTEGER,
						IntegerValue: 1234,
					},
				},
			},
			expected: "1234",
		},
		"failure | header not found": {
			field:    "missing-header",
			fieldSrc: request.HeaderFieldSource,
			header: metadata.MD{
				"customer-id": []string{"123"},
			},
			err: "Field missing-header not found in the request header",
		},
		"failure | variable not found": {
			field:    "missing-variable",
			fieldSrc: request.PredictionContextSource,
			body: &upiv1.PredictValuesRequest{
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "foo",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "bar",
					},
					{
						Name:        "customer-id",
						Type:        upiv1.Type_TYPE_INTEGER,
						StringValue: "1234",
					},
				},
			},
			err: "Variable missing-variable not found in the prediction context",
		},
		"failure | unknown source": {
			field:    "CustomerID",
			fieldSrc: request.PayloadFieldSource,
			err:      "Unrecognized field source payload",
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			val, err := request.GetValueFromUPIRequest(data.header, data.body, data.fieldSrc, data.field)
			assert.Equal(t, data.expected, val)
			// Check error
			if data.err != "" {
				if err == nil {
					t.Errorf("Expected error but got nil")
					t.FailNow()
				}
				assert.Equal(t, data.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
