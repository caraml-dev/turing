package missionctl

import (
	"context"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/fiberapi"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
	"google.golang.org/protobuf/proto"
)

type MissionControlUPI interface {
	Route(context.Context, fiber.Request) (*upiv1.PredictValuesResponse, *errors.TuringError)
}

type missionControlUpi struct {
	fiberRouter fiber.Component
}

// NewMissionControlUpi creates new instance of the MissingControl,
// based on the grpc configuration of fiber.yaml
func NewMissionControlUpi(
	cfgFilePath string,
	fiberDebugLog bool,
) (MissionControlUPI, error) {
	fiberRouter, err := fiberapi.CreateFiberRouterFromConfig(cfgFilePath, fiberDebugLog)
	if err != nil {
		return nil, err
	}

	return &missionControlUpi{
		fiberRouter: fiberRouter,
	}, nil
}

func (us *missionControlUpi) Route(
	ctx context.Context,
	fiberRequest fiber.Request) (
	*upiv1.PredictValuesResponse, *errors.TuringError) {
	var turingError *errors.TuringError
	defer metrics.GetMeasureDurationFunc(turingError, "route")()

	resp, ok := <-us.fiberRouter.Dispatch(ctx, fiberRequest).Iter()
	if !ok {
		turingError = errors.NewTuringError(
			errors.Newf(errors.BadResponse, "did not get back a valid response from the fiberHandler"), errors.GRPC,
		)
		return nil, turingError
	}
	if !resp.IsSuccess() {
		return nil, &errors.TuringError{
			Code:    resp.StatusCode(),
			Message: string(resp.Payload().([]byte)),
		}
	}

	var responseProto upiv1.PredictValuesResponse
	payload, ok := resp.Payload().(proto.Message)
	if !ok {
		turingError = errors.NewTuringError(
			errors.Newf(errors.BadResponse, "unable to parse fiber response into proto"), errors.GRPC,
		)
		return nil, turingError
	}
	payloadByte, err := proto.Marshal(payload)
	if err != nil {
		turingError = errors.NewTuringError(
			errors.Newf(errors.BadResponse, "unable to marshal payload"), errors.GRPC,
		)
		return nil, turingError
	}
	err = proto.Unmarshal(payloadByte, &responseProto)
	if err != nil {
		turingError = errors.NewTuringError(
			errors.Newf(errors.BadResponse, "unable to unmarshal into expected response proto"), errors.GRPC,
		)
		return nil, turingError
	}
	return &responseProto, nil
}
