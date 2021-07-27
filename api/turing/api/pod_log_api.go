package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gojek/turing/api/turing/service"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/cluster/servicebuilder"
	"github.com/gojek/turing/api/turing/models"
)

type PodLogController struct {
	BaseController
}

func (c PodLogController) ListPodLogs(_ *http.Request, vars RequestVars, _ interface{}) *Response {
	// Parse input
	var errResp *Response
	var project *mlp.Project
	var router *models.Router
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	var routerVersion *models.RouterVersion
	var err error
	if _, ok := vars.get("version"); ok {
		// Specific router version value is requested
		routerVersion, errResp = c.getRouterVersionFromRequestVars(vars)
		if errResp != nil {
			return errResp
		}
	} else {
		// Unspecified router version, so default to current router version (assumed to be not null i.e valid value)
		if !router.CurrRouterVersionID.Valid {
			return BadRequest("Current router version id is invalid", "Make sure current router is deployed")
		}
		routerVersion, err = c.RouterVersionsService.FindByID(models.ID(router.CurrRouterVersionID.Int32))
		if err != nil {
			return InternalServerError("Failed to find current router version", err.Error())
		}
	}

	var componentType string
	switch componentType, _ = vars.get("component_type"); componentType {
	case servicebuilder.ComponentTypes.Router,
		servicebuilder.ComponentTypes.Enricher,
		servicebuilder.ComponentTypes.Ensembler:
	case "":
		componentType = "router"
	default:
		return BadRequest(fmt.Sprintf("Invalid component type '%s'", componentType),
			"must be one of router, enricher or ensembler")
	}

	opts := &service.PodLogOptions{}
	opts.Container, _ = vars.get("container")

	if previous, ok := vars.get("previous"); ok {
		previous, err := strconv.ParseBool(previous)
		if err != nil {
			return BadRequest("Query string 'previous' must be a truthy value", err.Error())
		}
		opts.Previous = previous
	}

	if sinceTime, ok := vars.get("since_time"); ok {
		t, err := time.Parse(time.RFC3339, sinceTime)
		if err != nil {
			return BadRequest("Query string 'since_time' must be in RFC3339 format", err.Error())
		}
		opts.SinceTime = &t
	}

	if tailTimes, ok := vars.get("tail_lines"); ok {
		i, err := strconv.ParseInt(tailTimes, 10, 64)
		if err != nil {
			return BadRequest("Query string 'tail_lines' must be a positive number", err.Error())
		}
		if i <= 0 {
			return BadRequest("Query string 'tail_lines' must be a positive number", "")
		}
		opts.TailLines = &i
	}

	if headLines, ok := vars.get("head_lines"); ok {
		i, err := strconv.ParseInt(headLines, 10, 64)
		if err != nil {
			return BadRequest("Query string 'head_lines' must be a positive number", err.Error())
		}
		if i <= 0 {
			return BadRequest("Query string 'head_lines' must be a positive number", "")
		}
		opts.HeadLines = &i
	}

	logs, err := c.PodLogService.ListPodLogs(project, router, routerVersion, componentType, opts)
	if err != nil {
		return InternalServerError("Failed to list logs", err.Error())
	}

	return Ok(logs)
}

func (c PodLogController) Routes() []Route {
	return []Route{
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/logs",
			handler: c.ListPodLogs,
		},
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/versions/{version}/logs",
			handler: c.ListPodLogs,
		},
	}
}
