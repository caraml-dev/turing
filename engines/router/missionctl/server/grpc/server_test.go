package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/internal/mocks"
	"github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestUPIServer_PredictValues(t *testing.T) {

	tests := []struct {
		name        string
		request     *upiv1.PredictValuesRequest
		expected    *upiv1.PredictValuesResponse
		expectedErr *errors.TuringError
		mockReturn  func() (*upiv1.PredictValuesResponse, *errors.TuringError)
	}{
		{
			name:        "ok",
			request:     nil,
			expected:    testutils.GetDefaultMockResponse(),
			expectedErr: nil,
			mockReturn: func() (*upiv1.PredictValuesResponse, *errors.TuringError) {
				return testutils.GetDefaultMockResponse(), nil
			},
		},
		{
			name:    "error",
			request: nil,
			expectedErr: &errors.TuringError{
				Code:    14,
				Message: "did not get back a valid response from the fiberHandler",
			},
			mockReturn: func() (*upiv1.PredictValuesResponse, *errors.TuringError) {
				return nil, &errors.TuringError{
					Code:    14,
					Message: "did not get back a valid response from the fiberHandler",
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMc := &mocks.MissionControlUPI{}
			mockMc.On("Route", mock.Anything, mock.Anything).Return(tt.mockReturn())

			upiServer := NewUPIServer(mockMc, 50400)
			ctx := context.Background()
			resp, err := upiServer.PredictValues(ctx, tt.request)
			if tt.expectedErr != nil {
				require.Equal(t, err, tt.expectedErr)
			} else {
				require.Equal(t, resp, tt.expected)
			}
		})
	}
}

func TestNewUpiServer(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)
	log.SetGlobalLogger(logger.Sugar())

	mockMc := &mocks.MissionControlUPI{}
	upiServer := NewUPIServer(mockMc, 50400)
	go upiServer.Run()

	// Wait for server to run and check that there are no error logs
	time.Sleep(2 * time.Second)
	require.Zero(t, logs.Len())

	upiServer2 := NewUPIServer(mockMc, 50400)
	go upiServer2.Run()

	// Wait for server to run and expected error logs due to same port
	time.Sleep(2 * time.Second)
	require.Equal(t, 1, logs.Len())
	require.Contains(t, logs.All()[0].Message, "Failed to listen on port")
}
