package upi

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/metrics"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	fiberGrpc "github.com/gojek/fiber/grpc"
	fiberProtocol "github.com/gojek/fiber/protocol"

	"github.com/caraml-dev/turing/engines/router/missionctl"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/experiment"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/constant"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/upi/interceptors"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

const tracingComponentID = "grpc_handler"

type Server struct {
	upiv1.UnimplementedUniversalPredictionServiceServer

	missionControl missionctl.MissionControlUPI
	resultLogger   *resultlog.UPIResultLogger
}

func NewUPIServer(mc missionctl.MissionControlUPI, rl *resultlog.UPIResultLogger) *Server {
	return &Server{
		missionControl: mc,
		resultLogger:   rl,
	}
}

func (us *Server) Run(listener net.Listener) {

	s := grpc.NewServer(grpc.UnaryInterceptor(interceptors.PanicRecoveryInterceptor()))
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
		instrumentation.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(predictionErr == nil)
			},
			"component": func() string {
				return tracingComponentID
			},
			"traffic_rule": func() string { return "" },
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
		sp, ctx = tracing.Glob().StartSpanFromContext(ctx, tracingComponentID)
		if sp != nil {
			defer sp.Finish()
		}
	}

	resp, predictionErr := us.getPrediction(ctx, req, md, turingReqID)
	if predictionErr != nil {
		us.resultLogger.LogTuringRouterRequestError(ctx, predictionErr)
		return nil, status.Error(codes.Code(predictionErr.Code), predictionErr.Message)
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
	respCh := make(chan resultlog.GrpcRouterResponse, 1)

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
			us.resultLogger.LogTuringRouterRequestSummary(md, req, respCh)
		}()
	}()

	// Create a channel for experiment treatment response and add to context
	ch := make(chan *experiment.Response, 1)
	ctx = experiment.WithExperimentResponseChannel(ctx, ch)

	// Calling Routes via fiber
	resp, turingError := us.missionControl.Route(ctx, upiRequest)
	if turingError != nil {
		return nil, turingError
	}
	// type assert to grpc response to get metadata
	grpcResp := resp.(*fiberGrpc.Response)

	predictResponse := &upiv1.PredictValuesResponse{}
	err = proto.Unmarshal(grpcResp.Payload(), predictResponse)
	if err != nil {
		turingError = errors.NewTuringError(
			errors.Newf(errors.BadResponse, "unable to unmarshal into expected response proto"), fiberProtocol.GRPC,
		)
		return nil, turingError
	}

	// Get the experiment treatment channel from the request context, read result
	var experimentResponse *experiment.Response
	expResultCh, err := experiment.GetExperimentResponseChannel(ctx)
	if err == nil {
		select {
		case experimentResponse = <-expResultCh:
		default:
			break
		}
	}

	// Creates ResponseMetadata if its nil
	predictResponse = populateResponseMetadata(predictResponse, turingReqID, experimentResponse)

	us.resultLogger.SendResponseToLogChannel(
		respCh,
		resultlog.ResultLogKeys.Router,
		grpcResp.Metadata,
		predictResponse,
		turingError)
	return predictResponse, nil
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

func populateResponseMetadata(
	resp *upiv1.PredictValuesResponse,
	id string,
	experimentResponse *experiment.Response) *upiv1.PredictValuesResponse {
	if resp.Metadata == nil {
		resp.Metadata = &upiv1.ResponseMetadata{}
	}

	if experimentResponse != nil {
		// For now, log experiment engine error to console only.
		if experimentResponse.Error != "" {
			log.Glob().Errorf("error response from experiment engine %s", experimentResponse.Error)
		} else {
			resp.Metadata.TreatmentName = experimentResponse.TreatmentName
			resp.Metadata.ExperimentName = experimentResponse.ExperimentName
		}
	}

	resp.Metadata.PredictionId = id
	return resp
}
