package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/caraml-dev/turing/engines/router/missionctl"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
	fibergrpc "github.com/gojek/fiber/grpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

const turingReqIDHeaderKey = "Turing-Req-ID"
const tracingComponentID = "grpc_handler"

type UPIServer struct {
	upiv1.UnimplementedUniversalPredictionServiceServer

	missionControl missionctl.MissionControlUPI
	port           int
}

func NewUPIServer(mc missionctl.MissionControlUPI, port int) *UPIServer {
	return &UPIServer{
		missionControl: mc,
		port:           port,
	}
}

func (us *UPIServer) Run() {
	s := grpc.NewServer()
	upiv1.RegisterUniversalPredictionServiceServer(s, us)
	reflection.Register(s)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", us.port))
	if err != nil {
		log.Glob().Errorf("Failed to listen on port %d: %s", us.port, err)
	}
	if err := s.Serve(listener); err != nil {
		log.Glob().Errorf("Failed to start Turing Mission Control API: %s", err)
	}
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
	resp, err := us.missionControl.Route(ctx, fiberRequest)
	if err != nil {
		return nil, err
	}
	copyResponseToLogChannel(ctx, respCh, resultlog.ResultLogKeys.Router, resp, err)

	return resp, err
}
