package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gojek/turing/engines/router/missionctl"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	mchttp "github.com/gojek/turing/engines/router/missionctl/http"
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

type batchResult struct {
	BatchResult []batchResponse `json:"batch_result"`
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

	ctx = h.enableTracingSpan(ctx, req)

	// Read the request body
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		h.error(ctx, rw, errors.NewHTTPError(err))
		return
	}

	//Valid request
	if valid := json.Valid(requestBody); !valid {
		h.error(ctx, rw, errors.NewHTTPError(errors.Newf(errors.BadInput,
			`Invalid json request format`)))
		return
	}

	//Split into batches
	var batchRequests map[string][]interface{}
	err = json.Unmarshal(requestBody, &batchRequests)
	if err != nil {
		h.error(ctx, rw, errors.NewHTTPError(errors.Newf(errors.BadInput,
			`Invalid json request`)))
		return
	}

	//Validate request
	if _, ok := batchRequests["batch_request"]; !ok {
		h.error(ctx, rw, errors.NewHTTPError(errors.Newf(errors.BadInput,
			`batch_request" not found in request`)))
		return
	}

	//Handle request
	var resp mchttp.Response
	var batchResponses []batchResponse
	for _, v := range batchRequests["batch_request"] {
		batchRequestBody, _ := json.Marshal(v)
		var batchResponse batchResponse
		resp, httpErr = h.getPrediction(ctx, req, ctxLogger, batchRequestBody)
		if httpErr != nil {
			batchResponse.StatusCode = httpErr.Code
			batchResponse.ErrorMsg = httpErr.Message
			batchResponses = append(batchResponses, batchResponse)
			continue
		}
		err = json.Unmarshal(resp.Body(), &batchResponse.Data)
		if err != nil {
			batchResponse.StatusCode = http.StatusInternalServerError
			batchResponse.ErrorMsg = "Unable to marshall response into json"
		}
		batchResponse.StatusCode = http.StatusOK
		batchResponses = append(batchResponses, batchResponse)
	}

	// Write the json response to the writer
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set(turingReqIDHeaderKey, turingReqID)
	rw.WriteHeader(http.StatusOK)
	batchResponseByte, _ := json.Marshal(batchResult{BatchResult: batchResponses})
	contentLength, err := rw.Write(batchResponseByte)
	if err != nil {
		ctxLogger.Errorf("Error occurred when copying content: %v", err.Error())
	}
	ctxLogger.Debugf("Written %d bytes", contentLength)
}
