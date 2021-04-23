package api

import (
	"net/http"

	"github.com/gojek/turing/api/turing/models"
)

type EnsemblersController struct {
	*baseController
}

func (c EnsemblersController) ListEnsemblers(
	r *http.Request,
	vars map[string]string, _ interface{},
) *Response {
	return NotFound("Not implemented", "")
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
