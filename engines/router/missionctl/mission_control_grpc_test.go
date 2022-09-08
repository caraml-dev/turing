package missionctl

import (
	"context"
	"testing"

	"github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/require"
)

const port = 50550

func TestNewMissionControlGrpc(t *testing.T) {
	testutils.RunTestUPIServer(testutils.GrpcTestServer{Port: port})
	fiberDebugMsg := `"Time Taken","Component"`

	tests := []struct {
		name          string
		cfgFilePath   string
		fiberDebugLog bool
		expected      MissionControlGrpc
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
			logger, sink, _ := testutils.NewLoggerWithMemorySink()
			log.SetGlobalLogger(logger)

			got, err := NewMissionControlGrpc(tt.cfgFilePath, tt.fiberDebugLog)
			if err != nil {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				res, err := got.PredictValues(context.Background(), &upiv1.PredictValuesRequest{})
				require.NoError(t, err)
				require.NotNil(t, res)

				logData := sink.String()
				if tt.fiberDebugLog {
					require.Contains(t, logData, fiberDebugMsg)
				} else {
					require.NotContains(t, logData, fiberDebugMsg)
				}
			}
		})
	}
}
