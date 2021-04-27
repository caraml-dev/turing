package api

import (
	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gorilla/schema"
)

type Controller interface {
	Routes() []Route
}

// BaseController implements common methods that may be shared by all API controllers
type BaseController struct {
	*AppContext
	decoder *schema.Decoder
}

func NewBaseController(ctx *AppContext) *BaseController {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	return &BaseController{
		AppContext: ctx,
		decoder:    decoder,
	}
}

func (c *BaseController) getProjectFromRequestVars(vars map[string]string) (project *mlp.Project, error *Response) {
	id, err := getIDFromVars(vars, "project_id")
	if err != nil {
		return nil, BadRequest("invalid project id", err.Error())
	}
	project, err = c.MLPService.GetProject(id)
	if err != nil {
		return nil, NotFound("project not found", err.Error())
	}
	return project, nil
}

func (c *BaseController) getRouterFromRequestVars(vars map[string]string) (router *models.Router, error *Response) {
	id, err := getIDFromVars(vars, "router_id")
	if err != nil {
		return nil, BadRequest("invalid router id", err.Error())
	}
	router, err = c.RoutersService.FindByID(id)
	if err != nil {
		return nil, NotFound("router not found", err.Error())
	}
	return router, nil
}

func (c *BaseController) getRouterVersionFromRequestVars(
	vars map[string]string,
) (routerVersion *models.RouterVersion, error *Response) {
	routerID, err := getIDFromVars(vars, "router_id")
	if err != nil {
		return nil, BadRequest("invalid router id", err.Error())
	}
	versionNum, err := getIntFromVars(vars, "version")
	if err != nil {
		return nil, BadRequest("invalid router version value", err.Error())
	}
	routerVersion, err = c.RouterVersionsService.FindByRouterIDAndVersion(routerID, uint(versionNum))
	if err != nil {
		return nil, NotFound("router version not found", err.Error())
	}
	return routerVersion, nil
}

func (c *BaseController) Routes() []Route {
	return []Route{}
}
