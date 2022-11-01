package tracing

import (
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
)

// JaegerTracer implements the Tracer interface using the jaeger client library
type JaegerTracer struct {
	*baseTracer
}

// InitGlobalTracer creates a global tracer using the Jaeger client
func (t *JaegerTracer) InitGlobalTracer(name string, cfg *config.JaegerConfig) (io.Closer, error) {
	t.SetStartNewSpans(cfg.StartNewSpans)

	// Create a zipkin propagator as the HTTP extractor
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	// Initialize tracer with the default logger
	return buildConfig(cfg).InitGlobalTracer(name,
		jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator))
}

// IsEnabled satisfies the Tracer interface, always returning true
func (*JaegerTracer) IsEnabled() bool {
	return true
}

// buildConfig converts the input JaegerConfig into the format that can be interpreted
// by the Jaeger client library, applying const sampling
func buildConfig(jCfg *config.JaegerConfig) jaegercfg.Configuration {
	return jaegercfg.Configuration{
		Disabled: !jCfg.Enabled,
		Reporter: &jaegercfg.ReporterConfig{
			CollectorEndpoint: jCfg.CollectorEndpoint,
			LocalAgentHostPort: fmt.Sprintf("%s:%d",
				jCfg.ReporterAgentHost, jCfg.ReporterAgentPort),
			LogSpans: true,
		},
		// Sample all requests by default. Additional setting `startNewSpans`, local to
		// the package, may filter the actual requests being sampled.
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}
}

// newJaegerTracer is a creator for the Jaeger Tracer
func newJaegerTracer() Tracer {
	return &JaegerTracer{
		&baseTracer{
			&tracingConfig{
				startNewSpans: false,
			},
		},
	}
}
