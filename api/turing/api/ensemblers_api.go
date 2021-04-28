package api

import (
	"encoding/json"
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
	params := struct {
		ProjectID   *models.ID `schema:"project_id" validate:"required"`
		EnsemblerID *models.ID `schema:"ensembler_id" validate:"required"`
	}{}

	if err := c.ParseVars(&params, vars); err != nil {
		return BadRequest("failed to fetch ensembler",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	ensembler, err := c.EnsemblersService.FindByID(
		*params.EnsemblerID,
		service.EnsemblersFindByIDOptions{
			ProjectID: params.ProjectID,
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

	request := body.(*CreateOrUpdateEnsemblerRequest)
	ensembler := request.Ensembler
	ensembler.SetProjectID(models.ID(project.Id))

	var err error
	if ensembler, err = c.EnsemblersService.Save(ensembler); err != nil {
		return InternalServerError("unable to save an ensembler", err.Error())
	}

	return Created(ensembler)
}

func (c EnsemblersController) UpdateEnsembler(
	_ *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	return NotFound("", "")
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
			body:    CreateOrUpdateEnsemblerRequest{},
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
			body:    CreateOrUpdateEnsemblerRequest{},
			handler: c.UpdateEnsembler,
		},
	}
}

type CreateOrUpdateEnsemblerRequest struct {
	Ensembler models.EnsemblerLike
}

func (r *CreateOrUpdateEnsemblerRequest) UnmarshalJSON(data []byte) error {
	typeCheck := struct {
		Type models.EnsemblerType
	}{}

	if err := json.Unmarshal(data, &typeCheck); err != nil {
		return err
	}

	var ensembler models.EnsemblerLike
	switch typeCheck.Type {
	case models.EnsemblerTypePyFunc:
		ensembler = &models.PyFuncEnsembler{}
	default:
		return fmt.Errorf("unsupported ensembler type: %s", typeCheck.Type)
	}

	if err := json.Unmarshal(data, ensembler); err != nil {
		return err
	}

	r.Ensembler = ensembler
	return nil
}
