package upi

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/internal/mocks"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

var mockResponse = &upiv1.PredictValuesResponse{
	PredictionResultTable: &upiv1.Table{
		Name:    "table",
		Columns: nil,
		Rows:    nil,
	},
	Metadata: &upiv1.ResponseMetadata{
		PredictionId: "123",
		ExperimentId: "2",
	},
}

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
			expected:    mockResponse,
			expectedErr: nil,
			mockReturn: func() (*upiv1.PredictValuesResponse, *errors.TuringError) {
				return mockResponse, nil
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

			upiServer := NewUPIServer(mockMc)
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

	port := 50560
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Glob().Panicf("Failed to listen on port: %v", port)
	}

	mockMc := &mocks.MissionControlUPI{}
	upiServer := NewUPIServer(mockMc)
	go upiServer.Run(l)

	// Wait for server to run and check that there are no error logs
	time.Sleep(2 * time.Second)
	require.Zero(t, logs.Len())
}
