package upi

import (
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	fiberGrpc "github.com/gojek/fiber/grpc"
)

type Response struct {
	fiberGrpc.Response
	RequestProto *upiv1.PredictValuesResponse
}
