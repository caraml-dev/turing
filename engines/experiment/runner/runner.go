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
	Register(metrics []instrumentation.Metric) error
}

// ExperimentRunner is the generic interface for generating experiment configs
// for a given request
type ExperimentRunner interface {
	GetTreatmentForRequest(
		header http.Header,
		payload []byte,
		options GetTreatmentOptions,
	) (*Treatment, error)
	RegisterMetricsCollector(
		collector metrics.Collector,
		metricsRegistrationHelper MetricsRegistrationHelper,
	) error
}
