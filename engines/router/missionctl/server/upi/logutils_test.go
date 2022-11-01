package upi

import (
	"context"
	"fmt"
	"testing"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	fiberProtocol "github.com/gojek/fiber/protocol"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
)

func Test_logTuringRouterRequestSummary(t *testing.T) {
	resp := &upiv1.PredictValuesResponse{}
	key := "test"

	tests := []struct {
		name     string
		err      *errors.TuringError
		expected grpcRouterResponse
	}{
		{
			name: "ok",
			expected: grpcRouterResponse{
				key:  key,
				body: resp,
			},
		},
		{
			name: "error",
			err:  errors.NewTuringError(fmt.Errorf("test error"), fiberProtocol.GRPC),
			expected: grpcRouterResponse{
				key: key,
				err: "test error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := metadata.AppendToOutgoingContext(context.Background(), "test", "key")
			// Make response channel
			respCh := make(chan grpcRouterResponse, 1)
			copyResponseToLogChannel(ctx, respCh, key, resp, tt.err)

			close(respCh)
			data := <-respCh
			require.Equal(t, tt.expected, data)
		})
	}
}
