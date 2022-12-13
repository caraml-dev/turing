package runner

import (
	"encoding/json"
	"net/http"

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

// ExperimentRunner is the generic interface for generating experiment configs
// for a given request
type ExperimentRunner interface {
	GetTreatmentForRequest(
		header http.Header,
		payload []byte,
		options GetTreatmentOptions,
	) (*Treatment, error)
	RegisterCollector(
		collector metrics.Collector,
	) error
}
