package handlers

import (
	"github.com/gojek/turing/engines/router/missionctl"
	"io/ioutil"
	"net/http"

	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/log"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
)

type batchHttpHandler struct {
	httpHandler
}

// NewBatchHTTPHandler creates an instance of the Mission Control's prediction request handler
func NewBatchHTTPHandler(mc missionctl.MissionControl) http.Handler {
	return &batchHttpHandler{ httpHandler{mc}}
}

func (h *batchHttpHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

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

	ctx = h.enableTracingSpan(ctx, req)

	// Read the request body
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		h.error(ctx, rw, errors.NewHTTPError(err))
		return
	}

	resp, httpErr := h.getPrediction(ctx, req, ctxLogger, requestBody)
	if httpErr != nil {
		h.error(ctx, rw, httpErr)
		return
	}
	payload := resp.Body()

	// Write the json response to the writer
	rw.Header().Set("Content-Type", resp.Header().Get("Content-Type"))
	rw.Header().Set(turingReqIDHeaderKey, turingReqID)
	rw.WriteHeader(http.StatusOK)
	contentLength, err := rw.Write(payload)
	if err != nil {
		ctxLogger.Errorf("Error occurred when copying content: %v", err.Error())
	}
	ctxLogger.Debugf("Written %d bytes", contentLength)
}

