package nop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTreatmentForRequest(t *testing.T) {
	runner := ExperimentRunner{}
	treatment, err := runner.GetTreatmentForRequest(context.Background(), nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, nopTreatment, treatment)
}
