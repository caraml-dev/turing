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

	request, ok := body.(*models.EnsemblingJob)
	if !ok {
		return BadRequest("Unable to parse body as ensembling job", "could not unmarshal request body")
	}

	request.ProjectID = models.ID(project.Id)

	// Check if ensembler exists
	_, err := c.EnsemblersService.FindByID(
		request.EnsemblerID,
		service.EnsemblersFindByIDOptions{
			ProjectID: &request.ProjectID,
		},
	)
	if err != nil {
		return NotFound("ensembler not found", err.Error())
	}

	_, err = c.MLPService.GetEnvironment(request.EnvironmentName)
	if err != nil {
		return BadRequest(
			"invalid environment",
			fmt.Sprintf("environment %s does not exist", request.EnvironmentName),
		)
	}

	// Save ensembling job
	err = c.EnsemblingJobService.Save(request)
	if err != nil {
		return BadRequest("could not create request", err.Error())
	}

	return Accepted(request)
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
