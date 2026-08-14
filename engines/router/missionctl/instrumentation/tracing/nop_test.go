package tracing

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

func TestNopMethods(t *testing.T) {
	tr := newNopTracer()

	shutdown, err := tr.InitGlobalTracer("test", &config.JaegerConfig{})
	assert.NoError(t, err)
	assert.NoError(t, shutdown(context.Background()))
	assert.Equal(t, false, tr.IsEnabled())

	sp, ctx := tr.StartSpanFromRequestHeader(context.Background(), "test", http.Header{})
	assert.NotNil(t, sp)
	assert.NotNil(t, ctx)

	sp, ctx = tr.StartSpanFromContext(context.Background(), "test")
	assert.NotNil(t, sp)
	assert.NotNil(t, ctx)
}
