// Package testutils contains mocks that can be shared across tests for multiple packages
package testutils

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gojek/turing/engines/experiment/runner"
)

// MockExperimentRunner is a mock implementation for the Planner interface
type MockExperimentRunner struct {
	*runner.Treatment
	// If WantErr is true, GetTreatmentForRequest() will return a non-nil error
	WantErr bool
	// If WantTimeout is true, GetTreatmentForRequest() will wait for the duration of
	// Timeout and return a non-nil error
	WantTimeout bool
	// Timeout to wait for
	Timeout time.Duration
}

// GetTreatmentForRequest returns the experiment treatment provided when MockExperimentRunner
// is initialized with TestTreatment.
// If MockExperimentRunner.WantErr is true, GetTreatmentForRequest will return error.
func (mp MockExperimentRunner) GetTreatmentForRequest(
	ctx context.Context,
	_ runner.Logger,
	_ http.Header,
	_ []byte,
) (*runner.Treatment, error) {
	if mp.WantTimeout {
		time.Sleep(mp.Timeout)
		return nil, errors.New("timeout reached")
	} else if mp.WantErr {
		return nil, errors.New("failed to retrieve experiment treatment")
	}
	return mp.Treatment, nil

}
