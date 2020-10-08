package missionctl

/*
Most tests in this package may be considered integration tests that exercise the full
mission control workflow, which covers the fiber stack and the experimentation engine.
Some tests work with Litmus/XP which requires the respective passkeys
to be specified - these tests are skipped when running `make test` as it uses the -short
flag. They may instead be run with `make testall`.
If any of the endpoints need to be changed, the values in the testdata/ folder must be
updated accordingly. For the Litmus tests, the LITMUS_PASSKEY env var must be supplied.
Similarly, for the XP tests, the XP_PASSKEY env var must be specified.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	_ "github.com/gojek/turing/engines/experiment/runner/nop"
	"github.com/gojek/turing/engines/router/missionctl/config"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/fiberapi"
	mchttp "github.com/gojek/turing/engines/router/missionctl/http"
	tu "github.com/gojek/turing/engines/router/missionctl/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// Test config
var testCfg *config.Config = &config.Config{
	Port: 80,
	EnrichmentConfig: &config.EnrichmentConfig{
		Endpoint: "http://localhost:9000/enrich/",
		Timeout:  5 * time.Second,
	},
	RouterConfig: &config.RouterConfig{
		ConfigFile: filepath.Join("testdata", "nop_ensembling_router.yaml"),
		Timeout:    5 * time.Second,
	},
	EnsemblerConfig: &config.EnsemblerConfig{
		Endpoint: "http://localhost:9000/ensemble/",
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
	srv := startTestHTTPServer()
	defer func() {
		_ = srv.Shutdown(context.Background())
	}()

	// Enrich
	resp, httpErr := missionCtl.Enrich(context.Background(),
		http.Header{}, []byte(``))

	// Check that the error is nil
	assert.True(t, httpErr == nil)

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
	srv := startTestHTTPServer()
	defer func() {
		_ = srv.Shutdown(context.Background())
	}()

	// Ensemble
	resp, httpErr := missionCtl.Ensemble(context.Background(),
		http.Header{}, []byte(``), []byte(``))

	// Check that the error is nil
	assert.Equal(t, true, httpErr == nil)

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
	srv := startTestHTTPServer()
	defer func() {
		_ = srv.Shutdown(context.Background())
	}()

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
			assert.Equal(t, true, httpErr == nil)
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

//////////////////////// Benchmark Tests //////////////////////////////////////
// Global variables for benchmark tests, to ensure compiler optimization doesn't
// eliminate function calls
var benchMarkResp mchttp.Response
var benchMarkHTTPErr *errors.HTTPError

func benchmarkEnrich(payloadFileName string, b *testing.B) {
	missionCtl, _ := NewMissionControl(
		nil,
		testCfg.EnrichmentConfig,
		testCfg.RouterConfig,
		testCfg.EnsemblerConfig,
		testCfg.AppConfig,
	)

	// Set up Test HTTP Server
	srv := startTestHTTPServer()
	defer func() {
		_ = srv.Shutdown(context.Background())
	}()

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
	srv := startTestHTTPServer()
	defer func() {
		_ = srv.Shutdown(context.Background())
	}()

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
// Register HTTP handlers once
var once sync.Once

func registerHTTPHandlers() {
	http.HandleFunc("/enrich/", enricherHandler)
	http.HandleFunc("/ensemble/", ensemblerHandler)
	http.HandleFunc("/route1/", route1Handler)
	http.HandleFunc("/control/", controlHandler)
}

func startTestHTTPServer() *http.Server {
	srv := &http.Server{}

	go func() {
		once.Do(registerHTTPHandlers)
		_ = http.ListenAndServe(":9000", nil)
	}()

	return srv
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
