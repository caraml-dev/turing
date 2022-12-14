package fiberapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/gojek/fiber"
	fiberHttp "github.com/gojek/fiber/http"
	"github.com/opentracing/opentracing-go"
	opentracingLog "github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

// timeInterceptorLog is used to Unmarshal the time log data
type timeInterceptorLog struct {
	Level          string  `json:"level"`
	EventTimestamp float64 `json:"event_timestamp"`
	Caller         string  `json:"caller"`
	Msg            string  `json:"msg"`
	Component      string
	TuringReqID    string `json:"Turing Request"`
	Duration       string
}

// errorInterceptorLog is used to Unmarshal the error log data
type errorInterceptorLog struct {
	Level          string  `json:"level"`
	EventTimestamp float64 `json:"event_timestamp"`
	Caller         string  `json:"caller"`
	Msg            string  `json:"msg"`
	Component      string
	TuringReqID    string `json:"Turing Request"`
	Status         int
	Response       string
}

// Test suite type for error logging interceptor tests
type testSuiteErrLogInterceptor struct {
	kind       fiber.ComponentKind
	respStatus int
	errLog     bool
}

type testSuiteInstrInterceptor struct {
	kind   fiber.ComponentKind
	called bool
}

// mockMetricsCollector satisfies the metrics.Collector interface
type mockMetricsCollector struct {
	mock.Mock
}

func (*mockMetricsCollector) InitMetrics() {}
func (c *mockMetricsCollector) MeasureDurationMsSince(
	key metrics.MetricName,
	starttime time.Time,
	labels map[string]string,
) {
	c.Called(key, starttime, labels)
}
func (*mockMetricsCollector) MeasureDurationMs(
	key metrics.MetricName,
	labels map[string]func() string,
) func() {
	return func() {}
}

// mockSpan satisfies the opentracing.Span interface
type mockSpan struct {
	mock.Mock
}

