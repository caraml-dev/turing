package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopMethods(t *testing.T) {
	tr := newNopTracer()

	// Run through the nop method with no return values and side effects to ensure no error
	tr.SetStartNewSpans(true)

	// Create test context
	testCtx := context.Background()

	// Test methods with return values
	assert.Equal(t, false, tr.IsEnabled())
	assert.Equal(t, false, tr.IsStartNewSpans())
	assert.Equal(t, false, tr.IsStartNewSpans())
	assert.Equal(t, false, tr.IsStartNewSpans())

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
