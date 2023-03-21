package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation"
	fiberProtocol "github.com/gojek/fiber/protocol"
	"github.com/opentracing/opentracing-go"

	"github.com/caraml-dev/turing/engines/router/missionctl"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/constant"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"

	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
)

const batchHTTPHandlerID = "batch_http_handler"

type batchHTTPHandler struct {
	httpHandler
}

type batchResponse struct {
	StatusCode int             `json:"code"`
	ErrorMsg   string          `json:"error,omitempty"`
	Data       json.RawMessage `json:"data,omitempty"`
}

// NewBatchHTTPHandler creates an instance of the Mission Control's prediction request handler
func NewBatchHTTPHandler(mc missionctl.MissionControl, rl *resultlog.ResultLogger) http.Handler {
	return &batchHTTPHandler{httpHandler{MissionControl: mc, rl: rl}}
}

func (h *batchHTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var httpErr *errors.TuringError
	defer metrics.Glob().MeasureDurationMs(
		instrumentation.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(httpErr == nil)
			},
			"component": func() string {
				return batchHTTPHandlerID
			},
			"traffic_rule": func() string { return "" },
		},
	)()

	// Create context from the request context
	ctx := turingctx.NewTuringContext(req.Context())
	// Create context logger
	ctxLogger := log.WithContext(ctx)

	// Get the unique turing request id from the context
	turingReqID, err := turingctx.GetRequestID(ctx)
	if err != nil {
		ctxLogger.Errorf("Could not retrieve Turing Request ID from context: %v",
			err.Error())
	}
	ctxLogger.Debugf("Received batch request for %v", turingReqID)

	if tracing.Glob().IsEnabled() {
		var sp opentracing.Span
		ctx, sp = h.enableTracingSpan(ctx, req, batchHTTPHandlerID)
		if sp != nil {
			defer sp.Finish()
		}
	}

	// Read the request body
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		h.error(ctx, rw, errors.NewTuringError(err, fiberProtocol.HTTP))
		return
	}

	//Split into batches
	var batchRequests []json.RawMessage
	err = json.Unmarshal(requestBody, &batchRequests)
	if err != nil {
		h.error(ctx, rw, errors.NewTuringError(errors.Newf(errors.BadInput,
			`Invalid json request`), fiberProtocol.HTTP))
		return
	}

	//Handle request asynchronously
	var batchResponses = make([]batchResponse, len(batchRequests))
	var waitGroup sync.WaitGroup
	for index, value := range batchRequests {
		waitGroup.Add(1)

		go func(index int, jsonRequestBody json.RawMessage) {
			defer waitGroup.Done()
			requestContext := turingctx.NewTuringContextWithSuffix(ctx, strconv.Itoa(index))
			var batchResponse batchResponse
			resp, httpErr := h.getPrediction(requestContext, req, ctxLogger, jsonRequestBody)
			if httpErr != nil {
				batchResponse.StatusCode = httpErr.Code
				batchResponse.ErrorMsg = httpErr.Message
				batchResponses[index] = batchResponse
				return
			}
			batchResponse.StatusCode = http.StatusOK
			batchResponse.Data = resp.Body()
			batchResponses[index] = batchResponse
		}(index, value)
	}
	waitGroup.Wait()

	// Write the json response to the writer
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set(constant.TuringReqIDHeaderKey, turingReqID)
	rw.WriteHeader(http.StatusOK)
	batchResponseByte, _ := json.Marshal(batchResponses)
	contentLength, err := rw.Write(batchResponseByte)
	if err != nil {
		ctxLogger.Errorf("Error occurred when copying content: %v", err.Error())
	}
	ctxLogger.Debugf("Written %d bytes", contentLength)
}
