package missionctl

import (
	"bytes"
	"context"
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
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

const port = 50550

// TestMain does setup for all test case pre-run
func TestMain(m *testing.M) {

	testutils.RunTestUPIServer(testutils.GrpcTestServer{Port: port})
	os.Exit(m.Run())
}

// TestNewMissionControlGrpc tests for the creation of missionControlGrpc and fiberLog configuration.
// server reflection is required for the grpc mission control to be created
func TestNewMissionControlGrpc(t *testing.T) {
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

			got, err := NewMissionControlGrpc(tt.cfgFilePath, tt.fiberDebugLog)
			if err != nil {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				res, err := got.Route(context.Background(), &fibergrpc.Request{
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
				Message: "unable to parse fiber response into proto",
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
			upiv1.PredictionResultRow{},
			upiv1.NamedValue{},
			upiv1.ResponseMetadata{},
		))
}
