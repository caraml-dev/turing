package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

func TestGetGlob(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	globalTracer = &NopTracer{}
	assert.Equal(t, globalTracer, Glob())
}

func TestSetGlob(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	tr := &NopTracer{}
	SetGlob(tr)
	assert.Equal(t, tr, globalTracer)
}

func TestInitGlobalTracerNop(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	_, err := InitGlobalTracer("test", &config.JaegerConfig{})
	assert.NoError(t, err)
	assert.Equal(t, false, globalTracer.IsEnabled())
}

func TestInitGlobalTracerOtel(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	shutdown, err := InitGlobalTracer("test", &config.JaegerConfig{
		Enabled:           true,
		CollectorEndpoint: "http://localhost:4318",
	})
	assert.NoError(t, err)
	assert.Equal(t, true, globalTracer.IsEnabled())
	assert.NoError(t, shutdown(context.Background()))
}

// TestInitGlobalTracerOtel_ErrorReturnsNonNilShutdown ensures that even when tracer
// initialisation fails (e.g. an invalid CollectorEndpoint), callers get back a safe,
// callable no-op ShutdownFunc rather than nil. Package-level InitGlobalTracer is called
// from application.go, which currently relies on log.Glob().Fatalf to exit the process on
// error, but the contract shouldn't depend on that -- a nil ShutdownFunc would panic any
// caller that unconditionally defers it.
func TestInitGlobalTracerOtel_ErrorReturnsNonNilShutdown(t *testing.T) {
	tempTracer := globalTracer
	defer func() { globalTracer = tempTracer }()

	shutdown, err := InitGlobalTracer("test", &config.JaegerConfig{
		Enabled:           true,
		CollectorEndpoint: "",
	})
	assert.Error(t, err)
	require.NotNil(t, shutdown)
	assert.NoError(t, shutdown(context.Background()))
}
