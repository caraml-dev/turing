package fiberapi

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	fibererror "github.com/gojek/fiber/errors"
	"github.com/stretchr/testify/assert"
)

func TestCreateFiberError(t *testing.T) {
	tests := map[string]struct {
		err         error
		expectedErr fibererror.HTTPError
	}{
		"generic": {
			err: errors.Newf(errors.Unknown, "Test error"),
			expectedErr: fibererror.HTTPError{
				Code:    500,
				Message: "Test error",
			},
		},
		"bad input": {
			err: errors.Newf(errors.BadInput, "Input error"),
			expectedErr: fibererror.HTTPError{
				Code:    400,
				Message: "Input error",
			},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, data.expectedErr, createFiberError(data.err))
		})
	}
}

func TestCreateFiberRequestHandler(t *testing.T) {
	handler, err := CreateFiberRequestHandler(
		filepath.Join("..", "testdata", "nop_default_router.yaml"),
		time.Second*2,
		true,
	)

	assert.NoError(t, err)
	assert.Equal(t, "eager-router", handler.Component.ID())
	assert.Equal(t, "Combiner", string(handler.Component.Kind()))
}
