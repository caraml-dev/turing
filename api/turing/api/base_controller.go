package api

import (
	val "github.com/go-playground/validator/v10"
	mlp "github.com/gojek/mlp/api/client"
	"github.com/gorilla/schema"

	"github.com/caraml-dev/turing/api/turing/models"
)

type Controller interface {
	// Routes returns the list of routes of this controller
	Routes() []Route
}

// BaseController implements common methods that may be shared by all API controllers
type BaseController struct {
	*AppContext
	decoder   *schema.Decoder
	validator *val.Validate
}

// NewBaseController returns a new instance of BaseController
func NewBaseController(ctx *AppContext, validator *val.Validate) BaseController {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	return BaseController{
		AppContext: ctx,
		decoder:    decoder,
		validator:  validator,
	}
}

func (c BaseController) ParseVars(dst interface{}, vars RequestVars) error {
	if err := c.decoder.Decode(dst, vars); err != nil {
		return err
	}

	if c.validator != nil {
		if err := c.validator.Struct(dst); err != nil {
			return err
		}
	}
	return nil
}

func (c BaseController) getProjectFromRequestVars(vars RequestVars) (project *mlp.Project, error *Response) {
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

func (c BaseController) getRouterFromRequestVars(vars RequestVars) (router *models.Router, error *Response) {
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

func (c BaseController) getRouterVersionFromRequestVars(
	vars RequestVars,
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
