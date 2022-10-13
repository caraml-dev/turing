package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStructType struct {
	Name   string
	Number int
}

var testStruct = testStructType{
	Name:   "string_val",
	Number: 10,
}

func TestGetStructFieldValues(t *testing.T) {
	expected := []interface{}{"string_val", 10}
	actual := getStructFieldValues(testStruct)
	assert.Equal(t, expected, actual)
}

func TestGetStructFieldsAsString(t *testing.T) {
	expected := []string{"string_val"}
	actual := GetStructFieldsAsString(testStruct)
	assert.Equal(t, expected, actual)
}
