package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gojek/fiber/protocol"
	"github.com/stretchr/testify/assert"
)

type testSuiteError struct {
	codes        []int
	expectedCode int
}

func TestNewfGetType(t *testing.T) {
	err := Newf(BadInput, "New %s error", "test")

	// Test that the error object has been created as expected
	assert.Equal(t, "New test error", err.Error())
	assert.Equal(t, BadInput, GetType(err))
}

func TestWrapfKnownOuterError(t *testing.T) {
	err := errors.New("Inner error")
	wrappedErr := BadInput.Wrapf(err, "Outer error %s", "message")

	assert.Equal(t, BadInput, GetType(wrappedErr))
	assert.EqualError(t, wrappedErr, "Outer error message: Inner error")
}

func TestWrapfKnownNestedError(t *testing.T) {
	err := Newf(BadInput, "Inner error")
	wrappedErr := Wrapf(err, "Outer error %s", "message")

	assert.Equal(t, BadInput, GetType(wrappedErr))
	assert.EqualError(t, wrappedErr, "Outer error message: Inner error")
}

func TestWrapfUnknownError(t *testing.T) {
	err := errors.New("Inner error")
	wrappedErr := Wrapf(err, "Outer error %s", "message")

	assert.Equal(t, Unknown, GetType(wrappedErr))
	assert.EqualError(t, wrappedErr, "Outer error message: Inner error")
}

func TestGetErrorTypeUnknown(t *testing.T) {
	err := errors.New("Test error")
	assert.Equal(t, Unknown, GetType(err))
}

func TestGetHTTPErrorCode(t *testing.T) {
	testErrorSuite := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "Generic Error",
			err:          errors.New("Test error"),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "BadInput",
			err:          Newf(BadInput, ""),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "BadResponse",
			err:          Newf(BadResponse, ""),
			expectedCode: http.StatusBadGateway,
		},
		{
			name:         "NotFound",
			err:          Newf(NotFound, ""),
			expectedCode: http.StatusNotFound,
		},
	}

	for _, data := range testErrorSuite {
		t.Run(data.name, func(t *testing.T) {
			assert.Equal(t, data.expectedCode, GetErrorCode(data.err, protocol.HTTP))
		})
	}
}

func TestNewHTTPErrorMessage(t *testing.T) {
	message := "Test Error Message"
	err := fmt.Errorf(message)
	httpErr := NewTuringError(err, protocol.HTTP)
	assert.Equal(t, message, httpErr.Error())
}

func TestNewHTTPErrorStatus(t *testing.T) {
	tests := map[string]testSuiteError{
		"no HTTP status": {
			codes:        []int{},
			expectedCode: http.StatusInternalServerError,
		},
		"invalid HTTP status": {
			codes:        []int{-1},
			expectedCode: http.StatusInternalServerError,
		},
		"valid HTTP status": {
			codes:        []int{400},
			expectedCode: 400,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Create new error
			message := "Test Error"
			err := fmt.Errorf(message)
			// Create new HTTP error
			httpErr := NewTuringError(err, protocol.HTTP, data.codes...)
			// Validate
			assert.Equal(t, data.expectedCode, httpErr.Code)
			assert.Equal(t, message, httpErr.Message)
		})
	}
}
