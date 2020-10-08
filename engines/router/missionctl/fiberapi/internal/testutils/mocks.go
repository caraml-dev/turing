// Package testutils contains mocks that can be shared across tests for multiple packages
package testutils

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gojek/turing/engines/experiment/runner"
)

// MockExperimentRunner is a mock implementation for the Planner interface
type MockExperimentRunner struct {
	TestTreatment
	// If WantErr is true, GetTreatmentForRequest() will return a non-nil error
	WantErr bool
	// If WantTimeout is true, GetTreatmentForRequest() will wait for the duration of
	// Timeout and return a non-nil error
	WantTimeout bool
	// Timeout to wait for
	Timeout time.Duration
}

// TestTreatment is the wrapper for the experiment returned by the MockExperimentRunner,
// implementing the Plan interface

type TestTreatment struct {
	Name      string
	Treatment string
	Raw       json.RawMessage
}

// GetTreatmentForRequest returns the experiment treatment provided when MockExperimentRunner
// is initialized with TestTreatment.
// If MockExperimentRunner.WantErr is true, GetTreatmentForRequest will return error.
func (mp MockExperimentRunner) GetTreatmentForRequest(
	ctx context.Context,
	_ runner.Logger,
	_ http.Header,
	_ []byte,
) (runner.Treatment, error) {
	if mp.WantTimeout {
		time.Sleep(mp.Timeout)
		return nil, errors.New("timeout reached")
	} else if mp.WantErr {
		return nil, errors.New("failed to retrieve experiment treatment")
	}
	return mp.TestTreatment, nil

}

// GetExperimentName returns the name of the experiment
func (ex TestTreatment) GetExperimentName() string {
	return ex.Name
}

// GetName retrives the treatment (or control) name
func (ex TestTreatment) GetName() string {
	return ex.Treatment
}

// GetConfig returs the raw experiment config from the experiment engine
func (ex TestTreatment) GetConfig() json.RawMessage {
	return ex.Raw
}
