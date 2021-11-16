package handlers

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/gojek/turing/engines/experiment/runner/nop"
	"github.com/gojek/turing/engines/router/missionctl"
	"github.com/gojek/turing/engines/router/missionctl/config"
	"github.com/stretchr/testify/assert"
)

func TestNewBatchHTTPHandler(t *testing.T) {
	missionCtl := createGenericMissionControl(t)
	batchHTTPHandler := NewBatchHTTPHandler(missionCtl)
	assert.NotNil(t, batchHTTPHandler)

	//Create test routes endpoint. Route will write request body as response
	testRouterHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestBody, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)
		_, err = writer.Write(requestBody)
		assert.NoError(t, err)
	})
	server := createRouteServer(t, testRouterHandler)
	server.Start()
	defer server.Close()

	tests := map[string]struct {
		payload              string
		expectedResponseBody string
		expectedStatusCode   int
		expectedContentType  string
	}{
		"ok request": {
			payload: `{"batch_request": [
							{ "request1": "value1" },
							{ "request2": "value2" }
						]}`,
			expectedResponseBody: `{"batch_result": [
										{ "code": 200, 
										  "data": {"request1": "value1"}
										},
										{ "code": 200, 
										  "data": {"request2": "value2"}
										}
									]}`,
			expectedStatusCode:  200,
			expectedContentType: "application/json",
		},
		"invalid json": {
			payload:              `{"batch_request": [{ : }]}`,
			expectedResponseBody: "Invalid json request format\n",
			expectedStatusCode:   400,
			expectedContentType:  "text/plain; charset=utf-8",
		},
		"invalid json request": {
			payload:              `{"key": "value1"}`,
			expectedResponseBody: "Invalid json request\n",
			expectedStatusCode:   400,
			expectedContentType:  "text/plain; charset=utf-8",
		},
		"invalid json request, no batch_request key": {
			payload:              `{"key": [{"key1":"value1"}]}`,
			expectedResponseBody: "batch_request\" not found in request\n",
			expectedStatusCode:   400,
			expectedContentType:  "text/plain; charset=utf-8",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/batch_predict", bytes.NewBuffer([]byte(test.payload)))
			w := httptest.NewRecorder()
			batchHTTPHandler.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()
			result, _ := ioutil.ReadAll(res.Body)

			assert.Equal(t, test.expectedStatusCode, res.StatusCode)
			assert.Equal(t, test.expectedContentType, res.Header.Get("Content-Type"))
			if test.expectedContentType == "application/json" {
				assert.JSONEq(t, test.expectedResponseBody, string(result))
			} else {
				assert.Equal(t, test.expectedResponseBody, string(result))
			}
		})
	}
}

func TestNewBatchHTTPHandlerBadRoute(t *testing.T) {
	missionCtl := &MockMissionControlBadRoute{BaseMockMissionControl: *createTestBaseMissionControl()}
	batchHTTPHandler := NewBatchHTTPHandler(missionCtl)
	assert.NotNil(t, batchHTTPHandler)

	tests := map[string]struct {
		payload              string
		expectedResponseBody string
		expectedStatusCode   int
		expectedContentType  string
	}{
		"ok request": {
			payload: `{"batch_request": [
							{ "request1": "value1" },
							{ "request2": "value2" }
						]}`,
			expectedResponseBody: `{"batch_result": [
										{ "code": 500, 
										  "error": "Bad Route Called"
										},
										{ "code": 500, 
										  "error": "Bad Route Called"
										}
									]}`,
			expectedStatusCode:  200,
			expectedContentType: "application/json",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/batch_predict", bytes.NewBuffer([]byte(test.payload)))
			w := httptest.NewRecorder()
			batchHTTPHandler.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()
			result, _ := ioutil.ReadAll(res.Body)

			assert.Equal(t, test.expectedStatusCode, res.StatusCode)
			assert.Equal(t, test.expectedContentType, res.Header.Get("Content-Type"))
			assert.JSONEq(t, test.expectedResponseBody, string(result))
		})
	}
}

func createRouteServer(t *testing.T, testRouterHandler http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("/route1", testRouterHandler)
	server := httptest.NewUnstartedServer(mux)
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatal("Failed to start test http server: " + err.Error())
	}
	server.Listener = listener
	return server
}

func createGenericMissionControl(t *testing.T) missionctl.MissionControl {
	//Create missionctl with route for testing
	missionCtl, err := missionctl.NewMissionControl(
		nil,
		&config.EnrichmentConfig{
			Endpoint: "",
			Timeout:  time.Second,
		},
		&config.RouterConfig{
			//ConfigFile: filepath.Join("../testdata", "nop_default_router.yaml"),
			ConfigFile: filepath.Join("../testdata", "batch_router_test.yaml"),
			Timeout:    10 * time.Second,
		},
		&config.EnsemblerConfig{
			Endpoint: "",
			Timeout:  time.Second,
		},
		&config.AppConfig{
			FiberDebugLog: false,
		},
	)
	assert.Nil(t, err)
	return missionCtl
}
