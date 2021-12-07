package runner

import (
	"context"
	"net/http"
)

// ExperimentRunner is the generic interface for generating experiment configs for a
// given request
type ExperimentRunner interface {
	GetTreatmentForRequest(
		context.Context,
		Logger,
		http.Header,
		[]byte,
	) (*Treatment, error)
}
