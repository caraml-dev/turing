package models

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetCurrRouterVersion(t *testing.T) {
	router := Router{}
	routerVersion := RouterVersion{
		Model: Model{
			ID: 1,
		},
	}
	router.SetCurrRouterVersion(&routerVersion)
	// Validate
	assert.Equal(t, &routerVersion, router.CurrRouterVersion)
	assert.Equal(t, sql.NullInt32{Int32: int32(1), Valid: true}, router.CurrRouterVersionID)
}

func TestClearCurrRouterVersion(t *testing.T) {
	routerVersion := RouterVersion{
		Model: Model{
			ID: 1,
		},
	}
	router := Router{
		CurrRouterVersion:   &routerVersion,
		CurrRouterVersionID: sql.NullInt32{Int32: int32(1), Valid: true},
	}
	router.ClearCurrRouterVersion()
	// Validate
	assert.True(t, router.CurrRouterVersion == nil)
	assert.Equal(t, sql.NullInt32{Int32: int32(0), Valid: false}, router.CurrRouterVersionID)
}
