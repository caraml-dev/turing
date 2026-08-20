package tracing

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopMethods(t *testing.T) {
	tr := newNopTracer()

	assert.Equal(t, false, tr.IsEnabled())

	sp, ctx := tr.StartSpanFromRequestHeader(context.Background(), "test", http.Header{})
	assert.NotNil(t, sp)
	assert.NotNil(t, ctx)

	sp, ctx = tr.StartSpanFromContext(context.Background(), "test")
	assert.NotNil(t, sp)
	assert.NotNil(t, ctx)
}
