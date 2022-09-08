package grpc

import (
	"context"
	"time"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/fiberapi"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
	fibergrpc "github.com/gojek/fiber/grpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

const turingReqIDHeaderKey = "Turing-Req-ID"
const tracingComponentID = "grpc_handler"

type UPIServer struct {
	upiv1.UnimplementedUniversalPredictionServiceServer

	fiberRouter fiber.Component
}

func NewUPIServer(
	cfgFilePath string,
	fiberDebugLog bool,
) (*UPIServer, error) {

	fiberRouter, err := fiberapi.CreateFiberRouterFromConfig(cfgFilePath, fiberDebugLog)
	if err != nil {
		return nil, err
	}

	return &UPIServer{
		fiberRouter: fiberRouter,
	}, nil
}

func (us *UPIServer) PredictValues(parentCtx context.Context, req *upiv1.PredictValuesRequest) (
	*upiv1.PredictValuesResponse, error) {
	var predictionErr *errors.TuringError
	defer metrics.GetMeasureDurationFunc(predictionErr, tracingComponentID)()

	// Create context from the request context
	ctx := turingctx.NewTuringContext(parentCtx)
	// Create context logger
	ctxLogger := log.WithContext(ctx)
	defer func() {
		_ = ctxLogger.Sync()
	}()

	// Get the unique turing request id from the context
	turingReqID, err := turingctx.GetRequestID(ctx)
	if err != nil {
		ctxLogger.Errorf("Could not retrieve Turing Request ID from context: %v",
			err.Error())
	}
	ctxLogger.Debugf("Received request for %v", turingReqID)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	md.Set(turingReqIDHeaderKey, turingReqID)

	if tracing.Glob().IsEnabled() {
		var sp opentracing.Span
		sp, _ = tracing.Glob().StartSpanFromContext(ctx, tracingComponentID)
		if sp != nil {
			defer sp.Finish()
		}
	}

	fiberRequest := &fibergrpc.Request{
		Message:  req,
		Metadata: md,
	}

	resp, predictionErr := us.getPrediction(ctx, fiberRequest)
	if predictionErr != nil {
		logTuringRouterRequestError(ctx, predictionErr)
		return nil, predictionErr
	}
	return resp, nil
}

func (us *UPIServer) getPrediction(
	ctx context.Context,
	fiberRequest fiber.Request) (
	*upiv1.PredictValuesResponse, *errors.TuringError) {

	// Create response channel to store the response from each step. 1 for route now,
	// should be 4 when experiment engine, enricher and ensembler are added
	respCh := make(chan grpcRouterResponse, 1)
	defer close(respCh)

	// Defer logging request summary
	defer func() {
		go logTuringRouterRequestSummary(
			ctx,
			time.Now(),
			fiberRequest.Header(),
			fiberRequest.Payload().(*upiv1.PredictValuesRequest),
			respCh)
	}()

	// Calling Routes via fiber
	resp, err := us.Route(ctx, fiberRequest)
	if err != nil {
		return nil, err
	}
	copyResponseToLogChannel(ctx, respCh, resultlog.ResultLogKeys.Router, resp, err)

	return resp, err
}

func (us *UPIServer) Route(
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
