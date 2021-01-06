package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gojek/turing/api/turing/service"

	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/cluster/servicebuilder"
	"github.com/gojek/turing/api/turing/models"
)

type PodLogController struct {
	*baseController
}

func (c *PodLogController) ListPodLogs(r *http.Request, vars map[string]string, body interface{}) *Response {
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
	if vars["version"] != "" {
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
		routerVersion, err = c.RouterVersionsService.FindByID(uint(router.CurrRouterVersionID.Int32))
		if err != nil {
			return InternalServerError("Failed to find current router version", err.Error())
		}
	}

	var componentType string
	switch vars["component_type"] {
	case servicebuilder.ComponentTypes.Router,
		servicebuilder.ComponentTypes.Enricher,
		servicebuilder.ComponentTypes.Ensembler:
		componentType = vars["component_type"]
	case "":
		componentType = "router"
	default:
		return BadRequest(fmt.Sprintf("Invalid component type '%s'", vars["component_type"]),
			"must be one of router, enricher or ensembler")
	}

	opts := &service.PodLogOptions{}
	if vars["container"] != "" {
		opts.Container = vars["container"]
	}

	if vars["previous"] != "" {
		previous, err := strconv.ParseBool(vars["previous"])
		if err != nil {
			return BadRequest("Query string 'previous' must be a truthy value", err.Error())
		}
		opts.Previous = previous
	}

	if vars["since_time"] != "" {
		t, err := time.Parse(time.RFC3339, vars["since_time"])
		if err != nil {
			return BadRequest("Query string 'since_time' must be in RFC3339 format", err.Error())
		}
		opts.SinceTime = &t
	}

	if vars["tail_lines"] != "" {
		i, err := strconv.ParseInt(vars["tail_lines"], 10, 64)
		if err != nil {
			return BadRequest("Query string 'tail_lines' must be a positive number", err.Error())
		}
		if i <= 0 {
			return BadRequest("Query string 'tail_lines' must be a positive number", "")
		}
		opts.TailLines = &i
	}

	if vars["head_lines"] != "" {
		i, err := strconv.ParseInt(vars["head_lines"], 10, 64)
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
