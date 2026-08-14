package tracing

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

// OtelTracer implements the Tracer interface using the OpenTelemetry SDK,
// exporting spans via OTLP over HTTP.
type OtelTracer struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

// InitGlobalTracer creates an OTel TracerProvider exporting to jCfg.CollectorEndpoint
// via OTLP HTTP, registers it (and a composite W3C trace-context/baggage/B3 propagator) as
// the OTel globals, and returns its Shutdown function.
func (t *OtelTracer) InitGlobalTracer(name string, jCfg *config.JaegerConfig) (ShutdownFunc, error) {
	ctx := context.Background()

	// Validate the endpoint has a host before handing it to otlptracehttp: WithEndpointURL
	// silently falls back to an empty host on a parse error or an empty/host-less string,
	// which would construct an exporter that can never successfully connect.
	endpointURL, err := url.Parse(jCfg.CollectorEndpoint)
	if err != nil {
		return nil, err
	}
	if endpointURL.Host == "" {
		return nil, fmt.Errorf("invalid CollectorEndpoint %q: missing host", jCfg.CollectorEndpoint)
	}

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(jCfg.CollectorEndpoint))
	if err != nil {
		return nil, err
	}

	ratio := jCfg.SamplingRatio
	if ratio <= 0 {
		ratio = 1
	}

	res := resource.NewSchemaless(attribute.String("service.name", name))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
		sdktrace.WithResource(res),
	)

	// Register both W3C TraceContext and B3 extractors so the router can join traces from
	// callers using either propagation format. B3 is Istio's default, so this matters for
	// routers running behind Istio/Knative; W3C TraceContext remains the injection format.
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}, b3.New(),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagator)

	t.tracer = tp.Tracer(name)
	t.propagator = propagator

	return tp.Shutdown, nil
}

// IsEnabled satisfies the Tracer interface, always returning true
func (*OtelTracer) IsEnabled() bool {
	return true
}

// StartSpanFromRequestHeader extracts a remote span context from the request header
// (via the W3C traceparent/tracestate headers) and starts a child span from it.
func (t *OtelTracer) StartSpanFromRequestHeader(
	ctx context.Context,
	opName string,
	header http.Header,
) (trace.Span, context.Context) {
	ctx = t.propagator.Extract(ctx, propagation.HeaderCarrier(header))
	ctx, span := t.tracer.Start(ctx, opName)
	return span, ctx
}

// StartSpanFromContext starts a new / child span associated with the given context.
func (t *OtelTracer) StartSpanFromContext(
	ctx context.Context,
	opName string,
) (trace.Span, context.Context) {
	ctx, span := t.tracer.Start(ctx, opName)
	return span, ctx
}

// newOtelTracer is a creator for the OtelTracer
func newOtelTracer() Tracer {
	return &OtelTracer{}
}
