package api

import (
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
	request.ProjectID = models.ID(project.Id)

	// Check if ensembler exists
	ensembler, err := c.EnsemblersService.FindByID(
		request.EnsemblerID,
		service.EnsemblersFindByIDOptions{
			ProjectID: &request.ProjectID,
		},
	)
	if err != nil {
		return NotFound("ensembler not found", err.Error())
	}

	// Populate name if the user does not define a name for the job
	if request.Name == "" {
		request.Name = c.EnsemblingJobService.GenerateDefaultJobName(
			ensembler.GetName(),
		)
	}

	// Populate default environment
	request.EnvironmentName = c.EnsemblingJobService.GetDefaultEnvironment()

	// Populate ensembler directory
	ensemblerDirectory, err := c.EnsemblingJobService.GetEnsemblerDirectory(ensembler)
	if err != nil {
		return BadRequest("ensembler not supported", err.Error())
	}
	request.EnsemblerConfig.EnsemblerConfig.Spec.Ensembler.Uri = ensemblerDirectory

	// Populate ensembler artifact URI, error can be ignored since the type check is done prior
	artifactURI, _ := c.EnsemblingJobService.GetArtifactURI(ensembler)
	request.InfraConfig.ArtifactURI = artifactURI

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
