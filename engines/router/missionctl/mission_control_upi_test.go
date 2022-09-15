package missionctl

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/internal/mocks"
	"github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
	fiberErrors "github.com/gojek/fiber/errors"
	fibergrpc "github.com/gojek/fiber/grpc"
	fiberhttp "github.com/gojek/fiber/http"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	port            = 50550
	grpcport1       = 50556
	grpcport2       = 50557
	benchmarkConfig = "testdata/grpc/grpc_router_minimal.yaml"
	twoRouteConfig  = "testdata/grpc/grpc_router_minimal_two_route.yaml"
)

var benchMarkUpiResp *upiv1.PredictValuesResponse
var benchMarkUpiErr *errors.TuringError

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

var mockStream = &mocks.ServerTransportStream{}

// TestMain does setup for all test case pre-run
func TestMain(m *testing.M) {

	// this mock stream is required for grpc.SetHeader to have a stream context to work
	mockStream.On("SetHeader", mock.Anything).Return(nil)

	testutils.RunTestUPIServer(testutils.GrpcTestServer{Port: port})
	testutils.RunTestUPIServer(testutils.GrpcTestServer{Port: grpcport1})
	testutils.RunTestUPIServer(testutils.GrpcTestServer{Port: grpcport2})
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
			cfgFilePath:   "testdata/grpc/grpc_router_minimal.yaml",
			fiberDebugLog: false,
		},
		{
			name:          "ok with fiber debug",
			cfgFilePath:   "testdata/grpc/grpc_router_minimal_two_route.yaml",
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

				res, err := got.Route(ctx, &fibergrpc.Request{
					Message: []byte{},
				})
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
				Code:    14,
				Message: "did not get back a valid response from the fiberHandler",
			},
			mockReturn: fiber.NewResponseQueueFromResponses(),
		},
		{
			name: "error - fiber router error response",
			expectedErr: &errors.TuringError{
				Code:    13,
				Message: "{\n  \"code\": 13,\n  \"error\": \"fiber: request cannot be completed: err\"\n}",
			},
			mockReturn: fiber.NewResponseQueueFromResponses(
				fiber.NewErrorResponse(
					fiberErrors.FiberError{
						Code:    13,
						Message: "fiber: request cannot be completed: err",
					})),
		},
		{
			name: "error non proto payload",
			expectedErr: &errors.TuringError{
				Code:    14,
				Message: "unable to parse fiber response into grpc response",
			},
			mockReturn: fiber.NewResponseQueueFromResponses(fiberhttp.NewHTTPResponse(
				&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("dummy res"))),
				})),
		},
		{
			name: "error wrong response payload type",
			expectedErr: &errors.TuringError{
				Code:    14,
				Message: "unable to unmarshal into expected response proto",
			},
			mockReturn: fiber.NewResponseQueueFromResponses(&fibergrpc.Response{
				Message: []byte("test"),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFiberRouter := &mocks.Component{}
			mockFiberRouter.On("Dispatch", mock.Anything, mock.Anything).Return(tt.mockReturn, nil)
			mc := missionControlUpi{fiberRouter: mockFiberRouter}
			ctx := context.Background()
			ctx = grpc.NewContextWithServerTransportStream(ctx, mockStream)

			got, err := mc.Route(ctx, &fibergrpc.Request{})
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.True(t, compareUpiResponse(got, tt.expected), "response not equal to expected")
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
			name:           "small request",
			request:        &fibergrpc.Request{Message: smallRequestByte},
			compareAgainst: &smallRequestExpected,
			expectedEqual:  true,
		},
		{
			name:           "large request",
			request:        &fibergrpc.Request{Message: largeRequestByte},
			compareAgainst: &largeRequestExpected,
			expectedEqual:  true,
		},
		{
			name:           "large request",
			request:        &fibergrpc.Request{Message: largeRequestByte},
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
			diff := compareUpiResponse(got, tt.compareAgainst)
			require.Equal(t, tt.expectedEqual, diff, "Comparison result not expected")
		})
	}
}

func compareUpiResponse(x *upiv1.PredictValuesResponse, y *upiv1.PredictValuesResponse) bool {
	return cmp.Equal(x, y,
		cmpopts.IgnoreUnexported(
			upiv1.PredictValuesResponse{},
			upiv1.Table{},
			upiv1.Column{},
			upiv1.Row{},
			upiv1.Value{},
			upiv1.Variable{},
			upiv1.ResponseMetadata{},
		))
}

func benchmarkGrpcRoute(rows int, cols int, b *testing.B) {
	mc, err := NewMissionControlUPI(benchmarkConfig, false)
	require.NoError(b, err)
	ctx := context.Background()
	ctx = grpc.NewContextWithServerTransportStream(ctx, mockStream)
	byteReq, err := proto.Marshal(testutils.GenerateUPIRequest(rows, cols))
	require.NoError(b, err)

	req := &fibergrpc.Request{
		Message: byteReq,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchMarkUpiResp, _ = mc.Route(ctx, req)
	}
}

func benchmarkPlainGrpc(rows int, cols int, b *testing.B) {
	upiRequest := testutils.GenerateUPIRequest(rows, cols)

	conn, err := grpc.Dial(fmt.Sprintf(":%d", grpcport1), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(b, err)
	client := upiv1.NewUniversalPredictionServiceClient(conn)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchMarkUpiResp, _ = client.PredictValues(context.Background(), upiRequest)
	}
}

func BenchmarkMissionControlUpiDefaultRouteSmallPayload(b *testing.B) {
	benchmarkGrpcRoute(5, 5, b)
}

func BenchmarkMissionControlUpiDefaultRouteLargePayload(b *testing.B) {
	benchmarkGrpcRoute(100, 100, b)
}

func BenchmarkPlainGrpcUpiSmallPayload(b *testing.B) {
	benchmarkPlainGrpc(5, 5, b)
}

func BenchmarkPlainGrpcUpiLargePayload(b *testing.B) {
	benchmarkPlainGrpc(100, 100, b)
}
