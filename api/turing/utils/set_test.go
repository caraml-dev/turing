package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueStrings(t *testing.T) {
	uniqueStrings := []string{"string-1", "string-2"}
	nonUniqueStrings := []string{"string-1", "string-1"}

	isUnique := IsUniqueStrings(uniqueStrings)
	isNonUnique := IsUniqueStrings(nonUniqueStrings)
	assert.EqualError(t, true, isUnique)
	assert.EqualError(t, false, isNonUnique)
}