func (s *mockSpan) Finish() {
	s.Called()
}
func (*mockSpan) FinishWithOptions(opts opentracing.FinishOptions)            {}
func (*mockSpan) Context() opentracing.SpanContext                            { return nil }
func (*mockSpan) SetOperationName(operationName string) opentracing.Span      { return nil }
func (*mockSpan) SetTag(key string, value interface{}) opentracing.Span       { return nil }
func (*mockSpan) LogFields(fields ...opentracingLog.Field)                    {}
func (*mockSpan) LogKV(alternatingKeyValues ...interface{})                   {}
func (*mockSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span { return nil }
func (*mockSpan) BaggageItem(restrictedKey string) string                     { return "" }
func (*mockSpan) Tracer() opentracing.Tracer                                  { return nil }
func (*mockSpan) LogEvent(event string)                                       {}
func (*mockSpan) LogEventWithPayload(event string, payload interface{})       {}
func (*mockSpan) Log(data opentracing.LogData)                                {}

// mockTracer implements tracing.Tracer interface
type mockTracer struct {
	mock.Mock
}

func (*mockTracer) IsEnabled() bool   { return false }
func (*mockTracer) SetEnabled(_ bool) {}
func (*mockTracer) StartSpanFromRequestHeader(
	context.Context,
	string,
	http.Header,
) (opentracing.Span, context.Context) {
	return nil, nil
}
func (t *mockTracer) StartSpanFromContext(
	ctx context.Context,
	name string,
) (opentracing.Span, context.Context) {
	t.Called(ctx, name)
	return nil, nil
}
func (*mockTracer) InitGlobalTracer(_ string, _ *config.JaegerConfig) (io.Closer, error) {
	return io.NopCloser(nil), nil
}

// Test that a startTimeKey has been associated to the context
func TestTimeInterceptorBeforeDispatch(t *testing.T) {
	// Set up the logging interceptor and turing context
	i, _, ctx := setUpLoggingInterceptorReqs(t, NewTimeLoggingInterceptor)

	// Call BeforeDispatch
	ctx = i.BeforeDispatch(ctx, nil)

	// Test that there is a start time key with time value
	val := ctx.Value(startTimeKey)
	assert.NotEqual(t, nil, val)
	_, ok := val.(time.Time)
	assert.Equal(t, true, ok)
}

// Test After Completion log
func TestTimeInterceptorAfterCompletion(t *testing.T) {
	// Set up the logging interceptor and turing context
	i, sink, ctx := setUpLoggingInterceptorReqs(t, NewTimeLoggingInterceptor)
	// Get turing request id from context
	turingReqID, err := turingctx.GetRequestID(ctx)
	tu.FailOnError(t, err)

	// Run AfterCompletion
	i.AfterCompletion(ctx, nil, nil)

	// Get the contents from the log sink
	logData := sink.Bytes()

	// Unmarshal the result
	var logObj timeInterceptorLog
	err = json.Unmarshal(logData, &logObj)
	tu.FailOnError(t, err)

	// Validate relevant fields
	assert.Equal(t, "debug", logObj.Level)
	assert.Equal(t, "Time Taken", logObj.Msg)
	assert.Equal(t, "test_ComponentID", logObj.Component)
	assert.Equal(t, turingReqID, logObj.TuringReqID)
	// Check that the duration is valid
	_, err = time.ParseDuration(logObj.Duration)
	assert.NoError(t, err)
}

func TestErrorInterceptorAfterCompletion(t *testing.T) {
	// Define tests
	tests := map[string]testSuiteErrLogInterceptor{
		"combiner": {
			kind:       fiber.CombinerKind,
			respStatus: http.StatusInternalServerError,
			errLog:     false,
		},
		"caller_success_response": {
			kind:       fiber.CallerKind,
			respStatus: http.StatusOK,
			errLog:     false,
		},
		"caller_failure_response": {
			kind:       fiber.CallerKind,
			respStatus: http.StatusInternalServerError,
			errLog:     true,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Set up the logging interceptor and turing context
			i, sink, ctx := setUpLoggingInterceptorReqs(t, NewErrorLoggingInterceptor)
			// Add Fiber Component kind
			ctx = context.WithValue(ctx, fiber.CtxComponentKindKey, data.kind)
			// Create a fiber response queue
			// Get turing request id from context
			turingReqID, err := turingctx.GetRequestID(ctx)
			tu.FailOnError(t, err)

			// Construct Fiber response and place in response queue
			queue := createTestFiberResponseQueue(data.respStatus)

			// Run AfterCompletion
			i.AfterCompletion(ctx, nil, queue)

			// Get the contents from the log sink
			logData := sink.Bytes()

			if data.errLog {
				// Unmarshal the result
				var logObj errorInterceptorLog
				err = json.Unmarshal(logData, &logObj)
				tu.FailOnError(t, err)

				// Validate relevant fields
				assert.Equal(t, "warn", logObj.Level)
				assert.Equal(t, "Route Error", logObj.Msg)
				assert.Equal(t, "test_ComponentID", logObj.Component)
				assert.Equal(t, turingReqID, logObj.TuringReqID)
				assert.Equal(t, data.respStatus, logObj.Status)
				assert.Equal(t,
					"{\n  \"code\": 500,\n  \"error\": \"Test Body\"\n}",
					logObj.Response)
				assert.NoError(t, err)
			} else {
				assert.Empty(t, logData)
			}
		})
	}
}

// setUpLoggingInterceptorReqs
func setUpLoggingInterceptorReqs(t *testing.T, f func(log.Logger) fiber.Interceptor) (
	fiber.Interceptor,
	*tu.MemorySink,
	context.Context,
) {
	// Create new logging interceptor
	// uses the zap logger with in-memory sink
	logger, sink, err := tu.NewLoggerWithMemorySink()
	tu.FailOnError(t, err)
	i := f(logger)

	// Make test context with the start time and turing request id
	ctx := context.WithValue(context.Background(), startTimeKey, time.Now())
	ctx = context.WithValue(ctx, fiber.CtxComponentIDKey, "test_ComponentID")
	ctx = turingctx.NewTuringContext(ctx)

	return i, sink, ctx
}

