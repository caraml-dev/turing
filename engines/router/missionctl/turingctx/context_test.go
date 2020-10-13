package turingctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTuringContext(t *testing.T) {
	turCtx := NewTuringContext(context.Background())
	// Check that a turing request id is present
	reqID := turCtx.Value(turingReqIDKey)
	assert.NotEmpty(t, reqID)
}

func TestGetRequestID(t *testing.T) {
	testReqID := "test-uuid"
	turCtx := context.WithValue(context.Background(), turingReqIDKey, testReqID)
	reqID, err := GetRequestID(turCtx)
	// Check that the expected turing request is returned and no error
	assert.Equal(t, testReqID, reqID)
	assert.Equal(t, nil, err)
}

func TestGetMissingRequestID(t *testing.T) {
	turCtx := context.Background()
	reqID, err := GetRequestID(turCtx)
	// Check that the request id returned is empty, and we have an error
	assert.Empty(t, reqID)
	assert.Error(t, err)
}

func TestGetKeyValsFromContext(t *testing.T) {
	// Define expected values
	testReqID := "test-uuid"
	expectedSlice := []interface{}{"turing_req_id", testReqID}
	// Create context, get key-vals
	turCtx := context.WithValue(context.Background(), turingReqIDKey, testReqID)
	props := GetKeyValsFromContext(turCtx)
	// Test that the two slices are equal
	assert.Equal(t, expectedSlice, props)
}
