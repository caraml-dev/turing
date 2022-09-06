package missionctl

import (
	"context"
	"net/http"
	"strings"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/fiberapi"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
	fibergrpc "github.com/gojek/fiber/grpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/metadata"
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

const turingReqIDHeaderKey = "Turing-Req-ID"
const tracingComponentId = "grpc_handler"

func (mc *missionControlGrpc) IsEnricherEnabled() bool { return false }

func (mc *missionControlGrpc) IsEnsemblerEnabled() bool { return false }

// NewMissionControlGrpc creates new instance of the MissingControl,
// based on the grpc configuration of fiber.yaml
func NewMissionControlGrpc(
	cfgFilePath string,
	fiberDebugLog bool,
) (MissionControlGrpc, error) {
	fiberRouter, err := fiberapi.CreateFiberRouterFromConfig(cfgFilePath, fiberDebugLog)
	if err != nil {
		return nil, err
	}

	return &missionControlGrpc{
		fiberRouter: fiberRouter,
	}, nil
}

func (mc *missionControlGrpc) PredictValues(parentCtx context.Context, req *upiv1.PredictValuesRequest) (
	*upiv1.PredictValuesResponse, error) {

	var err error
	measureDurationFunc := getMeasureDurationFunc(err)
	defer measureDurationFunc()

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
	md.Set(turingReqIDHeaderKey, turingReqID)

	if tracing.Glob().IsEnabled() {
		var sp opentracing.Span
		ctx, sp = enableTracingSpan(ctx, md, tracingComponentId)
		if sp != nil {
			defer sp.Finish()
		}
	}

	fiberRequest := &fibergrpc.Request{
		Message:  req,
		Metadata: md,
	}
	resp, ok := <-mc.fiberRouter.Dispatch(ctx, fiberRequest).Iter()
	if !ok {
		err = errors.NewHTTPError(
			errors.Newf(errors.BadResponse, "did not get back a valid response from the fiberHandler"),
		)
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, errors.Newf(errors.BadResponse, string(resp.Payload().([]byte)))
	}

	var responseProto upiv1.PredictValuesResponse
	payload, ok := resp.Payload().(proto.Message)
	if !ok {
		return nil, errors.Newf(errors.BadResponse, "unable to parse fiber response into proto")
	}
	payloadByte, err := proto.Marshal(payload)
	if err != nil {
		return nil, errors.Newf(errors.BadResponse, "unable to marshal payload")
	}
	err = proto.Unmarshal(payloadByte, &responseProto)
	if err != nil {
		return nil, errors.Newf(errors.BadResponse, "unable to unmarshal into expected response proto")
	}
	return &responseProto, nil
}

func getMeasureDurationFunc(err error) func() {
	// Measure the duration of handler function
	return metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(err == nil)
			},
			"component": func() string {
				return tracingComponentId
			},
		},
	)
}

// enableTracingSpan associates span to context, if applicable.
// Converts grpc headers to http headers, since both are map of string underlying
func enableTracingSpan(ctx context.Context, req metadata.MD, operationName string) (context.Context, opentracing.Span) {
	var sp opentracing.Span
	var httpHeader http.Header
	for k, v := range req {
		httpHeader.Set(k, strings.Join(v, ","))
	}
	sp, ctx = tracing.Glob().StartSpanFromRequestHeader(ctx, operationName, httpHeader)
	return ctx, sp
}

// logTuringRouterRequestError logs the given turing request id and the error data
func logTuringRouterRequestError(ctx context.Context, err *errors.HTTPError) {
	logger := log.WithContext(ctx)
	defer func() {
		_ = logger.Sync()
	}()
	logger.Errorw("Turing Request Error",
		"error", err.Message,
		"status", err.Code,
	)
}
