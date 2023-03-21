package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/turing/engines/router/missionctl"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/experiment"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	mchttp "github.com/caraml-dev/turing/engines/router/missionctl/server/http"

	fiberProtocol "github.com/gojek/fiber/protocol"
)

// testBody is a simple struct with a string, for testing json payload in requests
type testBody struct {
	Value string `json:"value"`
}

// testSuiteHTTPHandler is used to create the test suite
type testSuiteHTTPHandler struct {
	mc     missionctl.MissionControl
	checks func()
}

// Define Mock Types //////////////////////////////////////////////////////////

// BaseMockMissionControl is a mock implementation for the missionctl.MissionControl interface
type BaseMockMissionControl struct {
	mock.Mock
}

// IsEnricherEnabled always returns true
func (mc *BaseMockMissionControl) IsEnricherEnabled() bool {
	return true
}

// IsEnsemblerEnabled always returns true
func (mc *BaseMockMissionControl) IsEnsemblerEnabled() bool {
	return true
}

// Enrich appends ":Enrich" to the value in the json payload
func (mc *BaseMockMissionControl) Enrich(
	ctx context.Context,
	header http.Header,
	body []byte,
) (mchttp.Response, *errors.TuringError) {
	mc.Called()
	return modifyRequestBody(body, map[string]string{"Enricher": "value"}, "Enrich")
}

// Route appends ":Route" to the value in the json payload
func (mc *BaseMockMissionControl) Route(
	ctx context.Context,
	header http.Header,
	body []byte,
) (*experiment.Response, mchttp.Response, *errors.TuringError) {
	mc.Called()
	resp, err := modifyRequestBody(body, map[string]string{}, "Route")
	return nil, resp, err
}

// Ensemble appends ":Ensemble" to the value in the json payload
func (mc *BaseMockMissionControl) Ensemble(
	ctx context.Context,
	header http.Header,
	requestBody []byte,
	routerResponse []byte,
) (mchttp.Response, *errors.TuringError) {
	mc.Called(header)
	return modifyRequestBody(routerResponse, map[string]string{}, "Ensemble")
}

// MockMissionControl simply inherits from BaseMockMissionControl
type MockMissionControl struct {
	BaseMockMissionControl
}

// MockMissionControlBadEnrich inherits from BaseMockMissionControl
// and provides an override for Enrich
type MockMissionControlBadEnrich struct {
	BaseMockMissionControl
}

// Enrich always returns an error
func (mc *MockMissionControlBadEnrich) Enrich(
	ctx context.Context,
	header http.Header,
	body []byte,
) (mchttp.Response, *errors.TuringError) {
	mc.Called()
	return nil, errors.NewTuringError(fmt.Errorf("Bad Enrich Called"), fiberProtocol.HTTP)
}

// MockMissionControlBadRoute inherits from BaseMockMissionControl
// and provides an override for Route
type MockMissionControlBadRoute struct {
	BaseMockMissionControl
}

// Route always returns an error
func (mc *MockMissionControlBadRoute) Route(
	ctx context.Context,
	header http.Header,
	body []byte,
) (*experiment.Response, mchttp.Response, *errors.TuringError) {
	mc.Called()
	return nil, nil, errors.NewTuringError(fmt.Errorf("Bad Route Called"), fiberProtocol.HTTP)
}

// MockMissionControlBadEnsemble inherits from BaseMockMissionControl
// and provides an override for Ensemble
type MockMissionControlBadEnsemble struct {
	BaseMockMissionControl
}

// Ensemble always returns an error
func (mc *MockMissionControlBadEnsemble) Ensemble(
	ctx context.Context,
	header http.Header,
	requestBody []byte,
	routerResponse []byte,
) (mchttp.Response, *errors.TuringError) {
	mc.Called()
	return nil, errors.NewTuringError(fmt.Errorf("Bad Ensemble Called"), fiberProtocol.HTTP)
}

// Tests //////////////////////////////////////////////////////////////////////

// TestHTTPService tests the successful sequence of Enrich -> Route -> Ensemble
func TestHTTPService(t *testing.T) {
	expectedResponse := string(`{"value": "Init:Enrich:Route:Ensemble"}`)
	requestPayload, err := json.Marshal(testBody{Value: "Init"})
	tu.FailOnError(t, err)

	// Create mock mission control, set up expectations
	mc := &MockMissionControl{BaseMockMissionControl: *createTestBaseMissionControl()}
	// Create new response recorder
	rr := httptest.NewRecorder()
	// Create request
	req := createTestRequest([]byte(requestPayload), t)
	// Make request
	doTestRequest(mc, req, rr)

	// Check response status code is Success
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code mismatch. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	// Check the result body is expected
	assert.JSONEq(t, expectedResponse, rr.Body.String(), "Response body mismatch.")

	// Check that ensembler was called with the expected headers
	mc.AssertCalled(t, "Ensemble",
		http.Header{"Context-Type": []string{"application/json"}, "Enricher": []string{"value"},
			"Turing-Req-Id": []string{rr.Header().Get("Turing-Req-Id")}})
}

