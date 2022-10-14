package upi

import (
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	fiberGrpc "github.com/gojek/fiber/grpc"
)

// Request satisfy the fiber.Request interface, by embedding fiber grpc request.
// Additional, the pointer to the initial upi request is saved, so that it can be used
// without unmarshalling from fiber request body which is []byte
type Request struct {
	*fiberGrpc.Request
	RequestProto *upiv1.PredictValuesRequest
}
