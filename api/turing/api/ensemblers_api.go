package api

import (
	"fmt"
	"net/http"

	mlp "github.com/gojek/mlp/api/client"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
)

type EnsemblersController struct {
	BaseController
}

func (c EnsemblersController) ListEnsemblers(
	r *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	options := service.EnsemblersListOptions{}

	if err := c.ParseVars(&options, vars); err != nil {
		return BadRequest("unable to list ensemblers",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	results, err := c.EnsemblersService.List(options)
	if err != nil {
		return InternalServerError("unable to list ensemblers", err.Error())
	}

	return Ok(results)
}

func (c EnsemblersController) GetEnsembler(
	_ *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	options := EnsemblersPathOptions{}

	if err := c.ParseVars(&options, vars); err != nil {
		return BadRequest("failed to fetch ensembler",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	ensembler, err := c.EnsemblersService.FindByID(
		*options.EnsemblerID,
		service.EnsemblersFindByIDOptions{
			ProjectID: options.ProjectID,
		})
	if err != nil {
		return NotFound("ensembler not found", err.Error())
	}

	return Ok(ensembler)
}

func (c EnsemblersController) CreateEnsembler(
	_ *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	var errResp *Response
	var project *mlp.Project
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}
	var err error
	ensembler := body.(*request.CreateOrUpdateEnsemblerRequest).EnsemblerLike
	ensembler.SetProjectID(models.ID(project.Id))

	ensembler, err = c.EnsemblersService.Save(ensembler)
	if err != nil {
		return InternalServerError("unable to save an ensembler", err.Error())
	}

	return Created(ensembler)
}

func (c EnsemblersController) UpdateEnsembler(
	_ *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	options := EnsemblersPathOptions{}

	if err := c.ParseVars(&options, vars); err != nil {
		return BadRequest("failed to fetch ensembler",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	ensembler, err := c.EnsemblersService.FindByID(
		*options.EnsemblerID,
		service.EnsemblersFindByIDOptions{
			ProjectID: options.ProjectID,
		})
	if err != nil {
		return NotFound("ensembler not found", err.Error())
	}

	request := body.(*request.CreateOrUpdateEnsemblerRequest)

	if ensembler.GetType() != request.GetType() {
		return BadRequest("invalid ensembler configuration",
			"Ensembler type cannot be changed after creation")
	}

	if err = ensembler.Patch(request.EnsemblerLike); err != nil {
		return BadRequest("invalid ensembler configuration", err.Error())
	}

	ensembler, err = c.EnsemblersService.Save(ensembler)
	if err != nil {
		return InternalServerError("failed to update an ensembler", err.Error())
	}

	return Ok(ensembler)
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
			body:    request.CreateOrUpdateEnsemblerRequest{},
			handler: c.CreateEnsembler,
		},
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/ensemblers/{ensembler_id}",
			handler: c.GetEnsembler,
		},
		{
			method:  http.MethodPut,
			path:    "/projects/{project_id}/ensemblers/{ensembler_id}",
			body:    request.CreateOrUpdateEnsemblerRequest{},
			handler: c.UpdateEnsembler,
		},
	}
}

type EnsemblersPathOptions struct {
	ProjectID   *models.ID `schema:"project_id" validate:"required"`
	EnsemblerID *models.ID `schema:"ensembler_id" validate:"required"`
}
