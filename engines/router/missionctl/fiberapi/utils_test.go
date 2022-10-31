package fiberapi

import (
	"testing"

	fiberErrors "github.com/gojek/fiber/errors"
	fiberProtocol "github.com/gojek/fiber/protocol"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
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
