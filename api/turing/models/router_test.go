package models

import (
	"database/sql"
	"encoding/json"
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

func TestRouterMarshalJSON(t *testing.T) {
	tests := map[string]struct {
		router   Router
		expected string
	}{
		"endpoint": {
			router: Router{
				Model: Model{
					ID: 1,
				},
				ProjectID: 2,
				Endpoint:  "test-endpoint",
			},
			expected: `{
				"id": 1,
				"created_at": "0001-01-01T00:00:00Z",
				"updated_at": "0001-01-01T00:00:00Z",
				"project_id": 2,
				"environment_name": "",
				"name": "",
				"status": "",
				"endpoint": "test-endpoint/v1/predict"
			}`,
		},
		"no endpoint": {
			router: Router{
				Model: Model{
					ID: 1,
				},
				ProjectID: 2,
			},
			expected: `{
				"id": 1,
				"created_at": "0001-01-01T00:00:00Z",
				"updated_at": "0001-01-01T00:00:00Z",
				"project_id": 2,
				"environment_name": "",
				"name": "",
				"status": ""
			}`,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			byteData, err := json.Marshal(&data.router)
			assert.NoError(t, err)
			assert.JSONEq(t, data.expected, string(byteData))
		})
	}
}
