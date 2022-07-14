// +build tools

package tools

// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

import (
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/envoyproxy/protoc-gen-validate"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
