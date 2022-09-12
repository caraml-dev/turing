package missionctl

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	port            = 50550
	grpcport1       = 50556
	grpcport2       = 50557
	benchmarkConfig = "testdata/grpc/grpc_router_minimal_two_route.yaml"
)

var benchMarkUpiResp *upiv1.PredictValuesResponse
var benchMarkUpiErr *errors.TuringError

// TestMain does setup for all test case pre-run
func TestMain(m *testing.M) {

	testutils.RunTestUPIServer(testutils.GrpcTestServer{Port: port})
	testutils.RunTestUPIServer(testutils.GrpcTestServer{Port: grpcport1})
	testutils.RunTestUPIServer(testutils.GrpcTestServer{Port: grpcport2})
	os.Exit(m.Run())
}

// TestNewMissionControlUpi tests for the creation of missionControlGrpc and fiberLog configuration.
// server reflection is required for the grpc mission control to be created
func TestNewMissionControlUpi(t *testing.T) {
	fiberDebugMsg := "Time Taken"

	// this mock stream is required for grpc.SetHeader to have a stream context to work
	mockStream := &mocks.ServerTransportStream{}
	mockStream.On("SetHeader", mock.Anything).Return(nil)

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
			cfgFilePath:   "testdata/grpc/grpc_router_minimal.yaml",
			fiberDebugLog: true,
		},
		{
			name:        "faulty port",
			cfgFilePath: "testdata/grpc/grpc_router_faulty_port.yaml",
			expectedErr: "fiber: request cannot be completed: grpc dispatcher: " +
				"unable to get reflection information, ensure server reflection is enable and config are correct",
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
					Message: &upiv1.PredictValuesRequest{},
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

func Test_missionControlUpi_Route(t *testing.T) {
	tests := []struct {
		name        string
		expected    *upiv1.PredictValuesResponse
		mockReturn  fiber.ResponseQueue
		expectedErr *errors.TuringError
	}{
		{
			name:     "ok",
			expected: testutils.GetDefaultMockResponse(),
			mockReturn: fiber.NewResponseQueueFromResponses(&fibergrpc.Response{
				Message: testutils.GetDefaultMockResponse(),
			}),
		},
		{
			name: "error wrong respaonse payload type",
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
				Message: &upiv1.ResponseMetadata{
					PredictionId: "123",
				},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFiberRouter := &mocks.Component{}
			mockFiberRouter.On("Dispatch", mock.Anything, mock.Anything).Return(tt.mockReturn, nil)
			mc := missionControlUpi{fiberRouter: mockFiberRouter}

			got, err := mc.Route(context.Background(), &fibergrpc.Request{})
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.True(t, compareUpiResponse(got, tt.expected), "response not equal to expected")
			}

		})
	}
}

func compareUpiResponse(x *upiv1.PredictValuesResponse, y *upiv1.PredictValuesResponse) bool {
	return cmp.Equal(x, y,
		cmpopts.IgnoreUnexported(
			upiv1.PredictValuesResponse{},
			upiv1.Table{},
			upiv1.NamedValue{},
			upiv1.ResponseMetadata{},
		))
}

func benchmarkGrpcRoute(payloadFileName string, b *testing.B) {

	mc, err := NewMissionControlUPI(benchmarkConfig, false)
	require.NoError(b, err)

	upiRequest := &upiv1.PredictValuesRequest{}
	fileByte, err := ioutil.ReadFile(filepath.Join("testdata", payloadFileName))
	require.NoError(b, err)
	err = protojson.Unmarshal(fileByte, upiRequest)
	require.NoError(b, err)

	req := &fibergrpc.Request{
		Message: upiRequest,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchMarkUpiResp, benchMarkUpiErr = mc.Route(context.Background(), req)
	}
}

func benchmarkPlainGrpc(payloadFileName string, b *testing.B) {

	upiRequest := &upiv1.PredictValuesRequest{}
	fileByte, err := ioutil.ReadFile(filepath.Join("testdata", payloadFileName))
	require.NoError(b, err)
	err = protojson.Unmarshal(fileByte, upiRequest)
	require.NoError(b, err)

	conn, err := grpc.Dial(fmt.Sprintf(":%d", grpcport1), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(b, err)
	client := upiv1.NewUniversalPredictionServiceClient(conn)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchMarkUpiResp, _ = client.PredictValues(context.Background(), upiRequest)
	}
}

func BenchmarkMissionControlUpiDefaultRouteSmallPayload(b *testing.B) {
	benchmarkGrpcRoute("upi_small_payload.json", b)
}

func BenchmarkMissionControlUpiDefaultRouteLargePayload(b *testing.B) {
	benchmarkGrpcRoute("upi_large_payload.json", b)
}

func BenchmarkPlainGrpcUpiSmallPayload(b *testing.B) {
	benchmarkPlainGrpc("upi_small_payload.json", b)
}

func BenchmarkPlainGrpcUpiLargePayload(b *testing.B) {
	benchmarkPlainGrpc("upi_large_payload.json", b)
}
