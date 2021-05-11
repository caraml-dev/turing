package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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

	// Ensembler URI will be a local directory
	// Dockerfile will build copy the artifact into the local directory.
	// See engines/batch-ensembler/app.Dockerfile
	artifactFolderName, err := getEnsemblerFolderName(ensembler)
	if err != nil {
		return BadRequest("ensembler not supported", err.Error())
	}
	request.EnsemblerConfig.EnsemblerConfig.Spec.Ensembler.Uri = fmt.Sprintf("/home/spark/%s", artifactFolderName)

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

func getEnsemblerFolderName(ensembler models.EnsemblerLike) (string, error) {
	switch ensembler.(type) {
	case *models.PyFuncEnsembler:
		pyFuncEnsembler := ensembler.(*models.PyFuncEnsembler)
		splitURI := strings.Split(pyFuncEnsembler.ArtifactURI, "/")
		return splitURI[len(splitURI)-1], nil
	default:
		return "", errors.New("only pyfunc ensemblers are supported for now")
	}
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
