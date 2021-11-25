package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/gojek/turing/engines/router/missionctl"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/gojek/turing/engines/router/missionctl/log"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
	"github.com/opentracing/opentracing-go"
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
func NewBatchHTTPHandler(mc missionctl.MissionControl) http.Handler {
	return &batchHTTPHandler{httpHandler{mc}}
}

func (h *batchHTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var httpErr *errors.HTTPError
	measureDurationFunc := h.getMeasureDurationFunc(httpErr)
	defer measureDurationFunc()

	// Create context from the request context
	ctx := turingctx.NewTuringContext(req.Context())
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
	ctxLogger.Debugf("Received batch request for %v", turingReqID)

	if tracing.Glob().IsEnabled() {
		var sp opentracing.Span
		ctx, sp = h.enableTracingSpan(ctx, req, batchHTTPHandlerID)
		if sp != nil {
			defer sp.Finish()
		}
	}

	// Read the request body
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		h.error(ctx, rw, errors.NewHTTPError(err))
		return
	}

	//Split into batches
	var batchRequests []json.RawMessage
	err = json.Unmarshal(requestBody, &batchRequests)
	if err != nil {
		h.error(ctx, rw, errors.NewHTTPError(errors.Newf(errors.BadInput,
			`Invalid json request`)))
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
	rw.Header().Set(turingReqIDHeaderKey, turingReqID)
	rw.WriteHeader(http.StatusOK)
	batchResponseByte, _ := json.Marshal(batchResponses)
	contentLength, err := rw.Write(batchResponseByte)
	if err != nil {
		ctxLogger.Errorf("Error occurred when copying content: %v", err.Error())
	}
	ctxLogger.Debugf("Written %d bytes", contentLength)
}
