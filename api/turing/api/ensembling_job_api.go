package api

import (
	"fmt"
	"net/http"

	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
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
	request, _ := body.(*models.EnsemblingJob)
	projectID := models.ID(project.Id)

	// Check if ensembler exists
	ensembler, err := c.EnsemblersService.FindByID(
		request.EnsemblerID,
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

	ensemblingJob, err := c.EnsemblingJobService.CreateEnsemblingJob(request, projectID, pyFuncEnsembler)
	if err != nil {
		return BadRequest("could not create request", err.Error())
	}

	return Accepted(ensemblingJob)
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
	}
}
