package api

import (
	"errors"
	"testing"

	mlp "github.com/gojek/mlp/api/client"
	testifyAssert "github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service/mocks"
)

func TestBaseControllerGetProjectFromRequestVars(t *testing.T) {
	// Create mock services
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", models.ID(1)).
		Return(nil, errors.New("test project error"))
	mlpSvc.On("GetProject", models.ID(2)).Return(&mlp.Project{}, nil)

	// Define test cases
	tests := map[string]struct {
		vars     RequestVars
		expected *Response
	}{
		"failure | invalid project id": {
			vars:     RequestVars{},
			expected: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | project not found": {
			vars:     RequestVars{"project_id": {"1"}},
			expected: NotFound("project not found", "test project error"),
		},
		"success": {
			vars:     RequestVars{"project_id": {"2"}},
			expected: nil,
		},
	}

	// Validate
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := BaseController{
				AppContext: &AppContext{
					MLPService: mlpSvc,
				},
			}
			// Run test method and validate
			_, response := ctrl.getProjectFromRequestVars(data.vars)
			testifyAssert.Equal(t, data.expected, response)
		})
	}
}

func TestBaseControllerGetRouterFromRequestVars(t *testing.T) {
	// Create mock services
	routerSvc := &mocks.RoutersService{}
	routerSvc.On("FindByID", models.ID(1)).
		Return(nil, errors.New("test router error"))
	routerSvc.On("FindByID", models.ID(2)).Return(&models.Router{}, nil)

	// Define test cases
	tests := map[string]struct {
		vars     RequestVars
		expected *Response
	}{
		"failure | invalid router id": {
			vars:     RequestVars{},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | router not found": {
			vars:     RequestVars{"router_id": {"1"}},
			expected: NotFound("router not found", "test router error"),
		},
		"success": {
			vars:     RequestVars{"router_id": {"2"}},
			expected: nil,
		},
	}

	// Validate
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := BaseController{
				AppContext: &AppContext{
					RoutersService: routerSvc,
				},
			}
			// Run test method and validate
			_, response := ctrl.getRouterFromRequestVars(data.vars)
			testifyAssert.Equal(t, data.expected, response)
		})
	}
}

func TestBaseControllerGetRouterVersionFromRequestVars(t *testing.T) {
	// Create mock services
	routerVersionSvc := &mocks.RouterVersionsService{}
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(1), uint(1)).
		Return(nil, errors.New("test router version error"))
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(1), uint(2)).
		Return(&models.RouterVersion{}, nil)

	// Define test cases
	tests := map[string]struct {
		vars     RequestVars
		expected *Response
	}{
		"failure | invalid router id": {
			vars:     RequestVars{},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | invalid router version": {
			vars:     RequestVars{"router_id": {"1"}},
			expected: BadRequest("invalid router version value", "key version not found in vars"),
		},
		"failure | router version not found": {
			vars:     RequestVars{"router_id": {"1"}, "version": {"1"}},
			expected: NotFound("router version not found", "test router version error"),
		},
		"success": {
			vars:     RequestVars{"router_id": {"1"}, "version": {"2"}},
			expected: nil,
		},
	}

	// Validate
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := BaseController{
				AppContext: &AppContext{
					RouterVersionsService: routerVersionSvc,
				},
			}
			// Run test method and validate
			_, response := ctrl.getRouterVersionFromRequestVars(data.vars)
			testifyAssert.Equal(t, data.expected, response)
		})
	}
}
