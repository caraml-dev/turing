package nop

import (
	"testing"

	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/stretchr/testify/assert"
)

func TestGetTreatmentForRequest(t *testing.T) {
	expRunner := ExperimentRunner{}
	treatment, err := expRunner.GetTreatmentForRequest(nil, nil, runner.GetTreatmentOptions{})
	assert.NoError(t, err)
	assert.Equal(t, nopTreatment, treatment)
}
