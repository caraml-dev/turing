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
	PredictionResultRows: []*upiv1.PredictionResultRow{
		{
			RowId: "1",
			Values: []*upiv1.NamedValue{
				{
					Name:        "one",
					Type:        upiv1.NamedValue_TYPE_DOUBLE,
					DoubleValue: 13.3,
				},
				{
					Name:        "two",
					Type:        upiv1.NamedValue_TYPE_STRING,
					StringValue: "23.2",
				},
			},
		},
		{
			RowId: "2",
		},
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

func RunTestUPIServer(srv GrpcTestServer) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", srv.Port))
	if err != nil {
		log.Fatalf("%v", err)
	}
	s := grpc.NewServer()
	upiv1.RegisterUniversalPredictionServiceServer(s, &srv)
	reflection.Register(s)
	log.Printf("Running Test Server at %v", srv.Port)
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

func GetDefaultMockResponse() *upiv1.PredictValuesResponse {
	return defaultMockResponse
}
