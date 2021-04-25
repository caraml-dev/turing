package api

import (
	"fmt"
	"net/http"

	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
)

type EnsemblersController struct {
	*BaseController
}

func (c EnsemblersController) ListEnsemblers(
	r *http.Request,
	vars map[string]string, _ interface{},
) *Response {
	var query service.ListEnsemblersQuery
	if err := c.decoder.Decode(&query, r.URL.Query()); err != nil {
		return BadRequest("unable to retrieve ensemblers",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	var errResp *Response
	var project *mlp.Project
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// List ensemblers
	results, err := c.EnsemblersService.List(models.ID(project.Id), query)
	if err != nil {
		return InternalServerError("unable to list ensemblers", err.Error())
	}

	return Ok(results)
}

func (c EnsemblersController) SaveEnsembler(r *http.Request,
	vars map[string]string, _ interface{},
) *Response {
	return NotFound("Not implemented", "")
}

func (c EnsemblersController) Routes() []Route {
	return []Route{
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/ensemblers",
			handler: c.ListEnsemblers,
		},
		{
			method:  http.MethodPost,
			path:    "/projects/{project_id}/ensemblers",
			body:    models.PyFuncEnsembler{},
			handler: c.SaveEnsembler,
		},
	}
}