// TestHTTPServiceBadRequest tests for a HTTP InternalServerError on bad
// Enrich / Route / Ensemble, with unclassified.
func TestHTTPServiceBadRequest(t *testing.T) {
	// Initialize mocks
	mcBadEnrich := &MockMissionControlBadEnrich{
		BaseMockMissionControl: *createTestBaseMissionControl(),
	}
	mcBadRoute := &MockMissionControlBadRoute{
		BaseMockMissionControl: *createTestBaseMissionControl(),
	}
	mcBadEnsemble := &MockMissionControlBadEnsemble{
		BaseMockMissionControl: *createTestBaseMissionControl(),
	}

	// Define test suite
	tests := map[string]testSuiteHTTPHandler{
		"bad enrich": {
			mc: mcBadEnrich,
			checks: func() {
				mcBadEnrich.AssertCalled(t, "Enrich")
				mcBadEnrich.AssertNotCalled(t, "Route")
				mcBadEnrich.AssertNotCalled(t, "Ensemble")
			},
		},
		"bad route": {
			mc: mcBadRoute,
			checks: func() {
				mcBadRoute.AssertCalled(t, "Enrich")
				mcBadRoute.AssertCalled(t, "Route")
				mcBadRoute.AssertNotCalled(t, "Ensemble")
			},
		},
		"bad ensemble": {
			mc: mcBadEnsemble,
			checks: func() {
				mcBadEnsemble.AssertCalled(t, "Enrich")
				mcBadEnsemble.AssertCalled(t, "Route")
				mcBadEnsemble.AssertCalled(t, "Ensemble")
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Create new response recorder
			rr := httptest.NewRecorder()
			// Create request with malformed json payload
			req := createTestRequest([]byte(`{"value": "test"}`), t)
			// Make request
			doTestRequest(data.mc, req, rr)

			// Check response status code indicates Internal Server Error
			if status := rr.Code; status != http.StatusInternalServerError {
				t.Errorf("Status code mismatch. Expected %d.\n Got %d instead",
					http.StatusInternalServerError, status)
			}

			// Assert that the expected function calls occurred
			data.checks()
		})
	}
}

func createTestBaseMissionControl() *BaseMockMissionControl {
	mc := &BaseMockMissionControl{}
	mc.On("Enrich").Return(nil)
	mc.On("Route").Return(nil)
	mc.On("Ensemble", mock.Anything).Return(nil)
	return mc
}

func createTestRequest(payload []byte, t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(payload))
	tu.FailOnError(t, err)
	return req
}

func doTestRequest(mc missionctl.MissionControl, req *http.Request, rr *httptest.ResponseRecorder) {
	handler := NewHTTPHandler(mc, resultlog.InitTuringResultLogger("", resultlog.NewNopLogger()))
	http.HandlerFunc(handler.ServeHTTP).ServeHTTP(rr, req)
}

func modifyRequestBody(
	body []byte,
	responseHeaders map[string]string,
	caller string,
) (mchttp.Response, *errors.TuringError) {
	// Parse the body
	var t testBody
	err := json.Unmarshal(body, &t)
	if err != nil {
		return nil, errors.NewTuringError(fmt.Errorf("Error occurred in %s: %v", caller, err), fiberProtocol.HTTP)
	}

	// Append to Value
	t.Value = fmt.Sprintf("%s:%s", t.Value, caller)

	// Convert the data back to string
	tBytes, err := json.Marshal(t)
	if err != nil {
		return nil, errors.NewTuringError(fmt.Errorf("Error occurred in %s: %v", caller, err), fiberProtocol.HTTP)
	}

	httpHeader := http.Header{
		"Context-Type": []string{"application/json"},
	}
	for key, value := range responseHeaders {
		httpHeader.Set(key, value)
	}

	// Return response
	httpResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBuffer(tBytes)),
		Header:     httpHeader,
	}
	mcResp, err := mchttp.NewCachedResponseFromHTTP(httpResponse)
	if err != nil {
		return nil, errors.NewTuringError(err, fiberProtocol.HTTP)
	}
	return mcResp, nil
}
