package api

import (
	"fmt"
	"net/http"
	"strings"

	mlp "github.com/gojek/mlp/api/client"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/models"

	"github.com/caraml-dev/turing/api/turing/log"
)

type RoutersController struct {
	RouterDeploymentController
}

// ListRouters lists all routers configured in the provided project.
// If none are found, an error will be thrown.
func (c RoutersController) ListRouters(
	r *http.Request,
	vars RequestVars, _ interface{},
) *Response {
	// Parse input
	var errResp *Response
	var project *mlp.Project
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// List routers
	routers, err := c.RoutersService.ListRouters(models.ID(project.ID), "")
	if err != nil {
		return InternalServerError("unable to list routers", err.Error())
	}

	return Ok(routers)
}

// GetRouter gets a router matching the provided routerID.
func (c RoutersController) GetRouter(
	r *http.Request,
	vars RequestVars, _ interface{},
) *Response {
	// Parse input
	var errResp *Response
	var router *models.Router
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// Return router
	return Ok(router)
}

// CreateRouter creates a router from the provided configuration. If there already exists
// a router within the provided project with the same name, this method will throw an error.
// If not, a new Router and associated RouterVersion will be created and deployed.
func (c RoutersController) CreateRouter(
	r *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	// Parse request vars
	var errResp *Response
	var project *mlp.Project
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}

	request := body.(*request.CreateOrUpdateRouterRequest)

	// check if router already exists
	router, _ := c.RoutersService.FindByProjectAndName(models.ID(project.ID), request.Name)
	if router != nil {
		return BadRequest("invalid router name",
			fmt.Sprintf("router with name %s already exists in project %d", request.Name, project.ID))
	}

	_, err := c.MLPService.GetEnvironment(request.Environment)
	if err != nil {
		return BadRequest("invalid environment", fmt.Sprintf("environment %s does not exist", request.Environment))
	}

	// if not, create
	router, err = c.RoutersService.Save(request.BuildRouter(models.ID(project.ID)))
	if err != nil {
		return InternalServerError("unable to create router", err.Error())
	}

	// then create the new version
	var routerVersion *models.RouterVersion
	if request.Config == nil {
		return InternalServerError("unable to create router", "router config is empty")
	}

	rVersion, err := request.Config.BuildRouterVersion(
		project.Name,
		router,
		c.RouterDefaults,
		c.AppContext.CryptoService,
		c.AppContext.ExperimentsService,
		c.EnsemblersService)
	if err == nil {
		// Save router version
		routerVersion, err = c.RouterVersionsService.Save(rVersion)
	}

	if err != nil {
		errorStrings := []string{err.Error()}
		// Set router status to failed and return error
		router.Status = models.RouterStatusFailed
		router, err = c.RoutersService.Save(router)
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
		}
		return InternalServerError("unable to create router", strings.Join(errorStrings, ". "))
	}

	// deploy the new version
	go func() {
		err := c.deployOrRollbackRouter(project, router, routerVersion)
		if err != nil {
			log.Errorf("Error deploying router %s:%s:%d: %v",
				project.Name, router.Name, routerVersion.Version, err)
		}
	}()

	return Ok(router)
}

// UpdateRouter updates a router from the provided configuration. If no router exists
// within the provided project with the provided id, this method will throw an error.
// If the update is valid, a new RouterVersion will be created and deployed.
func (c RoutersController) UpdateRouter(r *http.Request, vars RequestVars, body interface{}) *Response {
	// Parse request vars
	var errResp *Response
	var project *mlp.Project
	var router *models.Router
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	request := body.(*request.CreateOrUpdateRouterRequest)

	// Check if the router environment and name are unchanged
	if request.Environment != router.EnvironmentName || request.Name != router.Name {
		return BadRequest("invalid router configuration",
			"Router name and environment cannot be changed after creation")
	}

	// Check if any deployment is in progress, if yes, disallow the update
	if router.Status == models.RouterStatusPending {
		return BadRequest("invalid update request",
			"another version is currently pending deployment")
	}

	// Create new version
	var routerVersion *models.RouterVersion
	if request.Config == nil {
		return InternalServerError("unable to update router", "router config is empty")
	}

	rVersion, err := request.Config.BuildRouterVersion(
		project.Name,
		router,
		c.RouterDefaults,
		c.AppContext.CryptoService,
		c.AppContext.ExperimentsService,
		c.EnsemblersService)
	if err == nil {
		// Save router version, re-assign the value of err
		routerVersion, err = c.RouterVersionsService.Save(rVersion)
	}

	if err != nil {
		return InternalServerError("unable to update router", err.Error())
	}

	// Deploy the new version
	go func() {
		err := c.deployOrRollbackRouter(project, router, routerVersion)
		if err != nil {
			log.Errorf("Error deploying router %s:%s:%d: %v",
				project.Name, router.Name, routerVersion.Version, err)
		}
	}()

	return Ok(router)
}

