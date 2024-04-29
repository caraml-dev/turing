package api

import (
	"fmt"
	"net/http"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/log"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
)

type EnsemblerImagesController struct {
	BaseController
}

func (c EnsemblerImagesController) ListImages(
	_ *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	options := service.EnsemblerImagesListOptions{}
	if err := c.ParseVars(&options, vars); err != nil {
		return BadRequest("failed to list ensembler images",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	project, err := c.MLPService.GetProject(options.ProjectID)
	if err != nil {
		return InternalServerError("unable to get MLP project for the router", err.Error())
	}

	ensembler, err := c.EnsemblersService.FindByID(
		options.EnsemblerID,
		service.EnsemblersFindByIDOptions{
			ProjectID: &options.ProjectID,
		})
	if err != nil {
		return NotFound("ensembler not found", err.Error())
	}

	pyFuncEnsembler, isPyfunc := ensembler.(*models.PyFuncEnsembler)
	if !isPyfunc {
		return InternalServerError("unable to list ensembler image", "ensembler is not a PyFuncEnsembler")
	}

	images, err := c.EnsemblerImagesService.ListImages(project, pyFuncEnsembler, options.EnsemblerRunnerType)
	if err != nil {
		return InternalServerError("unable to list ensembler images", err.Error())
	}

	return Ok(images)
}

func (c EnsemblerImagesController) BuildImage(
	r *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	options := EnsemblerImagesPathOptions{}
	if err := c.ParseVars(&options, vars); err != nil {
		return BadRequest("failed to fetch ensembler",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	project, err := c.MLPService.GetProject(*options.ProjectID)
	if err != nil {
		return InternalServerError("unable to get MLP project for the router", err.Error())
	}

	ensembler, err := c.EnsemblersService.FindByID(
		*options.EnsemblerID,
		service.EnsemblersFindByIDOptions{
			ProjectID: options.ProjectID,
		})
	if err != nil {
		return NotFound("ensembler not found", err.Error())
	}

	pyFuncEnsembler, isPyfunc := ensembler.(*models.PyFuncEnsembler)
	if !isPyfunc {
		return InternalServerError("unable to build ensembler image", "ensembler is not a PyFuncEnsembler")
	}

	request := body.(*request.BuildEnsemblerImageRequest)

	go func() {
		if err := c.EnsemblerImagesService.BuildImage(project, pyFuncEnsembler, request.RunnerType); err != nil {
			log.Errorf("unable to build ensembler image", err.Error())
		}
	}()

	return Accepted(nil)
}

func (c EnsemblerImagesController) Routes() []Route {
	return []Route{
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/ensemblers/{ensembler_id}/images",
			handler: c.ListImages,
		},
		{
			method:  http.MethodPost,
			path:    "/projects/{project_id}/ensemblers/{ensembler_id}/images",
			body:    request.BuildEnsemblerImageRequest{},
			handler: c.BuildImage,
		},
	}
}

type EnsemblerImagesPathOptions struct {
	ProjectID   *models.ID `schema:"project_id" validate:"required"`
	EnsemblerID *models.ID `schema:"ensembler_id" validate:"required"`
}
