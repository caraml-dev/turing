package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopMethods(t *testing.T) {
	tr := newNopTracer()

	// Create test context
	testCtx := context.Background()

	// Test methods with return values
	assert.Equal(t, false, tr.IsEnabled())

	closer, err := tr.InitGlobalTracer("", nil)
	assert.NoError(t, err)
	err = closer.Close()
	assert.NoError(t, err)

	sp, ctx := tr.StartSpanFromRequestHeader(testCtx, "", nil)
	assert.Nil(t, sp)
	assert.Equal(t, testCtx, ctx)

	sp, ctx = tr.StartSpanFromContext(testCtx, "")
	assert.Nil(t, sp)
	assert.Equal(t, testCtx, ctx)
}