// DeleteRouter deletes a router and all its associated versions.
func (c RoutersController) DeleteRouter(
	r *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	// Parse request vars
	var errResp *Response
	var router *models.Router
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// Check if there are any pending or deployed versions, if yes, disallow the delete
	if router.Status == models.RouterStatusPending ||
		router.Status == models.RouterStatusDeployed {
		return BadRequest("invalid delete request", "router is currently deployed. Undeploy it first.")
	}
	pendingRouterVersions, err := c.RouterVersionsService.ListRouterVersionsWithStatus(
		router.ID, models.RouterVersionStatusPending)
	if err != nil {
		return InternalServerError("unable to retrieve router versions", err.Error())
	}
	if len(pendingRouterVersions) > 0 {
		return BadRequest("invalid delete request", "a router version is currently pending deployment")
	}

	err = c.RoutersService.Delete(router)
	if err != nil {
		return InternalServerError("unable to delete router", err.Error())
	}
	return Ok(map[string]int{"id": int(router.ID)})
}

// DeployRouter deploys the current version of the given router into the associated
// kubernetes cluster. If there is no current version, an error is returned.
func (c RoutersController) DeployRouter(
	r *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	// Parse request vars
	var errResp *Response
	var project *mlp.Project
	var router *models.Router
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// Check if router is already deployed / deploying
	switch router.Status {
	case models.RouterStatusDeployed:
		return BadRequest("invalid deploy request", "router is already deployed")
	case models.RouterStatusPending:
		return BadRequest("invalid deploy request",
			"router is currently deploying, cannot do another deployment")
	}

	// Get the current router version. If nil, it means there is no version of the
	// router to deploy.
	if router.CurrRouterVersion == nil {
		return BadRequest("invalid deploy request", "router has no current configuration")
	}

	// Query router version to load all relationships
	routerVersion, err := c.RouterVersionsService.FindByID(router.CurrRouterVersion.ID)
	if err != nil {
		return NotFound("router version not found", err.Error())
	}

	// Deploy the version asynchronously
	go func() {
		err := c.deployOrRollbackRouter(project, router, routerVersion)
		if err != nil {
			log.Errorf("Error deploying router version %s:%s:%d: %v",
				project.Name, router.Name, routerVersion.Version, err)
		}
	}()

	return Accepted(map[string]int{
		"router_id": int(router.ID),
		"version":   int(routerVersion.Version),
	})
}

// UndeployRouter deletes the given router specs from the associated kubernetes cluster
func (c RoutersController) UndeployRouter(
	r *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	// Parse request vars
	var errResp *Response
	var project *mlp.Project
	var router *models.Router
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// Delete the deployment
	err := c.undeployRouter(project, router)
	if err != nil {
		return InternalServerError("unable to undeploy router", err.Error())
	}

	return Ok(map[string]int{"router_id": int(router.ID)})
}

func (c RoutersController) ListRouterEvents(r *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	// Parse request vars
	var errResp *Response
	var router *models.Router
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// Get events
	events, err := c.EventService.ListEvents(int(router.ID))
	if err != nil {
		return NotFound("events not found", err.Error())
	}
	return Ok(map[string][]*models.Event{"events": events})
}

func (c RoutersController) Routes() []Route {
	return []Route{
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers",
			handler: c.ListRouters,
		},
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}",
			handler: c.GetRouter,
		},
		{
			method:  http.MethodPost,
			path:    "/projects/{project_id}/routers",
			body:    request.CreateOrUpdateRouterRequest{},
			handler: c.CreateRouter,
		},
		{
			method:  http.MethodPut,
			path:    "/projects/{project_id}/routers/{router_id}",
			body:    request.CreateOrUpdateRouterRequest{},
			handler: c.UpdateRouter,
		},
		{
			method:  http.MethodDelete,
			path:    "/projects/{project_id}/routers/{router_id}",
			handler: c.DeleteRouter,
		},
		// Deploy / Undeploy router version
		{
			method:  http.MethodPost,
			path:    "/projects/{project_id}/routers/{router_id}/deploy",
			handler: c.DeployRouter,
		},
		{
			method:  http.MethodPost,
			path:    "/projects/{project_id}/routers/{router_id}/undeploy",
			handler: c.UndeployRouter,
		},
		// Router Events
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/events",
			handler: c.ListRouterEvents,
		},
	}
}
