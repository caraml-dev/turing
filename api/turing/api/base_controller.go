package api

import (
	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
)

type Controller interface {
	Routes() []Route
}

// baseController implements common methods that may be shared by all API controllers
type baseController struct {
	*AppContext
}

func (c *baseController) getProjectFromRequestVars(vars map[string]string) (project *mlp.Project, error *Response) {
	id, err := getIntFromVars(vars, "project_id")
	if err != nil {
		return nil, BadRequest("invalid project id", err.Error())
	}
	project, err = c.MLPService.GetProject(models.ID(id))
	if err != nil {
		return nil, NotFound("project not found", err.Error())
	}
	return project, nil
}

func (c *baseController) getRouterFromRequestVars(vars map[string]string) (router *models.Router, error *Response) {
	id, err := getIntFromVars(vars, "router_id")
	if err != nil {
		return nil, BadRequest("invalid router id", err.Error())
	}
	router, err = c.RoutersService.FindByID(models.ID(id))
	if err != nil {
		return nil, NotFound("router not found", err.Error())
	}
	return router, nil
}

func (c *baseController) getRouterVersionFromRequestVars(
	vars map[string]string,
) (routerVersion *models.RouterVersion, error *Response) {
	routerID, err := getIntFromVars(vars, "router_id")
	if err != nil {
		return nil, BadRequest("invalid router id", err.Error())
	}
	versionNum, err := getIntFromVars(vars, "version")
	if err != nil {
		return nil, BadRequest("invalid router version value", err.Error())
	}
	routerVersion, err = c.RouterVersionsService.FindByRouterIDAndVersion(models.ID(routerID), uint(versionNum))
	if err != nil {
		return nil, NotFound("router version not found", err.Error())
	}
	return routerVersion, nil
}

func (c *baseController) Routes() []Route {
	return []Route{}
}
