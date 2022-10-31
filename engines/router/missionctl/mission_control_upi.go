package missionctl

import (
	"context"

	"github.com/gojek/fiber"
	fiberGrpc "github.com/gojek/fiber/grpc"
	fiberProtocol "github.com/gojek/fiber/protocol"
	"google.golang.org/grpc"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/fiberapi"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
)

type MissionControlUPI interface {
	Route(context.Context, fiber.Request) (fiber.Response, *errors.TuringError)
}

type missionControlUpi struct {
	fiberRouter fiber.Component
}

// NewMissionControlUPI creates new instance of the MissingControl,
// based on the grpc configuration of fiber.yaml
func NewMissionControlUPI(
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
	fiberRequest fiber.Request,
) (fiber.Response, *errors.TuringError) {
	var turingError *errors.TuringError
	// Measure execution time
	defer metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(turingError == nil)
			},
			"component": func() string {
				return "route"
			},
		},
	)()

	resp, ok := <-us.fiberRouter.Dispatch(ctx, fiberRequest).Iter()
	if !ok {
		turingError = errors.NewTuringError(
			errors.Newf(errors.BadResponse, "did not get back a valid response from the fiberHandler"), fiberProtocol.GRPC,
		)
		return nil, turingError
	}
	if !resp.IsSuccess() {
		return nil, &errors.TuringError{
			Code:    resp.StatusCode(),
			Message: string(resp.Payload()),
		}
	}

	grpcResponse, ok := resp.(*fiberGrpc.Response)
	if !ok {
		turingError = errors.NewTuringError(
			errors.Newf(errors.BadResponse, "unable to parse fiber response into grpc response"), fiberProtocol.GRPC,
		)
		return nil, turingError
	}

	// attach metadata to context if exist
	if len(grpcResponse.Metadata) > 0 {
		err := grpc.SetHeader(ctx, grpcResponse.Metadata)
		if err != nil {
			turingError = errors.NewTuringError(
				errors.Newf(errors.BadResponse, "unable to send headers: %s", err.Error()), fiberProtocol.GRPC,
			)
			return nil, turingError
		}
	}

	return grpcResponse, nil
}
