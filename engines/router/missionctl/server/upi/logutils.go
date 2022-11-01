package upi

import (
	"context"
	"time"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
)

type grpcRouterResponse struct {
	key    string
	header metadata.MD
	body   *upiv1.PredictValuesResponse
	err    string
}

var protoJSONMarshaller = protojson.MarshalOptions{}

func logTuringRouterRequestSummary(
	ctx context.Context,
	timestamp time.Time,
	header metadata.MD,
	upiReq *upiv1.PredictValuesRequest,
	mcRespCh <-chan grpcRouterResponse,
) {

	// Create a new TuringResultLogEntry record with the context and request info
	// send proto as json string
	reqStr := protoJSONMarshaller.Format(upiReq)
	logEntry := resultlog.NewTuringResultLogEntry(ctx, timestamp, header, reqStr)

	// Read incoming responses and prepare for logging
	for resp := range mcRespCh {
		// If error exists, add an error record
		if resp.err != "" {
			logEntry.AddResponse(resp.key, "", nil, resp.err)
		} else {
			upiResp := protoJSONMarshaller.Format(upiReq)
			logEntry.AddResponse(resp.key, upiResp, resultlog.FormatHeader(resp.header), "")
		}
	}

	// Log the responses. If an error occurs in logging the result to the
	// configured result log destination, log the error.
	if err := resultlog.LogEntry(logEntry); err != nil {
		log.Glob().Errorf("Result Logging Error: %s", err.Error())
	}
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

	// if no metadata was sent return, md will be nil
	md, _ := metadata.FromIncomingContext(ctx)

	// Copy to channel
	ch <- grpcRouterResponse{
		key:    key,
		header: md,
		body:   r,
	}
}
