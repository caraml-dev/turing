package experiment_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gojek/turing/engines/experiment"
	"github.com/stretchr/testify/assert"
)

func TestGetFieldSource(t *testing.T) {
	// header
	fieldSrc, err := experiment.GetFieldSource("header")
	assert.Equal(t, experiment.HeaderFieldSource, fieldSrc)
	assert.NoError(t, err)
	// payload
	fieldSrc, err = experiment.GetFieldSource("payload")
	assert.Equal(t, experiment.PayloadFieldSource, fieldSrc)
	assert.NoError(t, err)
	// unknown
	_, err = experiment.GetFieldSource("test")
	assert.Error(t, err)
}

func TestUnmarshalJSONFieldSource(t *testing.T) {
	var fieldSrc experiment.FieldSource
	// success
	err := json.Unmarshal([]byte(`"header"`), &fieldSrc)
	assert.Equal(t, experiment.HeaderFieldSource, fieldSrc)
	assert.NoError(t, err)
	// unknown string
	err = json.Unmarshal([]byte(`"test"`), &fieldSrc)
	assert.Error(t, err)
	// invalid data
	err = json.Unmarshal([]byte(`0`), &fieldSrc)
	assert.Error(t, err)
}

func TestGetValueFromRequest(t *testing.T) {
	tests := map[string]struct {
		field    string
		fieldSrc experiment.FieldSource
		header   http.Header
		body     []byte
		expected string
		err      string
	}{
		"success | header": {
			field:    "CustomerID",
			fieldSrc: experiment.HeaderFieldSource,
			header: func() http.Header {
				header := http.Header{}
				header.Set("CustomerID", "123")
				return header
			}(),
			expected: "123",
		},
		"success | nested payload": {
			field:    "customer.id",
			fieldSrc: experiment.PayloadFieldSource,
			body:     []byte(`{"customer": {"id": "test_customer"}}`),
			expected: "test_customer",
		},
		"success | payload integer field": {
			field:    "customer.id",
			fieldSrc: experiment.PayloadFieldSource,
			body:     []byte(`{"customer": {"id": 42}}`),
			expected: "42",
		},
		"success | payload bool field": {
			field:    "is_premium_customer",
			fieldSrc: experiment.PayloadFieldSource,
			body:     []byte(`{"is_premium_customer": true}`),
			expected: "true",
		},
		"success | payload null field": {
			field:    "session_id",
			fieldSrc: experiment.PayloadFieldSource,
			body:     []byte(`{"session_id": null}`),
			expected: "",
		},
		"failure | header": {
			field:    "CustomerID",
			fieldSrc: experiment.HeaderFieldSource,
			header: func() http.Header {
				header := http.Header{}
				header.Set("SessionID", "123")
				return header
			}(),
			err: "Field CustomerID not found in the request header",
		},
		"failure | payload": {
			field:    "customer_id",
			fieldSrc: experiment.PayloadFieldSource,
			body:     []byte(`{"customer": {"id": "test_customer"}}`),
			err:      "Field customer_id not found in the request payload: Key path not found",
		},
		"failure | payload unsupported type": {
			field:    "customer",
			fieldSrc: experiment.PayloadFieldSource,
			body:     []byte(`{"customer": {"id": 42, "email": "test@test.com"}`),
			err:      "Field customer can not be parsed as string value, unsupported type: object",
		},
		"failure | unknown source": {
			field:    "CustomerID",
			fieldSrc: experiment.FieldSource("unknown"),
			err:      "Unrecognized field source unknown",
		},
		"failure | malformed JSON": {
			field:    "customer.id",
			fieldSrc: experiment.PayloadFieldSource,
			body:     []byte(`{"customer: {}"id"`),
			err:      "Field customer.id not found in the request payload: Key path not found",
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			val, err := experiment.GetValueFromRequest(data.header, data.body, data.fieldSrc, data.field)
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
