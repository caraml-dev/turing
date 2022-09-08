package missionctl

import (
	"context"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/grpc"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
)

type MissionControlGrpc interface {
	PredictValues(
		ctx context.Context,
		req *upiv1.PredictValuesRequest,
	) (*upiv1.PredictValuesResponse, error)
}

// NewMissionControlGrpc creates new instance of the MissingControl,
// based on the grpc configuration of fiber.yaml
func NewMissionControlGrpc(
	cfgFilePath string,
	fiberDebugLog bool,
) (MissionControlGrpc, error) {
	upiServer, err := grpc.NewUPIServer(cfgFilePath, fiberDebugLog)
	if err != nil {
		return nil, err
	}

	return upiServer, nil
}
