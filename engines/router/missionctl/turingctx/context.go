package turingctx

import (
	"context"

	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/google/uuid"
)

// Set up context keys for logging with context
type ctxKeyType int

const (
	turingReqIDKey ctxKeyType = iota
	// TuringTreatmentChannelKey is used to store a channel to send experiment treatment
	TuringTreatmentChannelKey
)

// NewTuringContext returns a context which holds additional data pertaining
// to the unique turing request.
func NewTuringContext(parent context.Context) context.Context {
	reqID := uuid.New()
	return context.WithValue(parent, turingReqIDKey, reqID.String())
}

// GetRequestID returns the request id from the input context
func GetRequestID(ctx context.Context) (string, error) {
	if ctxValue, ok := ctx.Value(turingReqIDKey).(string); ok {
		return ctxValue, nil
	}
	return "", errors.Newf(errors.Unknown, "Request ID not found in the context")
}

// GetKeyValsFromContext retrieves all the possible turing related key-value(s) from the
// input context as a interface{} slice.
func GetKeyValsFromContext(ctx context.Context) []interface{} {
	props := []interface{}{}

	// Turing request id
	if ctxValue, ok := ctx.Value(turingReqIDKey).(string); ok {
		props = append(props, "turing_req_id", ctxValue)
	}

	return props
}