func TestMetricsInterceptorBeforeDispatch(t *testing.T) {
	i := NewMetricsInterceptor()
	ctx := i.BeforeDispatch(context.Background(), nil)
	// Verify that the start time has been set
	if _, ok := ctx.Value(startTimeKey).(time.Time); !ok {
		tu.FailOnError(t, fmt.Errorf("Start time has not been set in the context"))
	}
}

func TestMetricsInterceptorAfterCompletion(t *testing.T) {
	tests := map[string]testSuiteInstrInterceptor{
		"caller": {
			kind:   fiber.CallerKind,
			called: true,
		},
		"combiner": {
			kind:   fiber.CombinerKind,
			called: false,
		},
	}
	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Make instrumentation interceptor
			i := NewMetricsInterceptor()

			// Make test ctx
			starttime := time.Now()
			ctx := context.WithValue(context.Background(), startTimeKey, starttime)
			ctx = context.WithValue(ctx, fiber.CtxComponentIDKey, "test_ComponentID")
			ctx = context.WithValue(ctx, fiber.CtxComponentKindKey, data.kind)

			// Construct Fiber response and place in response queue
			queue := createTestFiberResponseQueue(http.StatusOK)

			// Patch global metrics collector and run AfterCompletion
			mc := &mockMetricsCollector{}
			mc.On("MeasureDurationMsSince",
				mock.Anything, mock.Anything, mock.Anything,
			).Return(nil)
			globMC := metrics.Glob()
			metrics.SetGlobMetricsCollector(mc)
			i.AfterCompletion(ctx, nil, queue)
			metrics.SetGlobMetricsCollector(globMC)

			// Verify invocation of MeasureDurationMsSince
			if data.called {
				mc.AssertCalled(t,
					"MeasureDurationMsSince",
					mock.Anything, mock.Anything, mock.Anything,
				)
			} else {
				mc.AssertNotCalled(t,
					"MeasureDurationMsSince",
					mock.Anything, mock.Anything, mock.Anything,
				)
			}
		})
	}
}

func TestTracingInterceptorBeforeDispatch(t *testing.T) {
	// Create test context
	compID := "test_ComponentID"
	ctx := context.WithValue(context.Background(),
		fiber.CtxComponentIDKey, compID)

	// Use mockTracer for testing
	mt := &mockTracer{}
	mt.On("StartSpanFromContext", ctx, compID)
	globalTracer := tracing.Glob()
	defer func() {
		tracing.SetGlob(globalTracer)
	}()
	tracing.SetGlob(mt)

	// Run Test
	i := NewTracingInterceptor()
	_ = i.BeforeDispatch(ctx, nil)

	// Validate that mockTracer.StartSpanFromContext was called
	mt.AssertCalled(t, "StartSpanFromContext", ctx, compID)
}

func TestTracingInterceptorAfterCompletion(t *testing.T) {
	// Create mock span
	mockSp := &mockSpan{}
	mockSp.On("Finish").Return(nil)

	// Patch opentracing.SpanFromContext to return the mock span
	monkey.Patch(opentracing.SpanFromContext,
		func(ctx context.Context) opentracing.Span {
			return mockSp
		})
	defer monkey.Unpatch(opentracing.SpanFromContext)

	// Run Test
	i := NewTracingInterceptor()
	_, ctx := opentracing.StartSpanFromContext(context.Background(), "test")
	i.AfterCompletion(ctx, nil, nil)

	// Validate that mockSpan.Finish() has been called
	mockSp.AssertCalled(t, "Finish")
}

func createTestFiberResponseQueue(respStatus int) fiber.ResponseQueue {
	testBody := []byte(`Test Body`)
	httpResp := http.Response{
		StatusCode: respStatus,
		Body:       io.NopCloser(bytes.NewBuffer(testBody)),
	}
	fiberResp := fiberHttp.NewHTTPResponse(&httpResp)
	queue := fiber.NewResponseQueueFromResponses(fiberResp)
	return queue
}
