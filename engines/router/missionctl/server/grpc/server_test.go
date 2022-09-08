package grpc

import (
	"context"
	"reflect"
	"testing"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/fiber"
)

func TestUPIServer_PredictValues(t *testing.T) {
	type fields struct {
		UnimplementedUniversalPredictionServiceServer upiv1.UnimplementedUniversalPredictionServiceServer
		fiberRouter                                   fiber.Component
	}
	type args struct {
		parentCtx context.Context
		req       *upiv1.PredictValuesRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *upiv1.PredictValuesResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := &UPIServer{
				UnimplementedUniversalPredictionServiceServer: tt.fields.UnimplementedUniversalPredictionServiceServer,
				fiberRouter: tt.fields.fiberRouter,
			}
			got, err := us.PredictValues(tt.args.parentCtx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("PredictValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PredictValues() got = %v, want %v", got, tt.want)
			}
		})
	}
}
