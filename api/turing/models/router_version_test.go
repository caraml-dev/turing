package models

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouterVersionSetters(t *testing.T) {
	rv := RouterVersion{}
	rv.SetEnricherID(1)
	rv.SetEnsemblerID(2)
	// Validate
	assert.Equal(t, sql.NullInt32{Int32: int32(1), Valid: true}, rv.EnricherID)
	assert.Equal(t, sql.NullInt32{Int32: int32(2), Valid: true}, rv.EnsemblerID)
}
