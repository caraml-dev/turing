package runner

import (
	"context"
	"encoding/json"
	"net/http"
)

// Treatment is the interface implemented by types that are representative of an
// experiment configuration
type Treatment interface {
	GetExperimentName() string
	GetName() string
	GetConfig() json.RawMessage
}

// ExperimentRunner is the generic interface for generating experiment configs for a
// given request
type ExperimentRunner interface {
	GetTreatmentForRequest(context.Context, Logger, http.Header, []byte,
	) (Treatment, error)
}
