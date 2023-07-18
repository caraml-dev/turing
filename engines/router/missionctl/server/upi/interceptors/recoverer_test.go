package interceptors

import (
	"context"
	"testing"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/assert"
)

func TestPanicRecovery(t *testing.T) {
	t.Run("Test panic recovery", func(t *testing.T) {
		panicIntercep := PanicRecoveryInterceptor()
		_, err := panicIntercep(
			context.Background(),
			&upiv1.PredictValuesRequest{},
			nil,
			func(ctx context.Context, req interface{}) (interface{}, error) {
				panic("something wrong")
			})
		assert.Equal(t, "panic: something wrong", err.Error())
	})

}
