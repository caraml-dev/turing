package handlers

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner/nop"
	"github.com/caraml-dev/turing/engines/router/missionctl"
	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

func TestNewBatchHTTPHandler(t *testing.T) {
	//Create the missionCtl required for test
	configFilePath := filepath.Join("../../../testdata", "batch_router_test.yaml")
	missionCtl, err := createGenericMissionControl(configFilePath)
	require.Nil(t, err)
	mockMissionCtlWithBadRoute := &MockMissionControlBadRoute{BaseMockMissionControl: *createTestBaseMissionControl()}

	//Create test routes endpoint. Route will write request body as response
	testRouterHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestBody, err := io.ReadAll(request.Body)
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
		missionCtl           missionctl.MissionControl
	}{
		"ok request": {
			payload: `[
							{ "request1": "value1" },
							{ "request2": "value2" }
					  ]`,
			expectedResponseBody: `[
										{ "code": 200, 
										  "data": {"request1": "value1"}
										},
										{ "code": 200, 
										  "data": {"request2": "value2"}
										}
								  ]`,
			expectedStatusCode:  200,
			expectedContentType: "application/json",
			missionCtl:          missionCtl,
		},
		"invalid json": {
			payload:              `[{ : }]`,
			expectedResponseBody: "Invalid json request\n",
			expectedStatusCode:   400,
			expectedContentType:  "text/plain; charset=utf-8",
			missionCtl:           missionCtl,
		},
		"invalid json request": {
			payload:              `{"key": "value1"}`,
			expectedResponseBody: "Invalid json request\n",
			expectedStatusCode:   400,
			expectedContentType:  "text/plain; charset=utf-8",
			missionCtl:           missionCtl,
		},
		"ok request bad route": {
			payload: `[
							{ "request1": "value1" },
							{ "request2": "value2" }
					  ]`,
			expectedResponseBody: `[
										{ "code": 500, 
										  "error": "Bad Route Called"
										},
										{ "code": 500, 
										  "error": "Bad Route Called"
										}
								   ]`,
			expectedStatusCode:  200,
			expectedContentType: "application/json",
			missionCtl:          mockMissionCtlWithBadRoute,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			batchHTTPHandler := NewBatchHTTPHandler(test.missionCtl)
			assert.NotNil(t, batchHTTPHandler)

			req := httptest.NewRequest(http.MethodPost, "/v1/batch_predict", bytes.NewBuffer([]byte(test.payload)))
			w := httptest.NewRecorder()
			batchHTTPHandler.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()
			result, _ := io.ReadAll(res.Body)

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

func createGenericMissionControl(configFilePath string) (missionctl.MissionControl, error) {
	//Create missionctl with route for testing
	missionCtl, err := missionctl.NewMissionControl(
		nil,
		&config.EnrichmentConfig{
			Endpoint: "",
			Timeout:  time.Second,
		},
		&config.RouterConfig{
			ConfigFile: configFilePath,
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
	return missionCtl, err
}
