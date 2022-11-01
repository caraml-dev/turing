package api

import (
	"errors"
	"testing"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service/mocks"
	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
)

const (
	monitoringURL = "http://www.example.com"
)

func TestListRouterVersions(t *testing.T) {
	// Create mock services
	routerSvc := &mocks.RoutersService{}
	routerSvc.
		On("FindByID", models.ID(1)).
		Return(&models.Router{Model: models.Model{ID: 1}}, nil)
	routerSvc.
		On("FindByID", models.ID(2)).
		Return(&models.Router{Model: models.Model{ID: 2}}, nil)
	testVersions := []*models.RouterVersion{{Version: 1, MonitoringURL: monitoringURL}}
	routerVersionSvc := &mocks.RouterVersionsService{}
	routerVersionSvc.
		On("ListRouterVersions", models.ID(1)).
		Return(nil, errors.New("test router versions error"))
	routerVersionSvc.
		On("ListRouterVersions", models.ID(2)).
		Return(testVersions, nil)

	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", models.ID(1)).Return(&mlp.Project{Id: 3, Name: "mlp-project"}, nil)

	// Define tests
	tests := map[string]struct {
		vars     RequestVars
		expected *Response
	}{
		"failure | bad request (missing router_id)": {
			vars:     RequestVars{"project_id": {"1"}},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | list router versions": {
			vars:     RequestVars{"router_id": {"1"}, "project_id": {"1"}},
			expected: InternalServerError("unable to retrieve router versions", "test router versions error"),
		},
		"success": {
			vars: RequestVars{"router_id": {"2"}, "project_id": {"1"}},
			expected: &Response{
				code: 200,
				data: testVersions,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RouterVersionsController{
				RouterDeploymentController{
					BaseController{
						AppContext: &AppContext{
							RoutersService:        routerSvc,
							RouterVersionsService: routerVersionSvc,
							MLPService:            mlpSvc,
						},
					},
				},
			}
			// Run test method and validate
			response := ctrl.ListRouterVersions(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestCreateRouterVersion(t *testing.T) {
	// Create mock services
	// MLP service
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", models.ID(1)).Return(nil, errors.New("test project error"))
	mlpSvc.On("GetProject", models.ID(2)).Return(&mlp.Project{Id: 2}, nil)
	mlpSvc.On("GetEnvironment", "dev-invalid").Return(nil, errors.New("test env error"))
	mlpSvc.On("GetEnvironment", "dev").Return(&merlin.Environment{}, nil)

	// Router Service
	router2 := &models.Router{
		Name:            "router2",
		ProjectID:       2,
		EnvironmentName: "dev",
		Model: models.Model{
			ID: 2,
		},
	}
	routerSvc := &mocks.RoutersService{}
	routerSvc.On("FindByID", models.ID(1)).
		Return(nil, errors.New("test router error"))
	routerSvc.On("FindByID", models.ID(2)).Return(router2, nil)

	// Router Version Service
	routerVersion := &models.RouterVersion{
		Router:   router2,
		RouterID: 2,
		ExperimentEngine: &models.ExperimentEngine{
			Type: models.ExperimentEngineTypeNop,
		},
		Protocol: routerConfig.HTTP,
		LogConfig: &models.LogConfig{
			ResultLoggerType: models.NopLogger,
		},
		Status: models.RouterVersionStatusUndeployed,
		AutoscalingPolicy: &models.AutoscalingPolicy{
			Metric: models.AutoscalingMetricCPU,
			Target: "80",
		},
	}
	routerVersionSvc := &mocks.RouterVersionsService{}
	routerVersionSvc.On("Save", routerVersion).Return(routerVersion, nil)

	// Define tests
	tests := map[string]struct {
		vars     RequestVars
		expected *Response
		body     *request.RouterConfig
	}{
		"failure | bad request (missing project_id)": {
			vars:     RequestVars{},
			expected: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | not found (project not found)": {
			vars:     RequestVars{"project_id": {"1"}},
			expected: NotFound("project not found", "test project error"),
		},
		"failure | bad request (missing router_id)": {
			vars:     RequestVars{"project_id": {"2"}},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | not found (router_id not found)": {
			vars:     RequestVars{"project_id": {"2"}, "router_id": {"1"}},
			expected: NotFound("router not found", "test router error"),
		},
		"failure | build router version": {
			body:     nil,
			vars:     RequestVars{"project_id": {"2"}, "router_id": {"2"}},
			expected: InternalServerError("unable to create router version", "router config is empty"),
		},
		"success": {
			body: &request.RouterConfig{
				ExperimentEngine: &request.ExperimentEngineConfig{
					Type: "nop",
				},
				LogConfig: &request.LogConfig{
					ResultLoggerType: models.NopLogger,
				},
				AutoscalingPolicy: &models.AutoscalingPolicy{
					Metric: models.AutoscalingMetricCPU,
					Target: "80",
				},
			},
			vars: RequestVars{"router_id": {"2"}, "project_id": {"2"}},
			expected: &Response{
				code: 200,
				data: routerVersion,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RouterVersionsController{
				RouterDeploymentController{
					BaseController{
						AppContext: &AppContext{
							MLPService:            mlpSvc,
							RoutersService:        routerSvc,
							RouterVersionsService: routerVersionSvc,
							RouterDefaults:        &config.RouterDefaults{},
						},
					},
				},
			}
			// Run test method and validate
			response := ctrl.CreateRouterVersion(nil, data.vars, data.body)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestGetRouterVersion(t *testing.T) {
	// Create mock services
	routerSvc := &mocks.RoutersService{}
	routerSvc.
		On("FindByID", models.ID(1)).
		Return(&models.Router{Model: models.Model{ID: 1}}, nil)
	testVersion := &models.RouterVersion{Version: 1, MonitoringURL: monitoringURL}
	routerVersionSvc := &mocks.RouterVersionsService{}
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(1), uint(1)).
		Return(nil, errors.New("test router version error"))
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(1), uint(2)).
		Return(testVersion, nil)

	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", models.ID(1)).Return(&mlp.Project{Id: 3, Name: "mlp-project"}, nil)

	// Define tests
	tests := map[string]struct {
		vars     RequestVars
		expected *Response
	}{
		"failure | bad request (missing router_id)": {
			vars:     RequestVars{"project_id": {"1"}},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | bad request (missing version)": {
			vars:     RequestVars{"router_id": {"1"}, "project_id": {"1"}},
			expected: BadRequest("invalid router version value", "key version not found in vars"),
		},
		"failure | get router version": {
			vars:     RequestVars{"router_id": {"1"}, "version": {"1"}, "project_id": {"1"}},
			expected: NotFound("router version not found", "test router version error"),
		},
		"success": {
			vars: RequestVars{"router_id": {"1"}, "version": {"2"}, "project_id": {"1"}},
			expected: &Response{
				code: 200,
				data: testVersion,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RouterVersionsController{
				RouterDeploymentController{
					BaseController{
						AppContext: &AppContext{
							RoutersService:        routerSvc,
							RouterVersionsService: routerVersionSvc,
							MLPService:            mlpSvc,
						},
					},
				},
			}
			// Run test method and validate
			response := ctrl.GetRouterVersion(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestDeleteRouterVersion(t *testing.T) {
	// Create mock services
	// Router versions service
	routerVersion2 := &models.RouterVersion{
		Version: 2,
		Status:  models.RouterVersionStatusPending,
	}
	routerVersion3 := &models.RouterVersion{
		Model: models.Model{
			ID: 10,
		},
		Version: 3,
		Status:  models.RouterVersionStatusUndeployed,
	}
	routerVersion4 := &models.RouterVersion{
		Model: models.Model{
			ID: 11,
		},
		Version: 4,
		Status:  models.RouterVersionStatusFailed,
	}
	routerVersion5 := &models.RouterVersion{
		Model: models.Model{
			ID: 12,
		},
		Version: 5,
		Status:  models.RouterVersionStatusFailed,
	}
	routerVersionSvc := &mocks.RouterVersionsService{}
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(2), uint(1)).
		Return(nil, errors.New("test router version error"))
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(2), uint(2)).
		Return(routerVersion2, nil)
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(2), uint(3)).
		Return(routerVersion3, nil)
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(2), uint(4)).
		Return(routerVersion4, nil)
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(2), uint(5)).
		Return(routerVersion5, nil)
	routerVersionSvc.
		On("Delete", routerVersion4).
		Return(errors.New("test router version delete error"))
	routerVersionSvc.
		On("Delete", routerVersion5).
		Return(nil)
	// Router service
	router2 := &models.Router{
		Model: models.Model{
			ID: 2,
		},
		CurrRouterVersion: &models.RouterVersion{
			Model: models.Model{
				ID: 10,
			},
		},
	}
	routerSvc := &mocks.RoutersService{}
	routerSvc.
		On("FindByID", models.ID(1)).
		Return(nil, errors.New("test router error"))
	routerSvc.
		On("FindByID", models.ID(2)).
		Return(router2, nil)

	// Define tests
	tests := map[string]struct {
		vars     RequestVars
		expected *Response
	}{
		"failure | bad request (missing router_id)": {
			vars:     RequestVars{},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | get router": {
			vars:     RequestVars{"router_id": {"1"}, "version": {"1"}},
			expected: NotFound("router not found", "test router error"),
		},
		"failure | bad request (missing version)": {
			vars:     RequestVars{"router_id": {"2"}},
			expected: BadRequest("invalid router version value", "key version not found in vars"),
		},
		"failure | get router version": {
			vars:     RequestVars{"router_id": {"2"}, "version": {"1"}},
			expected: NotFound("router version not found", "test router version error"),
		},
		"failure | router version deploying": {
			vars: RequestVars{"router_id": {"2"}, "version": {"2"}},
			expected: BadRequest(
				"invalid delete request",
				"unable to delete router version that is currently deploying",
			),
		},
		"failure | router version current": {
			vars: RequestVars{"router_id": {"2"}, "version": {"3"}},
			expected: &Response{
				code: 400,
				data: struct {
					Description string `json:"description"`
					Message     string `json:"error"`
				}{
					Description: "invalid delete request",
					Message:     "cannot delete current router configuration",
				},
			},
		},
		"failure | delete router version": {
			vars: RequestVars{"router_id": {"2"}, "version": {"4"}},
			expected: &Response{
				code: 500,
				data: struct {
					Description string `json:"description"`
					Message     string `json:"error"`
				}{
					Description: "unable to delete router version",
					Message:     "test router version delete error",
				},
			},
		},
		"success": {
			vars: RequestVars{"router_id": {"2"}, "version": {"5"}},
			expected: &Response{
				code: 200,
				data: map[string]int{"router_id": 2, "version": 5},
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RouterVersionsController{
				RouterDeploymentController{
					BaseController{
						AppContext: &AppContext{
							RoutersService:        routerSvc,
							RouterVersionsService: routerVersionSvc,
						},
					},
				},
			}
			// Run test method and validate
			response := ctrl.DeleteRouterVersion(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestDeployRouterVersion(t *testing.T) {
	// Create mock services
	// MLP service
	mlpSvc := &mocks.MLPService{}
	mlpSvc.
		On("GetProject", models.ID(1)).
		Return(nil, errors.New("test project error"))
	mlpSvc.
		On("GetProject", models.ID(2)).
		Return(&mlp.Project{}, nil)
	mlpSvc.
		On("GetEnvironment", "dev-invalid").
		Return(nil, errors.New("test env error"))
	mlpSvc.
		On("GetEnvironment", "dev").
		Return(&merlin.Environment{}, nil)
	// Router Service
	router2 := &models.Router{
		Name:            "router2",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusPending,
		Model: models.Model{
			ID: 2,
		},
	}
	router3 := &models.Router{
		Name:            "router3",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusUndeployed,
		Model: models.Model{
			ID: 3,
		},
	}
	router4 := &models.Router{
		Name:            "router4",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusUndeployed,
		Model: models.Model{
			ID: 4,
		},
	}
	routerSvc := &mocks.RoutersService{}
	routerSvc.
		On("FindByID", models.ID(1)).
		Return(nil, errors.New("test router error"))
	routerSvc.
		On("FindByID", models.ID(2)).
		Return(router2, nil)
	routerSvc.
		On("FindByID", models.ID(3)).
		Return(router3, nil)
	routerSvc.
		On("FindByID", models.ID(4)).
		Return(router4, nil)
	// For the deployment method
	routerSvc.
		On("Save", mock.Anything).
		Return(nil, errors.New("test Router Deployment Failure"))
	// Router Version Service
	routerVersion := &models.RouterVersion{
		RouterID: 3,
		Router:   router3,
		Version:  2,
	}
	routerVersion2 := &models.RouterVersion{
		RouterID: 3,
		Router:   router2,
		Version:  3,
		Status:   models.RouterVersionStatusDeployed,
	}
	routerVersion3 := &models.RouterVersion{
		RouterID: 3,
		Router:   router3,
		Version:  3,
		Status:   models.RouterVersionStatusUndeployed,
	}
	routerVersion4 := &models.RouterVersion{
		RouterID: 4,
		Router:   router3,
		Version:  4,
		Status:   models.RouterVersionStatusUndeployed,
	}
	routerVersionSvc := &mocks.RouterVersionsService{}
	routerVersionSvc.
		On("FindByID", models.ID(1)).
		Return(nil, errors.New("test router version error"))
	routerVersionSvc.
		On("FindByID", models.ID(2)).
		Return(routerVersion, nil)
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(2), uint(2)).
		Return(nil, nil)
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(3), uint(1)).
		Return(nil, errors.New("test router version error"))
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(3), uint(2)).
		Return(routerVersion2, nil)
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(3), uint(3)).
		Return(routerVersion3, nil)
	routerVersionSvc.
		On("FindByRouterIDAndVersion", models.ID(4), uint(4)).
		Return(routerVersion4, nil)

	// Define tests
	tests := map[string]struct {
		vars     RequestVars
		expected *Response
	}{
		"failure | bad request (missing project_id)": {
			vars:     RequestVars{},
			expected: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | project not found": {
			vars:     RequestVars{"project_id": {"1"}, "router_id": {"1"}, "version": {"1"}},
			expected: NotFound("project not found", "test project error"),
		},
		"failure | bad request (missing router_id)": {
			vars:     RequestVars{"project_id": {"2"}},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | router not found": {
			vars:     RequestVars{"project_id": {"2"}, "router_id": {"1"}},
			expected: NotFound("router not found", "test router error"),
		},
		"failure | bad request (missing version)": {
			vars:     RequestVars{"project_id": {"2"}, "router_id": {"2"}},
			expected: BadRequest("invalid router version value", "key version not found in vars"),
		},
		"failure | router version not found": {
			vars:     RequestVars{"project_id": {"2"}, "router_id": {"3"}, "version": {"1"}},
			expected: NotFound("router version not found", "test router version error"),
		},
		"failure | router status pending": {
			vars: RequestVars{"project_id": {"2"}, "router_id": {"2"}, "version": {"2"}},
			expected: BadRequest(
				"invalid deploy request",
				"router is currently deploying, cannot do another deployment",
			),
		},
		"failure | version already deployed": {
			vars:     RequestVars{"project_id": {"2"}, "router_id": {"3"}, "version": {"2"}},
			expected: BadRequest("invalid deploy request", "router version is already deployed"),
		},
		"success": {
			vars: RequestVars{"project_id": {"2"}, "router_id": {"4"}, "version": {"4"}},
			expected: &Response{
				code: 202,
				data: map[string]int{
					"router_id": 4,
					"version":   4,
				},
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RouterVersionsController{
				RouterDeploymentController{
					BaseController{
						AppContext: &AppContext{
							MLPService:            mlpSvc,
							RoutersService:        routerSvc,
							RouterVersionsService: routerVersionSvc,
							RouterDefaults:        &config.RouterDefaults{},
						},
					},
				},
			}
			// Run test method and validate
			response := ctrl.DeployRouterVersion(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}
