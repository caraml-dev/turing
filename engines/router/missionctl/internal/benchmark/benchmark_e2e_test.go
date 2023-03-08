package benchmark

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	_ "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner/nop"
	"github.com/caraml-dev/turing/engines/router/missionctl"
	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/http/handlers"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/upi"
)

const (
	turingUpiGRCPPort = 50603
	turingUpiHTTPPort = 9003
	routeGRPCPort     = 50604
	routeHTTPPort     = 9004
	// require running port at 50603
	singleUpiGRPCRouteServer = "../../testdata/benchmark/single_route_upi_grpc.yaml"
	// require running endpoint at localhost:9004/predict_values/
	singleUpiHTTPRouteServer = "../../testdata/benchmark/single_route_upi_http.yaml"
)

// prevent compiler optimization by setting benchmark function to global var
var benchMarkUpiResp *upiv1.PredictValuesResponse
var benchMarkHTTPResp *http.Response
var benchMarkUpiErr error

func TestMain(m *testing.M) {
	// Run route server
	testutils.RunTestUPIServer(routeGRPCPort)
	testutils.RunTestUPIHttpServer(routeHTTPPort)
	// Run Turing router
	runTuringUpiGRPCServer(turingUpiGRCPPort)
	runTuringUpiHTTPServer(turingUpiHTTPPort)
	os.Exit(m.Run())
}

func runTuringUpiGRPCServer(port int) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Glob().Panicf("failed to listen on port: %v", port)
	}
	mc, err := missionctl.NewMissionControlUPI(
		singleUpiGRPCRouteServer,
		false,
	)
	if err != nil {
		log.Glob().Panicf("failed to create mc: %v", err.Error())
	}
	rl := resultlog.InitTuringResultLogger("", resultlog.NewNopLogger())
	upiResultLogger, err := resultlog.InitUPIResultLogger("route-name-3.project", config.NopLogger, nil, rl)
	if err != nil {
		log.Glob().Panicf("failed to create upi result logger: %v", err.Error())
	}
	upiServer := upi.NewUPIServer(mc, upiResultLogger)
	go upiServer.Run(l)
}

func runTuringUpiHTTPServer(port int) {
	testCfg := &config.Config{
		Port: port,
		RouterConfig: &config.RouterConfig{
			ConfigFile: singleUpiHTTPRouteServer,
			Timeout:    5 * time.Second,
		},
		EnrichmentConfig: &config.EnrichmentConfig{},
		EnsemblerConfig:  &config.EnsemblerConfig{},
		AppConfig: &config.AppConfig{
			FiberDebugLog: false,
		},
	}

	// Init mission control
	mc, err := missionctl.NewMissionControl(
		nil,
		testCfg.EnrichmentConfig,
		testCfg.RouterConfig,
		testCfg.EnsemblerConfig,
		testCfg.AppConfig,
	)
	if err != nil {
		log.Glob().Fatalf("fail to create mission control: %v", err.Error())
	}
	rl := resultlog.InitTuringResultLogger("", resultlog.NewNopLogger())
	http.Handle("/v1/predict", handlers.NewHTTPHandler(mc, rl))
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", testCfg.Port), http.DefaultServeMux); err != nil {
			log.Glob().Fatalf("failed to serve: %s", err)
		}
	}()
}

func generatedUpiClientCall(rows int, cols int, port int, b *testing.B) {
	upiRequest := testutils.GenerateUPIRequest(rows, cols)
	conn, err := grpc.Dial(fmt.Sprintf(":%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(b, err)
	client := upiv1.NewUniversalPredictionServiceClient(conn)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchMarkUpiResp, benchMarkUpiErr = client.PredictValues(context.Background(), upiRequest)
	}
}
func httpClientCall(rows int, cols int, endpoint string, b *testing.B) {
	upiRequest := testutils.GenerateUPIRequest(rows, cols)
	resBytes, err := protojson.Marshal(upiRequest)
	require.NotNil(b, resBytes)
	b.ResetTimer()
	var resp *http.Response
	for n := 0; n < b.N; n++ {
		resp, err = http.Post(endpoint, "application/json", io.NopCloser(bytes.NewBuffer(resBytes)))
		resp.Body.Close()
	}
	benchMarkHTTPResp = resp
	benchMarkUpiErr = err
}

func Benchmark_Upi_Grpc_Direct_Small(b *testing.B) {
	generatedUpiClientCall(5, 5, routeGRPCPort, b)
}

func Benchmark_Upi_Grpc_Direct_Large(b *testing.B) {
	generatedUpiClientCall(100, 100, routeGRPCPort, b)
}

func Benchmark_Upi_Grpc_Turing_Small(b *testing.B) {
	generatedUpiClientCall(5, 5, turingUpiGRCPPort, b)
}

func Benchmark_Upi_Grpc_Turing_Large(b *testing.B) {
	generatedUpiClientCall(100, 100, turingUpiGRCPPort, b)
}

func Benchmark_Upi_Http_Direct_Small(b *testing.B) {
	httpClientCall(5, 5, fmt.Sprintf("http://localhost:%d/predict_values", routeHTTPPort), b)
}

func Benchmark_Upi_Http_Direct_Large(b *testing.B) {
	httpClientCall(100, 100, fmt.Sprintf("http://localhost:%d/predict_values", routeHTTPPort), b)
}

func Benchmark_Upi_Http_Turing_Small(b *testing.B) {
	httpClientCall(5, 5, fmt.Sprintf("http://localhost:%d/v1/predict", turingUpiHTTPPort), b)
}

func Benchmark_Upi_Http_Turing_Large(b *testing.B) {
	httpClientCall(100, 100, fmt.Sprintf("http://localhost:%d/v1/predict", turingUpiHTTPPort), b)
}
