package api

import (
	"fmt"
	"net/http"

	mlp "github.com/gojek/mlp/api/client"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
)

// EnsemblingJobController is the HTTP controller that handles the orchestration of ensembling jobs.
type EnsemblingJobController struct {
	BaseController
}

// Create is a HTTP Post Method that creates an Ensembling job.
// This method will only return the acceptance/rejection of the job rather than the actual result of the job.
func (c EnsemblingJobController) Create(
	r *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	var errResp *Response
	var project *mlp.Project
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// Check is done in api.handlers
	job, _ := body.(*models.EnsemblingJob)
	projectID := models.ID(project.Id)

	// Check if ensembler exists
	ensembler, err := c.EnsemblersService.FindByID(
		job.EnsemblerID,
		service.EnsemblersFindByIDOptions{
			ProjectID: &projectID,
		},
	)
	if err != nil {
		return NotFound("ensembler not found", err.Error())
	}

	var pyFuncEnsembler *models.PyFuncEnsembler
	switch v := ensembler.(type) {
	case *models.PyFuncEnsembler:
		pyFuncEnsembler = ensembler.(*models.PyFuncEnsembler)
	default:
		return BadRequest("only pyfunc ensemblers allowed", fmt.Sprintf("ensembler type given: %T", v))
	}

	ensemblingJob, err := c.EnsemblingJobService.CreateEnsemblingJob(job, projectID, pyFuncEnsembler)
	if err != nil {
		return InternalServerError("could not create job request", err.Error())
	}

	return Accepted(ensemblingJob)
}

// GetEnsemblingJob is HTTP handler that will get a single EnsemblingJob
func (c EnsemblingJobController) GetEnsemblingJob(
	_ *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	options := &GetEnsemblingJobOptions{}

	if err := c.ParseVars(options, vars); err != nil {
		return BadRequest(
			"failed to fetch ensembling job",
			fmt.Sprintf("failed to parse query string: %s", err),
		)
	}

	ensemblingJob, err := c.EnsemblingJobService.FindByID(
		*options.ID,
		service.EnsemblingJobFindByIDOptions{
			ProjectID: options.ProjectID,
		},
	)

	if err != nil {
		return NotFound("ensembling job not found", err.Error())
	}

	return Ok(ensemblingJob)
}

// ListEnsemblingJobs is HTTP handler that will get a list of EnsemblingJobs
func (c EnsemblingJobController) ListEnsemblingJobs(
	_ *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	options := service.EnsemblingJobListOptions{}
	if err := c.ParseVars(&options, vars); err != nil {
		return BadRequest(
			"unable to list ensembling jobs",
			fmt.Sprintf("failed to parse query string: %s", err),
		)
	}

	results, err := c.EnsemblingJobService.List(options)
	if err != nil {
		return InternalServerError("unable to list ensemblers", err.Error())
	}

	return Ok(results)
}

type deleteEnsemblingJobResponse struct {
	ID models.ID `json:"id"`
}

// DeleteEnsemblingJob deletes and ensembling job and cancels any ongoing process.
func (c EnsemblingJobController) DeleteEnsemblingJob(
	_ *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	options := &GetEnsemblingJobOptions{}
	if err := c.ParseVars(options, vars); err != nil {
		return BadRequest(
			"failed to fetch ensembling job",
			fmt.Sprintf("failed to parse query string: %s", err),
		)
	}

	ensemblingJob, err := c.EnsemblingJobService.FindByID(
		*options.ID,
		service.EnsemblingJobFindByIDOptions{
			ProjectID: options.ProjectID,
		},
	)
	if err != nil {
		return NotFound("ensembling job not found", err.Error())
	}

	err = c.EnsemblingJobService.MarkEnsemblingJobForTermination(ensemblingJob)
	if err != nil {
		return InternalServerError("unable to delete ensembling job", err.Error())
	}

	return Accepted(
		deleteEnsemblingJobResponse{
			ID: *options.ID,
		},
	)
}

// Routes returns all the HTTP routes given by the EnsemblingJobController.
func (c EnsemblingJobController) Routes() []Route {
	return []Route{
		{
			method:  http.MethodPost,
			path:    "/projects/{project_id}/jobs",
			handler: c.Create,
			body:    models.EnsemblingJob{},
		},
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/jobs",
			handler: c.ListEnsemblingJobs,
		},
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/jobs/{job_id}",
			handler: c.GetEnsemblingJob,
		},
		{
			method:  http.MethodDelete,
			path:    "/projects/{project_id}/jobs/{job_id}",
			handler: c.DeleteEnsemblingJob,
		},
	}
}

// GetEnsemblingJobOptions is the options that can be parsed
// from query params for the GET ensembling job method
type GetEnsemblingJobOptions struct {
	ProjectID *models.ID `schema:"project_id" validate:"required"`
	ID        *models.ID `schema:"job_id" validate:"required"`
}
