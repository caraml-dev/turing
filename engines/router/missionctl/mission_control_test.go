package missionctl

/*
Most tests in this package may be considered integration tests that exercise the full
mission control workflow, which covers the fiber stack and the experimentation engine.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_ "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner/nop"
	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/fiberapi"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	mchttp "github.com/caraml-dev/turing/engines/router/missionctl/server/http"
)

// testHTTPServerAddr is the address of the test HTTP server assumed to be running and
// serving various endpoints required by the router during testing.
var testHTTPServerAddr = "127.0.0.1:9000"

// Test config
var testCfg = &config.Config{
	Port: 80,
	EnrichmentConfig: &config.EnrichmentConfig{
		Endpoint: fmt.Sprintf("http://%s/enrich/", testHTTPServerAddr),
		Timeout:  5 * time.Second,
	},
	RouterConfig: &config.RouterConfig{
		ConfigFile: filepath.Join("testdata", "nop_ensembling_router.yaml"),
		Timeout:    5 * time.Second,
	},
	EnsemblerConfig: &config.EnsemblerConfig{
		Endpoint: fmt.Sprintf("http://%s/ensemble", testHTTPServerAddr),
		Timeout:  5 * time.Second,
	},
	AppConfig: &config.AppConfig{
		FiberDebugLog: false,
	},
}

// Define test suite type
type testSuiteMissionCtl struct {
	routerCfgFilePath    string
	resultFilePath       string
	expCfgResultFilePath string
	compareResultFunc    func(*testing.T, []byte, string)
	compareExpRespFunc   func(*testing.T, []byte, string)
}

////////////////////////////////// Tests //////////////////////////////////////

func TestNewMissionControl(t *testing.T) {
	missionCtl, err := NewMissionControl(
		nil,
		testCfg.EnrichmentConfig,
		testCfg.RouterConfig,
		testCfg.EnsemblerConfig,
		testCfg.AppConfig,
	)
	assert.NoError(t, err)
	assert.Equal(t, true, missionCtl.IsEnricherEnabled())
	assert.Equal(t, true, missionCtl.IsEnsemblerEnabled())
}

func TestMissionControlEnrich(t *testing.T) {
	missionCtl, err := NewMissionControl(
		nil,
		testCfg.EnrichmentConfig,
		testCfg.RouterConfig,
		testCfg.EnsemblerConfig,
		testCfg.AppConfig,
	)
	tu.FailOnError(t, err)

	// Set up Test HTTP Server
	stopServer := startTestHTTPServer(t, testHTTPServerAddr)
	defer stopServer()

	// Enrich
	resp, httpErr := missionCtl.Enrich(context.Background(),
		http.Header{}, []byte(``))

	// Check that the error is nil
	assert.Nil(t, httpErr)
	assert.NotNil(t, resp)

	// Check that the response body is expected
	data, err := tu.ReadFile(filepath.Join("testdata", "enricher_response.json"))
	tu.FailOnError(t, err)
	assert.JSONEq(t, string(data), string(resp.Body()))
}

func TestMakeEnsemblerPayload(t *testing.T) {
	payload1 := []byte(`{"key1": "data1"}`)
	payload2 := []byte(`{"key2": "data2"}`)

	// Make the combined response
	combinedPayload, err := makeEnsemblerPayload(payload1, payload2)

	// Test success
	assert.Nil(t, err)
	assert.JSONEq(t,
		`{"request": {"key1": "data1"}, "response": {"key2": "data2"}}`,
		string(combinedPayload),
	)
}

func TestMissionControlEnsemble(t *testing.T) {
	missionCtl, err := NewMissionControl(
		nil,
		testCfg.EnrichmentConfig,
		testCfg.RouterConfig,
		testCfg.EnsemblerConfig,
		testCfg.AppConfig,
	)
	tu.FailOnError(t, err)

	// Set up Test HTTP Server
	stopServer := startTestHTTPServer(t, testHTTPServerAddr)
	defer stopServer()

	// Ensemble
	resp, httpErr := missionCtl.Ensemble(context.Background(),
		http.Header{}, []byte(``), []byte(``))

	// Check that the error is nil
	assert.Nil(t, httpErr)
	assert.NotNil(t, resp)

	// Check that the response body is expected
	data, err := tu.ReadFile(filepath.Join("testdata", "ensembler_response.json"))
	tu.FailOnError(t, err)
	assert.JSONEq(t, string(data), string(resp.Body()))
}

func TestMissionControlRoute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}

	// Define test suite
	tests := map[string]testSuiteMissionCtl{
		"nop experiment | custom ensembler": {
			routerCfgFilePath:  filepath.Join("testdata", "nop_ensembling_router.yaml"),
			resultFilePath:     filepath.Join("testdata", "nop_ensembling_response.json"),
			compareResultFunc:  compareFanInResponse,
			compareExpRespFunc: assertNil,
		},
		"nop experiment | default ensembler": {
			routerCfgFilePath:  filepath.Join("testdata", "nop_default_router.yaml"),
			resultFilePath:     filepath.Join("testdata", "route_response_route_id_1.json"),
			compareResultFunc:  compareResponse,
			compareExpRespFunc: assertNil,
		},
	}

	// Set up Test HTTP Server
	stopServer := startTestHTTPServer(t, testHTTPServerAddr)
	defer stopServer()

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Create missionctl config
			testCfgLocal := &config.Config{
				Port: 80,
				EnrichmentConfig: &config.EnrichmentConfig{
					Endpoint: "",
					Timeout:  time.Second,
				},
				RouterConfig: &config.RouterConfig{
					ConfigFile: data.routerCfgFilePath,
					Timeout:    3 * time.Second,
				},
				EnsemblerConfig: &config.EnsemblerConfig{
					Endpoint: "",
					Timeout:  time.Second,
				},
			}

			// Init mission control
			missionCtl, err := NewMissionControl(
				nil,
				testCfgLocal.EnrichmentConfig,
				testCfgLocal.RouterConfig,
				testCfgLocal.EnsemblerConfig,
				testCfg.AppConfig,
			)
			tu.FailOnError(t, err)

			// Route
			expResp, resp, httpErr := missionCtl.Route(context.Background(),
				http.Header{}, []byte(`{"customer_id": "2", "country_id": "TH"}`))

			// Check that the error is nil
			assert.Nil(t, httpErr)
			assert.NotNil(t, resp)

			// Check that the response header has application/json content type
			assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
			// Check that the response body is not empty
			if resp == nil {
				tu.FailOnError(t, fmt.Errorf("Response body empty"))
			}
			// Compare response body and experiment response with expected
			data.compareResultFunc(t, resp.Body(), data.resultFilePath)
			data.compareExpRespFunc(t, expResp.Configuration, data.expCfgResultFilePath)
		})
	}
}

// compareFanInResponse takes the response body and a file path with expected payload,
// creates a CombinedResponse object for comparison
func compareFanInResponse(t *testing.T, actualResp []byte, expectedFilePath string) {
	// Unmarshal the JSON response, sort the individual responses by treatment name
	// and compare. Expected is already defined in the sorted order.
	var f fiberapi.CombinedResponse
	tu.FailOnError(t, json.Unmarshal(actualResp, &f))
	sort.Slice(f.RouteResponses, func(i, j int) bool {
		return f.RouteResponses[i].Route < f.RouteResponses[j].Route
	})
	fBytes, err := json.Marshal(f)
	tu.FailOnError(t, err)

	// Check that the response body is expected
	data, err := tu.ReadFile(expectedFilePath)
	tu.FailOnError(t, err)
	assert.JSONEq(t, string(data), string(fBytes))
}

// compareResponse compares the actual response with the expected response
// defined in the input file path
func compareResponse(t *testing.T, actualResp []byte, expectedFilePath string) {
	data, err := tu.ReadFile(expectedFilePath)
	tu.FailOnError(t, err)
	// Compare
	assert.JSONEq(t, string(data), string(actualResp))
}

// assertNil compares that the actual response is nil
func assertNil(t *testing.T, actualResp []byte, _ string) {
	assert.Nil(t, actualResp)
}

// ////////////////////// Benchmark Tests //////////////////////////////////////
// Global variables for benchmark tests, to ensure compiler optimization doesn't
// eliminate function calls
var benchMarkResp mchttp.Response
var benchMarkHTTPErr *errors.TuringError

func benchmarkEnrich(payloadFileName string, b *testing.B) {
	missionCtl, _ := NewMissionControl(
		nil,
		testCfg.EnrichmentConfig,
		testCfg.RouterConfig,
		testCfg.EnsemblerConfig,
		testCfg.AppConfig,
	)

	// Set up Test HTTP Server
	stopServer := startTestHTTPServer(b, testHTTPServerAddr)
	defer stopServer()

	// Read payload
	payload, err := tu.ReadFile(filepath.Join("testdata", payloadFileName))
	if err != nil {
		b.FailNow()
	}

	// Enrich, measure call
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchMarkResp, benchMarkHTTPErr = missionCtl.Enrich(context.Background(),
			http.Header{}, payload)
	}
}

func BenchmarkMissionControlEnrichSmallPayload(b *testing.B) {
	benchmarkEnrich("small_payload.json", b)
}
func BenchmarkMissionControlEnrichMediumPayload(b *testing.B) {
	benchmarkEnrich("medium_payload.json", b)
}
func BenchmarkMissionControlEnrichLargePayload(b *testing.B) {
	benchmarkEnrich("large_payload.json", b)
}

func benchmarkRoute(cfgFileName string, payloadFileName string, b *testing.B) {
	// Create missionctl config
	testCfgLocal := &config.Config{
		Port: 80,
		EnrichmentConfig: &config.EnrichmentConfig{
			Endpoint: "",
			Timeout:  time.Second,
		},
		RouterConfig: &config.RouterConfig{
			ConfigFile: filepath.Join("testdata", cfgFileName),
			Timeout:    time.Second,
		},
		EnsemblerConfig: &config.EnsemblerConfig{
			Endpoint: "",
			Timeout:  time.Second,
		},
	}

	// Init mission control
	mc, _ := NewMissionControl(
		nil,
		testCfgLocal.EnrichmentConfig,
		testCfgLocal.RouterConfig,
		testCfgLocal.EnsemblerConfig,
		testCfg.AppConfig,
	)

	// Set up Test HTTP Server
	stopServer := startTestHTTPServer(b, testHTTPServerAddr)
	defer stopServer()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, benchMarkResp, benchMarkHTTPErr = mc.Route(context.Background(),
			http.Header{}, []byte(filepath.Join("testdata", payloadFileName)))
	}
}

func BenchmarkMissionControlDefaultRouteSmallPayload(b *testing.B) {
	benchmarkRoute("nop_default_router.yaml", "small_payload.json", b)
}
func BenchmarkMissionControlEnsemblingRouteSmallPayload(b *testing.B) {
	benchmarkRoute("nop_ensembling_router.yaml", "small_payload.json", b)
}
func BenchmarkMissionControlDefaultRouteLargePayload(b *testing.B) {
	benchmarkRoute("nop_default_router.yaml", "large_payload.json", b)
}
func BenchmarkMissionControlEnsemblingRouteLargePayload(b *testing.B) {
	benchmarkRoute("nop_ensembling_router.yaml", "large_payload.json", b)
}

//////////////////////// Test HTTP Server Methods /////////////////////////////

// startTestHTTPServer starts an HTTP server at the configured address. It returns
// a stopServer function that the caller should call to stop the server after each
// test completes, so that the TCP socket can be reused.
//
// The server handles requests with the following paths:
// - /enrich/
// - /ensemble/
// - /route1/
// - /control/
//
// This test server can be used to test sending requests to various endpoints configured
// in the router.
func startTestHTTPServer(t testing.TB, addr string) (stopServer func()) {
	handler := http.NewServeMux()
	handler.HandleFunc("/enrich/", enricherHandler)
	handler.HandleFunc("/ensemble/", ensemblerHandler)
	handler.HandleFunc("/route1/", route1Handler)
	handler.HandleFunc("/control/", controlHandler)
	server := httptest.NewUnstartedServer(handler)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal("Failed to start test http server: " + err.Error())
	}

	server.Listener = listener
	server.Start()

	return func() {
		server.Close()
	}
}

// Define HTTP Server handlers for the enricher, ensembler and route endpoints

func enricherHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(`{"customer_id": "1230"}`))
}

func ensemblerHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(`{"result": "ensembled"}`))
}

func route1Handler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(`{"version": 1}`))
}

func controlHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(`{"version": 2}`))
}
