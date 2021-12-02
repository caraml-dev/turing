package nop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTreatmentGetters(t *testing.T) {
	treatment := treatment{}
	assert.Equal(t, "", treatment.GetExperimentName())
	assert.Equal(t, "", treatment.GetName())
	assert.Nil(t, treatment.GetConfig())
}

func TestGetTreatmentForRequest(t *testing.T) {
	runner := ExperimentRunner{}
	_, err := runner.GetTreatmentForRequest(context.Background(), nil, nil, nil)
	assert.NoError(t, err)
}
