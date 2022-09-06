// Package errors provides definitions for some common errors that may be encountered
// through the course of the turning mission control's functions. These errors may be
// eventually mapped to HTTP error codes, when being returned as a response.
package errors

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// ErrorType captures some common error types
type ErrorType uint

const (
	// Unknown error type is used for all generic errors
	Unknown = ErrorType(iota)
	// BadInput is used when any function encounters bad/incomplete input
	BadInput
	// BadResponse is used when a go method or an external service returns bad data
	BadResponse
	// BadConfig is used when an initialization step fails due to missing / invalid config values
	BadConfig
	// NotFound is used when a resource cannot be located
	NotFound
	// TimeOut is used when a request / go routine times out
	TimeOut
)

type turingError struct {
	errorType ErrorType
	errorInfo error
}

// Error satisfies error interface
func (error turingError) Error() string {
	return error.errorInfo.Error()
}

// Newf creates a new turingError of the specified type, with formatted message
func Newf(et ErrorType, msg string, args ...interface{}) error {
	err := fmt.Errorf(msg, args...)
	return turingError{errorType: et, errorInfo: err}
}

// Wrapf method creates a new wrapped turingError with formatted message,
// of the specified error type
func (errorType ErrorType) Wrapf(err error, msg string, args ...interface{}) error {
	return turingError{errorType: errorType, errorInfo: errors.Wrapf(err, msg, args...)}
}

// Wrapf creates a new wrapped turingError with formatted message
func Wrapf(err error, msg string, args ...interface{}) error {
	newErr := errors.Wrapf(err, msg, args...)
	// Try casting the inner error to turingErr
	if turingErr, ok := err.(turingError); ok {
		return turingError{
			errorType: turingErr.errorType,
			errorInfo: newErr,
		}
	}
	return turingError{errorType: Unknown, errorInfo: newErr}
}

// GetType returns the error type
func GetType(err error) ErrorType {
	if turingErr, ok := err.(turingError); ok {
		return turingErr.errorType
	}
	return Unknown
}

// GetHTTPErrorCode maps the ErrorType to http status codes and returns it
func GetHTTPErrorCode(err error) int {
	var code int

	// Get ErrorType if its turingError else set to default
	et := GetType(err)

	switch et {
	case BadInput:
		code = http.StatusBadRequest
	case BadResponse:
		code = http.StatusBadGateway
	case NotFound:
		code = http.StatusNotFound
	default:
		code = http.StatusInternalServerError
	}
	return code
}

// HTTPError associates an error message with a HTTP status code.
type HTTPError struct {
	Code    int
	Message string
}

// Error satisfies the error interface
func (e *HTTPError) Error() string {
	return e.Message
}

// NewHTTPError creates an error with a HTTP Status code
func NewHTTPError(err error, code ...int) *HTTPError {
	var httpErrCode int
	if len(code) > 0 {
		httpErrCode = code[0]
	}

	// If code unknown, create a status code from the error
	if http.StatusText(httpErrCode) == "" {
		httpErrCode = GetHTTPErrorCode(err)
	}
	return &HTTPError{
		Code:    httpErrCode,
		Message: err.Error(),
	}
}
