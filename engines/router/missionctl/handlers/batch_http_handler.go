package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/gojek/turing/engines/router/missionctl"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/gojek/turing/engines/router/missionctl/log"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
)

type batchHTTPHandler struct {
	httpHandler
}

type batchResponse struct {
	StatusCode int         `json:"code"`
	ErrorMsg   string      `json:"error,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

// NewBatchHTTPHandler creates an instance of the Mission Control's prediction request handler
func NewBatchHTTPHandler(mc missionctl.MissionControl) http.Handler {
	return &batchHTTPHandler{httpHandler{mc}}
}

func (h *batchHTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var httpErr *errors.HTTPError
	h.measureRequestDuration(httpErr)

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
		ctx = h.enableTracingSpan(ctx, req)
	}

	// Read the request body
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		h.error(ctx, rw, errors.NewHTTPError(err))
		return
	}

	//Split into batches
	var batchRequests []interface{}
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

		go func(index int, jsonRequestBody interface{}) {
			defer waitGroup.Done()
			batchRequestBody, _ := json.Marshal(jsonRequestBody)
			var batchResponse batchResponse
			resp, httpErr := h.getPrediction(ctx, req, ctxLogger, batchRequestBody)
			if httpErr != nil {
				batchResponse.StatusCode = httpErr.Code
				batchResponse.ErrorMsg = httpErr.Message
				batchResponses[index] = batchResponse
				return
			}
			err := json.Unmarshal(resp.Body(), &batchResponse.Data)
			if err != nil {
				batchResponse.StatusCode = http.StatusInternalServerError
				batchResponse.ErrorMsg = "Unable to marshall response into json"
			}
			batchResponse.StatusCode = http.StatusOK
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
