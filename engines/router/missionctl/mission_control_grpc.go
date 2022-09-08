package missionctl

import (
	"context"
	"net/http"
	"strings"
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

type MissionControlGrpc interface {
	PredictValues(
		ctx context.Context,
		req *upiv1.PredictValuesRequest,
	) (*upiv1.PredictValuesResponse, error)
}

type missionControlGrpc struct {
	fiberRouter fiber.Component
}

const turingReqIDHeaderKey = "Turing-Req-ID"
const tracingComponentID = "grpc_handler"

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
	var predictionErr *errors.TuringError
	measureDurationFunc := getMeasureDurationFunc(predictionErr, tracingComponentID)
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
	if !ok {
		ctxLogger.Errorf("Could not fetch headers from context: %v",
			err.Error())
	}
	md.Set(turingReqIDHeaderKey, turingReqID)

	if tracing.Glob().IsEnabled() {
		var sp opentracing.Span
		ctx, sp = enableTracingSpan(ctx, md, tracingComponentID)
		if sp != nil {
			defer sp.Finish()
		}
	}

	fiberRequest := &fibergrpc.Request{
		Message:  req,
		Metadata: md,
	}

	resp, predictionErr := mc.getPrediction(ctx, fiberRequest)
	if predictionErr != nil {
		logTuringRouterRequestError(ctx, predictionErr)
		return nil, predictionErr
	}
	return resp, nil
}

func (mc *missionControlGrpc) getPrediction(
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
	resp, err := mc.Route(ctx, fiberRequest)
	if err != nil {
		return nil, err
	}
	copyResponseToLogChannel(ctx, respCh, resultlog.ResultLogKeys.Router, resp, err)

	return resp, err
}

func (mc *missionControlGrpc) Route(
	ctx context.Context,
	fiberRequest fiber.Request) (
	*upiv1.PredictValuesResponse, *errors.TuringError) {
	var turingError *errors.TuringError
	measureDurationFunc := getMeasureDurationFunc(turingError, "route")
	defer measureDurationFunc()

	resp, ok := <-mc.fiberRouter.Dispatch(ctx, fiberRequest).Iter()
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
func getMeasureDurationFunc(err error, componentName string) func() {
	// Measure the duration of handler function
	return metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(err == nil)
			},
			"component": func() string {
				return componentName
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
func logTuringRouterRequestError(ctx context.Context, err *errors.TuringError) {
	logger := log.WithContext(ctx)
	defer func() {
		_ = logger.Sync()
	}()
	logger.Errorw("Turing Request Error",
		"error", err.Message,
		"status", err.Code,
	)
}

type grpcRouterResponse struct {
	key    string
	header metadata.MD
	body   *upiv1.PredictValuesResponse
	err    string
}

func logTuringRouterRequestSummary(
	ctx context.Context,
	timestamp time.Time,
	header metadata.MD,
	body *upiv1.PredictValuesRequest,
	mcRespCh <-chan grpcRouterResponse,
) {

	// Create a new TuringResultLogEntry record with the context and request info
	logEntry := resultlog.NewTuringResultLogEntry(ctx, timestamp, header, body.String())

	// Read incoming responses and prepare for logging
	for resp := range mcRespCh {
		// If error exists, add an error record
		if resp.err != "" {
			logEntry.AddResponse(resp.key, "", nil, resp.err)
		} else {
			logEntry.AddResponse(resp.key, resp.body.String(), resultlog.FormatHeader(resp.header), "")
		}
	}

	// Log the responses. If an error occurs in logging the result to the
	// configured result log destination, log the error.
	if err := resultlog.LogEntry(logEntry); err != nil {
		log.Glob().Errorf("Result Logging Error: %s", err.Error())
	}
}

// copyResponseToLogChannel copies the response from the turing router to the given channel
// as a routerResponse object
func copyResponseToLogChannel(
	ctx context.Context,
	ch chan<- grpcRouterResponse,
	key string,
	r *upiv1.PredictValuesResponse,
	err *errors.TuringError) {
	// if error is not nil, use error as response
	if err != nil {
		ch <- grpcRouterResponse{
			key: key,
			err: err.Message,
		}
		return
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		ch <- grpcRouterResponse{
			key: key,
			err: "fail to read metadata from fiber",
		}
		return
	}

	// Copy to channel
	ch <- grpcRouterResponse{
		key:    key,
		header: md,
		body:   r,
	}
}
