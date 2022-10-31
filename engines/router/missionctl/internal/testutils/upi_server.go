package testutils

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

type GrpcTestServer struct {
}

type HTTPTestServer struct {
}

func (s *GrpcTestServer) PredictValues(
	_ context.Context,
	req *upiv1.PredictValuesRequest,
) (*upiv1.PredictValuesResponse, error) {
	return &upiv1.PredictValuesResponse{PredictionResultTable: req.GetPredictionTable()}, nil
}

func (h *HTTPTestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req := &upiv1.PredictValuesRequest{}
	err = protojson.Unmarshal(requestBody, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := &upiv1.PredictValuesResponse{
		PredictionResultTable: req.PredictionTable,
	}

	resBytes, err := protojson.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(resBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RunTestUPIServer(port int) *grpc.Server {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("%v", err)
	}
	s := grpc.NewServer()
	upiv1.RegisterUniversalPredictionServiceServer(s, &GrpcTestServer{})
	reflection.Register(s)
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return s
}

func RunTestUPIHttpServer(port int) {
	http.Handle("/predict_values", &HTTPTestServer{})
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), http.DefaultServeMux); err != nil {
			log.Fatalf("failed to serve: %s", err)
		}
	}()
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
