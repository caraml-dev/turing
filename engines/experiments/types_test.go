package experiments_test

import (
	"encoding/json"
	"testing"

	experiments "github.com/gojek/turing/engines/experiment/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetFieldSource(t *testing.T) {
	// header
	fieldSrc, err := experiments.GetFieldSource("header")
	assert.Equal(t, experiments.HeaderFieldSource, fieldSrc)
	assert.NoError(t, err)
	// payload
	fieldSrc, err = experiments.GetFieldSource("payload")
	assert.Equal(t, experiments.PayloadFieldSource, fieldSrc)
	assert.NoError(t, err)
	// unknown
	_, err = experiments.GetFieldSource("test")
	assert.Error(t, err)
}

func TestUnmarshalJSONFieldSource(t *testing.T) {
	var fieldSrc experiments.FieldSource
	// success
	err := json.Unmarshal([]byte(`"header"`), &fieldSrc)
	assert.Equal(t, experiments.HeaderFieldSource, fieldSrc)
	assert.NoError(t, err)
	// unknown string
	err = json.Unmarshal([]byte(`"test"`), &fieldSrc)
	assert.Error(t, err)
	// invalid data
	err = json.Unmarshal([]byte(`0`), &fieldSrc)
	assert.Error(t, err)
}
