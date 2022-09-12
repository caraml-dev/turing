package testutils

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcTestServer struct {
	Port         int
	MockResponse *upiv1.PredictValuesResponse
	DelayTimer   time.Duration
}

var defaultMockResponse = &upiv1.PredictValuesResponse{
	PredictionResultTable: &upiv1.Table{
		Name:    "table",
		Columns: nil,
		Rows:    nil,
	},
	Metadata: &upiv1.ResponseMetadata{
		PredictionId: "123",
		ExperimentId: "2",
	},
}

func (s *GrpcTestServer) PredictValues(
	_ context.Context,
	_ *upiv1.PredictValuesRequest,
) (*upiv1.PredictValuesResponse, error) {
	time.Sleep(s.DelayTimer)

	if s.MockResponse != nil {
		return s.MockResponse, nil
	}

	return defaultMockResponse, nil
}

func RunTestUPIServer(srv GrpcTestServer) *grpc.Server {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", srv.Port))
	if err != nil {
		log.Fatalf("%v", err)
	}
	s := grpc.NewServer()
	upiv1.RegisterUniversalPredictionServiceServer(s, &srv)
	reflection.Register(s)
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return s
}

func GetDefaultMockResponse() *upiv1.PredictValuesResponse {
	return defaultMockResponse
}
