package testutils

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcTestServer struct {
	Port         int
	MockResponse *upiv1.PredictValuesResponse
	DelayTimer   time.Duration
}

func (s *GrpcTestServer) PredictValues(
	_ context.Context,
	req *upiv1.PredictValuesRequest,
) (*upiv1.PredictValuesResponse, error) {
	time.Sleep(s.DelayTimer)

	if s.MockResponse != nil {
		return s.MockResponse, nil
	}

	return &upiv1.PredictValuesResponse{PredictionResultTable: req.GetPredictionTable()}, nil
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

func GenerateUPIRequest(n int, m int) *upiv1.PredictValuesRequest {
	rows := make([]*upiv1.Row, n)
	columns := make([]*upiv1.Column, m)
	values := make([]*upiv1.Value, m)

	for i := 0; i < m; i++ {
		columns[i] = &upiv1.Column{
			Name: "Feature " + strconv.Itoa(i),
			Type: upiv1.Type(i % 4),
		}
		var val *upiv1.Value
		switch upiv1.Type(i % 4) {
		case upiv1.Type_TYPE_UNSPECIFIED, upiv1.Type_TYPE_DOUBLE:
			val = &upiv1.Value{
				DoubleValue: 1.234 * float64(i),
			}
		case upiv1.Type_TYPE_INTEGER:
			val = &upiv1.Value{
				IntegerValue: int64(i),
			}
		case upiv1.Type_TYPE_STRING:
			val = &upiv1.Value{
				StringValue: strconv.Itoa(i),
			}
		}
		values[i] = val
	}

	for i := 0; i < n; i++ {
		rows[i] = &upiv1.Row{
			RowId:  strconv.Itoa(i),
			Values: values,
		}
	}

	return &upiv1.PredictValuesRequest{
		PredictionTable: &upiv1.Table{
			Name:    "generated table",
			Columns: columns,
			Rows:    rows,
		},
		Metadata: &upiv1.RequestMetadata{
			PredictionId: uuid.New().String(),
		},
	}
}
