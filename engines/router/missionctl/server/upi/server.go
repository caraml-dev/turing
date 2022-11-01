package upi

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	fiberGrpc "github.com/gojek/fiber/grpc"
	fiberProtocol "github.com/gojek/fiber/protocol"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/caraml-dev/turing/engines/router/missionctl"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/constant"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

const tracingComponentID = "grpc_handler"

type Server struct {
	upiv1.UnimplementedUniversalPredictionServiceServer

	missionControl missionctl.MissionControlUPI
}

func NewUPIServer(mc missionctl.MissionControlUPI) *Server {
	return &Server{
		missionControl: mc,
	}
}

func (us *Server) Run(listener net.Listener) {
	s := grpc.NewServer()
	//TODO: the unmarshalling can be done more efficiently by using partial deserialization
	upiv1.RegisterUniversalPredictionServiceServer(s, us)
	reflection.Register(s)

	errChan := make(chan error, 1)
	stopChan := make(chan os.Signal, 1)

	// bind OS events to the signal channel
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := s.Serve(listener); err != nil {
			errChan <- err
		}
	}()

	defer func() {
		s.GracefulStop()
	}()

	// block until either OS signal, or server fatal error
	select {
	case err := <-errChan:
		log.Glob().Errorf("Failed to start Turing Mission Control API: %s", err)
	case <-stopChan:
		log.Glob().Info("Signal to stop server")
	}

}

func (us *Server) PredictValues(ctx context.Context, req *upiv1.PredictValuesRequest) (
	*upiv1.PredictValuesResponse, error) {
	var predictionErr *errors.TuringError // Measure execution time
	defer metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(predictionErr == nil)
			},
			"component": func() string {
				return tracingComponentID
			},
		},
	)()

	// Create context from the request context
	ctx = turingctx.NewTuringContext(ctx)
	// Create context logger
	ctxLogger := log.WithContext(ctx)

	// if request comes with metadata, attach it to metadata to be sent with fiber
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}

	// Get the unique turing request id from the context
	turingReqID, err := turingctx.GetRequestID(ctx)
	if err != nil {
		ctxLogger.Errorf("Could not retrieve Turing Request ID from context: %v",
			err.Error())
	}
	ctxLogger.Debugf("Received request for %v", turingReqID)
	md.Append(constant.TuringReqIDHeaderKey, turingReqID)

	if tracing.Glob().IsEnabled() {
		var sp opentracing.Span
		sp, _ = tracing.Glob().StartSpanFromContext(ctx, tracingComponentID)
		if sp != nil {
			defer sp.Finish()
		}
	}

	resp, predictionErr := us.getPrediction(ctx, req, md, turingReqID)
	if predictionErr != nil {
		logTuringRouterRequestError(ctx, predictionErr)
		return nil, predictionErr
	}
	return resp, nil
}

func (us *Server) getPrediction(
	ctx context.Context,
	req *upiv1.PredictValuesRequest,
	md metadata.MD,
	turingReqID string) (
	*upiv1.PredictValuesResponse, *errors.TuringError) {

	// Create response channel to store the response from each step. 1 for route now,
	// should be 4 when experiment engine, enricher and ensembler are added
	respCh := make(chan grpcRouterResponse, 1)

	req = populateRequestMetadata(req, turingReqID)
	requestByte, err := proto.Marshal(req)
	if err != nil {
		turingError := errors.NewTuringError(
			errors.Newf(errors.BadInput, "unable to parse request into byte"), fiberProtocol.GRPC,
		)
		return nil, turingError
	}

	upiRequest := fiberGrpc.NewRequest(md, requestByte, req)

	// Defer logging req summary
	defer func() {
		go func() {
			close(respCh)
			logTuringRouterRequestSummary(ctx, time.Now(), md, req, respCh)
		}()
	}()

	// Calling Routes via fiber
	resp, turingError := us.missionControl.Route(ctx, upiRequest)
	if turingError != nil {
		return nil, turingError
	}

	responseProto := &upiv1.PredictValuesResponse{}
	err = proto.Unmarshal(resp.Payload(), responseProto)
	if err != nil {
		turingError = errors.NewTuringError(
			errors.Newf(errors.BadResponse, "unable to unmarshal into expected response proto"), fiberProtocol.GRPC,
		)
		return nil, turingError
	}

	responseProto = populateResponseMetadata(responseProto, turingReqID)
	copyResponseToLogChannel(ctx, respCh, resultlog.ResultLogKeys.Router, responseProto, turingError)
	return responseProto, nil
}

func populateRequestMetadata(req *upiv1.PredictValuesRequest, id string) *upiv1.PredictValuesRequest {
	if req.Metadata == nil {
		req.Metadata = &upiv1.RequestMetadata{}
	}

	if req.Metadata.RequestTimestamp == nil {
		req.Metadata.RequestTimestamp = timestamppb.Now()
	}

	req.Metadata.PredictionId = id
	return req
}

func populateResponseMetadata(resp *upiv1.PredictValuesResponse, id string) *upiv1.PredictValuesResponse {
	if resp.Metadata == nil {
		resp.Metadata = &upiv1.ResponseMetadata{}
	}

	resp.Metadata.PredictionId = id
	return resp
}
