package upi

import (
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	fiberGrpc "github.com/gojek/fiber/grpc"
)

type Request struct {
	*fiberGrpc.Request
	RequestProto *upiv1.PredictValuesRequest
}
