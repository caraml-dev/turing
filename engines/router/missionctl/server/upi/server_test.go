package upi

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
	fiberGrpc "github.com/gojek/fiber/grpc"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/internal/mocks"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
)

var mockResponse = &upiv1.PredictValuesResponse{
	PredictionResultTable: &upiv1.Table{
		Name:    "table",
		Columns: nil,
		Rows:    nil,
	},
}

func TestUPIServer_PredictValues(t *testing.T) {

	responseByte, err := proto.Marshal(mockResponse)
	require.NoError(t, err)
	tests := []struct {
		name        string
		request     *upiv1.PredictValuesRequest
		expected    *upiv1.PredictValuesResponse
		expectedErr *errors.TuringError
		mockReturn  func() (fiber.Response, *errors.TuringError)
	}{
		{
			name:        "ok",
			request:     &upiv1.PredictValuesRequest{},
			expected:    mockResponse,
			expectedErr: nil,
			mockReturn: func() (fiber.Response, *errors.TuringError) {
				return &fiberGrpc.Response{
					Message: responseByte,
				}, nil
			},
		},
		{
			name:    "error",
			request: &upiv1.PredictValuesRequest{},
			expectedErr: &errors.TuringError{
				Code:    int(codes.Unavailable),
				Message: "did not get back a valid response from the fiberHandler",
			},
			mockReturn: func() (fiber.Response, *errors.TuringError) {
				return nil, &errors.TuringError{
					Code:    int(codes.Unavailable),
					Message: "did not get back a valid response from the fiberHandler",
				}
			},
		},
		{
			name:    "error wrong response payload type",
			request: &upiv1.PredictValuesRequest{},
			expectedErr: &errors.TuringError{
				Code:    int(codes.Internal),
				Message: "unable to unmarshal into expected response proto",
			},
			mockReturn: func() (fiber.Response, *errors.TuringError) {
				return &fiberGrpc.Response{
					Message: []byte("test"),
				}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMc := &mocks.MissionControlUPI{}
			mockMc.On("Route", mock.Anything, mock.Anything).Return(tt.mockReturn()).Run(func(args mock.Arguments) {
				fiberRequest, ok := args.Get(1).(*fiberGrpc.Request)
				require.True(t, ok, "not fiber grpc request")

				upiReq, ok := fiberRequest.Proto.(*upiv1.PredictValuesRequest)
				require.True(t, ok, "not upi request")

				require.True(t, proto.Equal(upiReq.PredictionTable, tt.request.PredictionTable), "invalid prediction table")
				require.NotEmpty(t, upiReq.Metadata.PredictionId, "prediction id is empty")
				require.NotEmpty(t, upiReq.Metadata.RequestTimestamp, "request timestamp is empty")
			})

			upiServer := NewUPIServer(mockMc)
			ctx := context.Background()
			resp, err := upiServer.PredictValues(ctx, tt.request)
			if tt.expectedErr != nil {
				require.Equal(t, err, tt.expectedErr)
				return
			}

			require.True(t, proto.Equal(resp.PredictionResultTable, tt.expected.PredictionResultTable),
				"response not equal to expected")
			require.NotEmpty(t, resp.Metadata.PredictionId, "prediction id is empty")
		})
	}
}

func TestNewUpiServer(t *testing.T) {
	_, logs := observer.New(zap.ErrorLevel)

	port := 50560
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Glob().Panicf("Failed to listen on port %v: %s", port, err)
	}

	mockMc := &mocks.MissionControlUPI{}
	upiServer := NewUPIServer(mockMc)
	go upiServer.Run(l)

	// Wait for server to run and check that there are no error logs
	time.Sleep(2 * time.Second)
	require.Zero(t, logs.Len())
}
