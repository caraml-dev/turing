package api

import (
	"fmt"
	"net/http"

	mlp "github.com/caraml-dev/mlp/api/client"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/log"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/webhook"
)

type RouterVersionsController struct {
	RouterDeploymentController
}

// ListRouterVersions lists the router versions of the provided router.
func (c RouterVersionsController) ListRouterVersions(
	_ *http.Request,
	vars RequestVars, _ interface{},
) *Response {
	// Parse request vars
	var errResp *Response
	var router *models.Router
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// List router versions
	routerVersions, err := c.RouterVersionsService.ListRouterVersions(router.ID)
	if err != nil {
		return InternalServerError("unable to retrieve router versions", err.Error())
	}

	return Ok(routerVersions)
}

// CreateRouterVersion creates a router version from the provided configuration. If no router exists
// within the provided project with the provided id, this method will throw an error.
// If the update is valid, a new RouterVersion will be created but NOT deployed.
func (c RouterVersionsController) CreateRouterVersion(req *http.Request, vars RequestVars, body interface{}) *Response {
	// Parse request vars
	var (
		ctx     = req.Context()
		errResp *Response
		router  *models.Router
		project *mlp.Project
	)

	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
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
		project.Name,
		router,
		c.RouterDefaults,
		c.AppContext.CryptoService,
		c.AppContext.ExperimentsService,
		c.EnsemblersService)

	if err == nil {
		// Save router version, re-assign the value of err
		routerVersion.Status = models.RouterVersionStatusUndeployed
		routerVersion, err = c.RouterVersionsService.Save(routerVersion)
	}

	if err != nil {
		return InternalServerError("unable to create router version", err.Error())
	}

	// call webhook for router version creation event
	if errWebhook := c.webhookClient.TriggerWebhooks(
		ctx, webhook.OnRouterVersionCreated, routerVersion,
	); errWebhook != nil {
		log.Warnf(
			"Error triggering webhook for event %s, router id: %d, router version id: %d, %v",
			webhook.OnRouterVersionCreated, router.ID, routerVersion.ID, errWebhook,
		)
	}

	return Ok(routerVersion)
}

// GetRouterVersion gets the router version for the provided router id and version number.
func (c RouterVersionsController) GetRouterVersion(
	_ *http.Request,
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
	req *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	// Parse request vars
	var (
		ctx           = req.Context()
		errResp       *Response
		router        *models.Router
		routerVersion *models.RouterVersion
	)

	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if routerVersion, errResp = c.getRouterVersionFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// Check router version's status
	if routerVersion.Status == models.RouterVersionStatusPending {
		return BadRequest("invalid delete request",
			"unable to delete router version that is currently deploying")
	}

	// If router version is current, prevent delete
	if router.CurrRouterVersion != nil && routerVersion.ID == router.CurrRouterVersion.ID {
		return BadRequest("invalid delete request", "cannot delete current router configuration")
	}

	err := c.RouterVersionsService.Delete(routerVersion)
	if err != nil {
		return InternalServerError("unable to delete router version", err.Error())
	}

	// call webhook for router version deletion event
	if errWebhook := c.webhookClient.TriggerWebhooks(
		ctx, webhook.OnRouterVersionDeleted, routerVersion,
	); errWebhook != nil {
		log.Warnf(
			"Error triggering webhook for event %s, router id: %d, router version id: %d, %v",
			webhook.OnRouterVersionDeleted, router.ID, routerVersion.ID, errWebhook,
		)
	}
	return Ok(map[string]int{"router_id": int(router.ID), "version": int(routerVersion.Version)})
}

// DeployRouterVersion deploys the given router version into the associated kubernetes cluster
func (c RouterVersionsController) DeployRouterVersion(
	req *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	// Parse request vars
	var (
		ctx           = req.Context()
		errResp       *Response
		project       *mlp.Project
		router        *models.Router
		routerVersion *models.RouterVersion
	)

	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if routerVersion, errResp = c.getRouterVersionFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// Check if router is already deploying
	if router.Status == models.RouterStatusPending {
		return BadRequest("invalid deploy request",
			"router is currently deploying, cannot do another deployment")
	}

	// Check if the version is already deployed
	if routerVersion.Status == models.RouterVersionStatusDeployed {
		return BadRequest("invalid deploy request",
			"router version is already deployed")
	}

	// Deploy the version asynchronously
	go func() {
		err := c.deployOrRollbackRouter(project, router, routerVersion)
		if err != nil {
			log.Errorf("Error deploying router version %s:%s:%d: %v",
				project.Name, router.Name, routerVersion.Version, err)
		}

		// call webhook for router version deployment event
		if errWebhook := c.webhookClient.TriggerWebhooks(
			ctx, webhook.OnRouterVersionDeployed, routerVersion,
		); errWebhook != nil {
			log.Warnf(
				"Error triggering webhook for event %s, router id: %d, router version id: %d, %v",
				webhook.OnRouterVersionDeployed, router.ID, routerVersion.ID, errWebhook,
			)
		}
	}()

	return Accepted(map[string]int{
		"router_id": int(router.ID),
		"version":   int(routerVersion.Version),
	})
}
func (c RouterVersionsController) ListRouterVersionsWithFilter(
	_ *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	options := service.RouterVersionListOptions{}

	if err := c.ParseVars(&options, vars); err != nil {
		return BadRequest("failed to fetch router versions",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	routers, err := c.RouterVersionsService.ListRouterVersionsWithFilter(options)
	if err != nil {
		return InternalServerError("unable to list router version", err.Error())
	}
	return Ok(routers)
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
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/router-versions",
			handler: c.ListRouterVersionsWithFilter,
		},
	}
}
