package missionctl

import (
	"context"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/fiberapi"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
	fibergrpc "github.com/gojek/fiber/grpc"
	"google.golang.org/protobuf/proto"
)

type MissionControlGrpc interface {
	IsEnricherEnabled() bool
	IsEnsemblerEnabled() bool
	PredictValues(
		ctx context.Context,
		req *upiv1.PredictValuesRequest,
	) (*upiv1.PredictValuesResponse, error)
}

type missionControlGrpc struct {
	fiberRouter fiber.Component
}

func (mc *missionControlGrpc) IsEnricherEnabled() bool { return false }

func (mc *missionControlGrpc) IsEnsemblerEnabled() bool { return false }

// NewMissionControlGrpc creates new instance of the MissingControl,
// based on the http.Client and configuration passed into it
func NewMissionControlGrpc(
	routerCfg *config.RouterConfig,
	appCfg *config.AppConfig,
) (MissionControlGrpc, error) {
	fiberRouter, err := fiberapi.CreateFiberRouterFromConfig(routerCfg.ConfigFile, appCfg.FiberDebugLog)
	if err != nil {
		return nil, err
	}

	return &missionControlGrpc{
		fiberRouter: fiberRouter,
	}, nil
}

func (mc *missionControlGrpc) PredictValues(ctx context.Context, req *upiv1.PredictValuesRequest) (
	*upiv1.PredictValuesResponse, error) {
	fiberRequest := &fibergrpc.Request{
		Message: req,
	}
	resp, ok := <-mc.fiberRouter.Dispatch(ctx, fiberRequest).Iter()
	// TODO need to refactor and use generic error + correct response
	if !ok {
		return nil, errors.NewHTTPError(errors.Newf(errors.BadResponse,
			"did not get back a valid response from the fiberHandler"))
	}
	if !resp.IsSuccess() {
		return nil, errors.NewHTTPError(errors.NewHTTPError(errors.Newf(errors.BadResponse,
			string(resp.Payload().([]byte)))))
	}

	var responseProto upiv1.PredictValuesResponse
	payload, ok := resp.Payload().(proto.Message)
	if !ok {
		return nil, errors.NewHTTPError(errors.Newf(errors.BadResponse, "unable to parse fiber response into proto"))
	}
	payloadByte, err := proto.Marshal(payload)
	if err != nil {
		return nil, errors.NewHTTPError(errors.Newf(errors.BadResponse, "unable to marshal paryload"))
	}
	err = proto.Unmarshal(payloadByte, &responseProto)
	if err != nil {
		return nil, errors.NewHTTPError(errors.Newf(errors.BadResponse, "unable to unmarshal into expected response proto"))
	}
	return &responseProto, nil
}
