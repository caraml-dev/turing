package api

import (
	"net/http"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/models"
	mlp "github.com/gojek/mlp/api/client"
)

type RouterVersionsController struct {
	BaseController
}

// ListRouterVersions lists the router versions of the provided router.
func (c RouterVersionsController) ListRouterVersions(
	r *http.Request,
	vars RequestVars, _ interface{},
) *Response {
	// Parse request vars
	var errResp *Response
	var router *models.Router
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// List router versions
	routerVersions, err := c.Services.RouterVersionsService.ListRouterVersions(router.ID)
	if err != nil {
		return InternalServerError("unable to retrieve router versions", err.Error())
	}

	return Ok(routerVersions)
}

// CreateRouterVersion creates a router version from the provided configuration. If no router exists
// within the provided project with the provided id, this method will throw an error.
// If the update is valid, a new RouterVersion will be created but NOT deployed.
func (c RouterVersionsController) CreateRouterVersion(
	r *http.Request, vars RequestVars, body interface{}) *Response {
	// Parse request vars
	var errResp *Response
	var router *models.Router
	if _, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	request := body.(*request.RouterConfig)

	// Create new version
	var routerVersion *models.RouterVersion
	if request == nil {
		return InternalServerError("unable to create router version", "router config is empty")
	}

	routerVersion, err := request.BuildRouterVersion(
		router, c.RouterDefaults,
		c.Services.CryptoService,
		c.Services.ExperimentsService,
		c.Services.EnsemblersService,
	)

	if err == nil {
		// Save router version, re-assign the value of err
		routerVersion.Status = models.RouterVersionStatusUndeployed
		routerVersion, err = c.Services.RouterVersionsService.CreateRouterVersion(routerVersion)
	}

	if err != nil {
		return InternalServerError("unable to create router version", err.Error())
	}

	return Ok(routerVersion)
}

// GetRouterVersion gets the router version for the provided router id and version number.
func (c RouterVersionsController) GetRouterVersion(
	r *http.Request,
	vars RequestVars, _ interface{},
) *Response {
	// Parse request vars
	var errResp *Response
	var routerVersion *models.RouterVersion
	if routerVersion, errResp = c.getRouterVersionFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// Return router version
	return Ok(routerVersion)
}

// DeleteRouterVersion deletes the config for the given version number.
func (c RouterVersionsController) DeleteRouterVersion(
	r *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	// Parse request vars
	var errResp *Response
	var router *models.Router
	var routerVersion *models.RouterVersion
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if routerVersion, errResp = c.getRouterVersionFromRequestVars(vars); errResp != nil {
		return errResp
	}

	err := c.Services.RouterVersionsService.Delete(routerVersion)
	if err != nil {
		return InternalServerError("unable to delete router version", err.Error())
	}
	return Ok(map[string]int{"router_id": int(router.ID), "version": int(routerVersion.Version)})
}

// DeployRouterVersion deploys the given router version into the associated kubernetes cluster
func (c RouterVersionsController) DeployRouterVersion(
	r *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	// Parse request vars
	var errResp *Response
	var project *mlp.Project
	var router *models.Router
	var routerVersion *models.RouterVersion
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if routerVersion, errResp = c.getRouterVersionFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// Attempt tp deploy the router version
	err := c.Services.RouterVersionsService.DeployRouterVersion(project, router, routerVersion)
	if err != nil {
		return BadRequest("invalid deploy request", err.Error())
	}

	return Accepted(map[string]int{
		"router_id": int(router.ID),
		"version":   int(routerVersion.Version),
	})
}

func (c RouterVersionsController) Routes() []Route {
	return []Route{
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/versions",
			handler: c.ListRouterVersions,
		},
		{
			method:  http.MethodPost,
			path:    "/projects/{project_id}/routers/{router_id}/versions",
			body:    request.RouterConfig{},
			handler: c.CreateRouterVersion,
		},
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/versions/{version}",
			handler: c.GetRouterVersion,
		},
		{
			method:  http.MethodDelete,
			path:    "/projects/{project_id}/routers/{router_id}/versions/{version}",
			handler: c.DeleteRouterVersion,
		},
		{
			method:  http.MethodPost,
			path:    "/projects/{project_id}/routers/{router_id}/versions/{version}/deploy",
			handler: c.DeployRouterVersion,
		},
	}
}
