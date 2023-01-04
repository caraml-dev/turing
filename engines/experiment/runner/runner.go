package runner

import (
	"encoding/json"
	"net/http"

	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation"
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
)

type GetTreatmentOptions struct {
	TuringRequestID string
}

type Treatment struct {
	ExperimentName string
	Name           string
	Config         json.RawMessage
}

// MetricsRegistrationHelper is the generic interface for the Turing router to
// register additional metrics needed by the experiment engine
type MetricsRegistrationHelper interface {
	// Register is a method that should be called within ExperimentRunner.RegisterMetricsCollector to register
	// additional metrics that the experiment engine requires on the metrics collector of the Turing Router
	Register(metrics []instrumentation.Metric) error
}

// ExperimentRunner is the generic interface for generating experiment configs
// for a given request
type ExperimentRunner interface {
	// GetTreatmentForRequest is a method that is called by the Turing Router for each request it receives to retrieve
	// an appropriate treatment for it
	GetTreatmentForRequest(
		header http.Header,
		payload []byte,
		options GetTreatmentOptions,
	) (*Treatment, error)
	// RegisterMetricsCollector is a method that should only be called on startup of the experiment runner to register
	// the metrics collector of the Turing Router (to be called just after the experiment runner has been initialised)
	RegisterMetricsCollector(
		collector metrics.Collector,
		metricsRegistrationHelper MetricsRegistrationHelper,
	) error
}
