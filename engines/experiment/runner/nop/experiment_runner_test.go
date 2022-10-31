package nop

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/experiment/runner"
)

func TestGetTreatmentForRequest(t *testing.T) {
	expRunner := ExperimentRunner{}
	treatment, err := expRunner.GetTreatmentForRequest(nil, nil, runner.GetTreatmentOptions{})
	assert.NoError(t, err)
	assert.Equal(t, nopTreatment, treatment)
}
