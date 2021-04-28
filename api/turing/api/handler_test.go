package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockHandler(_ *http.Request, _ RequestVars, _ interface{}) *Response {
	return nil
}

func TestRoute_Name(t *testing.T) {
	ctrl := RoutersController{nil}

	tests := map[string]struct {
		route    Route
		expected string
	}{
		"success | custom handler": {
			route: Route{
				method:  http.MethodGet,
				path:    "/projects/{project_id}/entities",
				handler: mockHandler,
			},
			expected: "github.com/gojek/turing/api/turing/api.mockHandler",
		},
		"success | override handler name": {
			route: Route{
				method:  http.MethodGet,
				path:    "/projects/{project_id}/entities",
				handler: mockHandler,
				name:    "AnotherName",
			},
			expected: "AnotherName",
		},
		"success | controller's method": {
			route: Route{
				method:  http.MethodGet,
				path:    "/projects/{project_id}/entities",
				handler: ctrl.ListRouters,
			},
			expected: "github.com/gojek/turing/api/turing/api.RoutersController.ListRouters-fm",
		},
		"success | nil handler": {
			route: Route{
				method: http.MethodGet,
				path:   "/projects/{project_id}/entities",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.route.Name())
		})
	}
}
