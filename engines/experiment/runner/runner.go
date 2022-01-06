package runner

import (
	"net/http"
)

type GetTreatmentOptions struct {
	TuringRequestID string
}

// ExperimentRunner is the generic interface for generating experiment configs
// for a given request
type ExperimentRunner interface {
	GetTreatmentForRequest(
		header http.Header,
		payload []byte,
		options GetTreatmentOptions,
	) (*Treatment, error)
}
