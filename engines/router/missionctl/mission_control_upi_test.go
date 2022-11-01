package missionctl

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
	fiberErrors "github.com/gojek/fiber/errors"
	fibergrpc "github.com/gojek/fiber/grpc"
	fiberHttp "github.com/gojek/fiber/http"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/internal/mocks"
	"github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
)

const (
	port              = 50550
	grpcport1         = 50556
	grpcport2         = 50557
	singleRouteConfig = "testdata/grpc/grpc_router_minimal.yaml"
	twoRouteConfig    = "testdata/grpc/grpc_router_minimal_two_route.yaml"
)

var mockResponse = &upiv1.PredictValuesResponse{
	PredictionResultTable: &upiv1.Table{
		Name:    "table",
		Columns: nil,
		Rows:    nil,
	},
	Metadata: &upiv1.ResponseMetadata{
		PredictionId:   "123",
		ExperimentName: "2",
	},
}

var mockStream = &mocks.ServerTransportStream{}

// TestMain does setup for all test case pre-run
func TestMain(m *testing.M) {

	// this mock stream is required for grpc.SetHeader to have a stream context to work
	mockStream.On("SetHeader", mock.Anything).Return(nil)

	testutils.RunTestUPIServer(port)
	testutils.RunTestUPIServer(grpcport1)
	testutils.RunTestUPIServer(grpcport2)
	os.Exit(m.Run())
}

// TestNewMissionControlUpi tests for the creation of missionControlGrpc and fiberLog configuration
func TestNewMissionControlUpi(t *testing.T) {
	fiberDebugMsg := "Time Taken"
	tests := []struct {
		name          string
		cfgFilePath   string
		fiberDebugLog bool
		expected      MissionControlUPI
		expectedErr   string
	}{
		{
			name:          "ok with no fiber debug",
			cfgFilePath:   singleRouteConfig,
			fiberDebugLog: false,
		},
		{
			name:          "ok with fiber debug",
			cfgFilePath:   twoRouteConfig,
			fiberDebugLog: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zap.DebugLevel)
			logger := zap.New(core)
			log.SetGlobalLogger(logger.Sugar())

			got, err := NewMissionControlUPI(tt.cfgFilePath, tt.fiberDebugLog)
			if err != nil {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				ctx := context.Background()
				ctx = grpc.NewContextWithServerTransportStream(ctx, mockStream)

				res, err := got.Route(ctx,
					&fibergrpc.Request{
						Message: []byte{},
						Proto:   &upiv1.PredictValuesRequest{},
					},
				)

				require.Nil(t, err)
				require.NotNil(t, res)

				logData := logs.FilterMessage(fiberDebugMsg)
				if tt.fiberDebugLog {
					require.NotZero(t, logData.Len())
				} else {
					require.Zero(t, logData.Len())
				}
			}
		})
	}
}

// Test_missionControlUpi_Route mock the response of fiber to test mission control response
func Test_missionControlUpi_Route(t *testing.T) {
	mockResponseByte, err := proto.Marshal(mockResponse)
	require.NoError(t, err)
	tests := []struct {
		name        string
		expected    *upiv1.PredictValuesResponse
		mockReturn  fiber.ResponseQueue
		expectedErr *errors.TuringError
	}{
		{
			name:     "ok",
			expected: mockResponse,
			mockReturn: fiber.NewResponseQueueFromResponses(&fibergrpc.Response{
				Message: mockResponseByte,
			}),
		},
		{
			name: "error wrong response payload type",
			expectedErr: &errors.TuringError{
				Code:    int(codes.Internal),
				Message: "did not get back a valid response from the fiberHandler",
			},
			mockReturn: fiber.NewResponseQueueFromResponses(),
		},
		{
			name: "error - fiber router error response",
			expectedErr: &errors.TuringError{
				Code:    int(codes.Internal),
				Message: "{\n  \"code\": 13,\n  \"error\": \"fiber: request cannot be completed: err\"\n}",
			},
			mockReturn: fiber.NewResponseQueueFromResponses(
				fiber.NewErrorResponse(
					fiberErrors.FiberError{
						Code:    int(codes.Internal),
						Message: "fiber: request cannot be completed: err",
					})),
		},
		{
			name: "error non proto payload",
			expectedErr: &errors.TuringError{
				Code:    int(codes.Internal),
				Message: "unable to parse fiber response into grpc response",
			},
			mockReturn: fiber.NewResponseQueueFromResponses(fiberHttp.NewHTTPResponse(
				&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("dummy res"))),
				})),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFiberRouter := &mocks.Component{}
			mockFiberRouter.On("Dispatch", mock.Anything, mock.Anything).Return(tt.mockReturn, nil)
			mc := missionControlUpi{fiberRouter: mockFiberRouter}
			ctx := context.Background()
			ctx = grpc.NewContextWithServerTransportStream(ctx, mockStream)

			var err error
			got, err := mc.Route(ctx, &fibergrpc.Request{})
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				responseProto := &upiv1.PredictValuesResponse{}
				require.NoError(t, proto.Unmarshal(got.Payload(), responseProto))
				require.True(t, proto.Equal(responseProto, tt.expected), "response not equal to expected")
			}
		})
	}
}

// Test_missionControlUpi_Route will send request to the test server which will duplicate the request table in response
// this test will check for the correctness of byte marshaling
func Test_missionControlUpi_Route_Integration(t *testing.T) {
	smallRequest := testutils.GenerateUPIRequest(5, 5)
	smallRequestByte, err := proto.Marshal(smallRequest)
	require.NoError(t, err)
	smallRequestExpected := upiv1.PredictValuesResponse{PredictionResultTable: smallRequest.PredictionTable}

	largeRequest := testutils.GenerateUPIRequest(500, 500)
	largeRequestByte, err := proto.Marshal(largeRequest)
	require.NoError(t, err)
	largeRequestExpected := upiv1.PredictValuesResponse{PredictionResultTable: largeRequest.PredictionTable}

	tests := []struct {
		name           string
		request        fiber.Request
		compareAgainst *upiv1.PredictValuesResponse
		expectedEqual  bool
	}{
		{
			name: "small request",
			request: &fibergrpc.Request{
				Message: smallRequestByte,
				Proto:   smallRequest,
			},
			compareAgainst: &smallRequestExpected,
			expectedEqual:  true,
		},
		{
			name: "large request",
			request: &fibergrpc.Request{
				Message: largeRequestByte,
				Proto:   largeRequest,
			},
			compareAgainst: &largeRequestExpected,
			expectedEqual:  true,
		},
		{
			name: "large request",
			request: &fibergrpc.Request{
				Message: largeRequestByte,
				Proto:   largeRequest,
			},
			compareAgainst: &smallRequestExpected,
			expectedEqual:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc, err := NewMissionControlUPI(twoRouteConfig, false)
			require.NoError(t, err)
			ctx := context.Background()
			ctx = grpc.NewContextWithServerTransportStream(ctx, mockStream)
			got, err := mc.Route(ctx, tt.request)
			require.Nil(t, err)
			responseProto := &upiv1.PredictValuesResponse{}
			require.NoError(t, proto.Unmarshal(got.Payload(), responseProto))
			require.Equal(t, tt.expectedEqual, proto.Equal(responseProto, tt.compareAgainst), "Comparison result not expected")
		})
	}
}
