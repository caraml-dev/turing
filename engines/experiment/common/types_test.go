package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFieldSource(t *testing.T) {
	// header
	fieldSrc, err := GetFieldSource("header")
	assert.Equal(t, HeaderFieldSource, fieldSrc)
	assert.NoError(t, err)
	// payload
	fieldSrc, err = GetFieldSource("payload")
	assert.Equal(t, PayloadFieldSource, fieldSrc)
	assert.NoError(t, err)
	// unknown
	_, err = GetFieldSource("test")
	assert.Error(t, err)
}

func TestUnmarshalJSONFieldSource(t *testing.T) {
	var fieldSrc FieldSource
	// success
	err := json.Unmarshal([]byte(`"header"`), &fieldSrc)
	assert.Equal(t, HeaderFieldSource, fieldSrc)
	assert.NoError(t, err)
	// unknown string
	err = json.Unmarshal([]byte(`"test"`), &fieldSrc)
	assert.Error(t, err)
	// invalid data
	err = json.Unmarshal([]byte(`0`), &fieldSrc)
	assert.Error(t, err)
}
