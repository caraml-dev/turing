package api

import (
	"errors"
	"testing"

	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service/mocks"
	testifyAssert "github.com/stretchr/testify/assert"
)

func TestBaseControllerGetProjectFromRequestVars(t *testing.T) {
	// Create mock services
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", 1).Return(nil, errors.New("Test project error"))
	mlpSvc.On("GetProject", 2).Return(&mlp.Project{}, nil)

	// Define test cases
	tests := map[string]struct {
		vars     map[string]string
		expected *Response
	}{
		"failure | invalid project id": {
			vars:     map[string]string{},
			expected: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | project not found": {
			vars:     map[string]string{"project_id": "1"},
			expected: NotFound("project not found", "Test project error"),
		},
		"success": {
			vars:     map[string]string{"project_id": "2"},
			expected: nil,
		},
	}

	// Validate
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &baseController{
				&AppContext{
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
	routerSvc.On("FindByID", uint(1)).Return(nil, errors.New("Test router error"))
	routerSvc.On("FindByID", uint(2)).Return(&models.Router{}, nil)

	// Define test cases
	tests := map[string]struct {
		vars     map[string]string
		expected *Response
	}{
		"failure | invalid router id": {
			vars:     map[string]string{},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | router not found": {
			vars:     map[string]string{"router_id": "1"},
			expected: NotFound("router not found", "Test router error"),
		},
		"success": {
			vars:     map[string]string{"router_id": "2"},
			expected: nil,
		},
	}

	// Validate
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &baseController{
				&AppContext{
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
		On("FindByRouterIDAndVersion", uint(1), uint(1)).
		Return(nil, errors.New("Test router version error"))
	routerVersionSvc.
		On("FindByRouterIDAndVersion", uint(1), uint(2)).
		Return(&models.RouterVersion{}, nil)

	// Define test cases
	tests := map[string]struct {
		vars     map[string]string
		expected *Response
	}{
		"failure | invalid router id": {
			vars:     map[string]string{},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | invalid router version": {
			vars:     map[string]string{"router_id": "1"},
			expected: BadRequest("invalid router version value", "key version not found in vars"),
		},
		"failure | router version not found": {
			vars:     map[string]string{"router_id": "1", "version": "1"},
			expected: NotFound("router version not found", "Test router version error"),
		},
		"success": {
			vars:     map[string]string{"router_id": "1", "version": "2"},
			expected: nil,
		},
	}

	// Validate
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &baseController{
				&AppContext{
					RouterVersionsService: routerVersionSvc,
				},
			}
			// Run test method and validate
			_, response := ctrl.getRouterVersionFromRequestVars(data.vars)
			testifyAssert.Equal(t, data.expected, response)
		})
	}
}
