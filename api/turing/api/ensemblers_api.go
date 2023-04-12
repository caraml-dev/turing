package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	mlp "github.com/caraml-dev/mlp/api/client"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
)

type EnsemblersController struct {
	BaseController
}

func (c EnsemblersController) ListEnsemblers(
	_ *http.Request,
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
	ensembler.SetProjectID(models.ID(project.ID))

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

func (c EnsemblersController) DeleteEnsembler(
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
	ensemblerID := ensembler.GetID()
	// CHECK IF STATUS ROUTER IS DEPLOYED
	routerVersionStatusActive := []models.RouterVersionStatus{
		models.RouterVersionStatusDeployed,
		models.RouterVersionStatusPending,
	}
	activeRouter, err := c.RouterVersionsService.FindActiveRouterUsingEnsembler(ensemblerID, routerVersionStatusActive)
	if err != nil {
		return InternalServerError("Delete ensembler failed", err.Error())
	}
	if len(activeRouter) >= 1 {
		return BadRequest("Delete ensembler failed", "There are active router using this ensembler")
	}

	// CHECK IF THERE ARE ANY ENSEMBLING JOBS WITH STATUS PENDING, BUILDING, RUNNING USING THE ENSEMBLER
	ensemblingJobActiveOption := service.EnsemblingJobListOptions{
		EnsemblerID: &ensemblerID,
		Statuses: []models.Status{
			models.JobPending,
			models.JobBuildingImage,
			models.JobRunning,
		},
	}
	activeEnsemblingJob, err := c.EnsemblingJobService.List(ensemblingJobActiveOption)
	if err != nil {
		return InternalServerError("Delete ensembler failed", err.Error())
	}
	if activeEnsemblingJob.Paging.Total >= 1 {
		return BadRequest("Delete ensembler failed", "There are active ensembling job using this ensembler")
	}

	// DELETING UNUSED ROUTER
	routerVersionStatusInactive := []models.RouterVersionStatus{
		models.RouterVersionStatusFailed,
		models.RouterVersionStatusUndeployed,
	}
	inactiveRouter, err := c.RouterVersionsService.FindActiveRouterUsingEnsembler(ensemblerID, routerVersionStatusInactive)
	if err != nil {
		return InternalServerError("Delete ensembler failed", err.Error())
	}
	for _, routerVersion := range inactiveRouter {
		err = c.RouterVersionsService.Delete(routerVersion)
	}

	// DELETING UNUSED ENSEMBLING JOBS
	ensemblingJobInactiveOption := service.EnsemblingJobListOptions{
		EnsemblerID: &ensemblerID,
		Statuses: []models.Status{
			models.JobFailed,
			models.JobCompleted,
			models.JobFailedBuildImage,
			models.JobFailedSubmission,
		},
	}
	inactiveEnsemblingJob, err := c.EnsemblingJobService.List(ensemblingJobInactiveOption)
	if err != nil {
		return InternalServerError("Delete ensembler failed", err.Error())
	}
	for _, ensemblingJob := range inactiveEnsemblingJob.Results {
		err = c.EnsemblingJobService.Delete(ensemblingJob)
	}

	// CHECK IF THE ENSEMBLER IS A PYFUNC ENSEMBLER
	if pyFuncEnsembler, ok := ensembler.(*models.PyFuncEnsembler); ok {
		// IF PYFUNC, ALSO DELELETE FROM MLFLOW
		s := strconv.FormatUint(uint64(pyFuncEnsembler.ExperimentID), 10)
		fmt.Println(s)
		err := c.MlflowService.DeleteExperiment(context.Background(), s, true)
		if err != nil {
			// Handle 404
			return InternalServerError("Delete Failed", err.Error())
		}

		err = c.EnsemblersService.Delete(pyFuncEnsembler)
		if err != nil {
			return InternalServerError("Delete from db failed", err.Error())
		}
	} else {
		// IF NOT PYFUNC ONLY DELETE ENSEMBLER FROM DB
		err = c.EnsemblersService.Delete(ensembler)
		if err != nil {
			return InternalServerError("Delete from db failed", err.Error())
		}
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
		{
			method:  http.MethodDelete,
			path:    "/projects/{project_id}/ensemblers/{ensembler_id}",
			handler: c.DeleteEnsembler,
		},
	}
}

type EnsemblersPathOptions struct {
	ProjectID   *models.ID `schema:"project_id" validate:"required"`
	EnsemblerID *models.ID `schema:"ensembler_id" validate:"required"`
}
