package fiberapi

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	fiberErrors "github.com/gojek/fiber/errors"
	fiberProtocol "github.com/gojek/fiber/protocol"
	"github.com/stretchr/testify/assert"
)

func TestCreateFiberError(t *testing.T) {
	tests := map[string]struct {
		err         error
		expectedErr fiberErrors.FiberError
	}{
		"generic": {
			err: errors.Newf(errors.Unknown, "Test error"),
			expectedErr: fiberErrors.FiberError{
				Code:    500,
				Message: "Test error",
			},
		},
		"bad input": {
			err: errors.Newf(errors.BadInput, "Input error"),
			expectedErr: fiberErrors.FiberError{
				Code:    400,
				Message: "Input error",
			},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, data.expectedErr, createFiberError(data.err, fiberProtocol.HTTP))
		})
	}
}

func TestCreateFiberRequestHandler(t *testing.T) {
	router, err := CreateFiberRouterFromConfig(filepath.Join("..", "testdata", "nop_default_router.yaml"), true)
	assert.NoError(t, err)
	assert.Equal(t, "eager-router", router.ID())
	assert.Equal(t, "Combiner", string(router.Kind()))

	handler := CreateFiberRequestHandler(router, time.Second*2)

	assert.Equal(t, "eager-router", handler.Component.ID())
	assert.Equal(t, "Combiner", string(handler.Component.Kind()))
}
