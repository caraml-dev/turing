package api

import (
	"net/http"

	"github.com/gojek/mlp/api/pkg/authz/enforcer"
)

type ProjectsController struct {
	BaseController
}

func (c ProjectsController) ListProjects(
	_ *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	projectName, _ := vars.get("name")
	projects, err := c.MLPService.GetProjects(projectName)
	if err != nil {
		return InternalServerError("failed to fetch projects", err.Error())
	}

	if c.Authorizer != nil {
		user, _ := vars.get("user")
		projects, err = c.Authorizer.FilterAuthorizedProjects(user, projects, enforcer.ActionRead)
		if err != nil {
			return InternalServerError("failed to fetch projects", err.Error())
		}
	}

	return Ok(projects)
}

func (c ProjectsController) Routes() []Route {
	return []Route{
		{
			method:  http.MethodGet,
			path:    "/projects",
			handler: c.ListProjects,
		},
	}
}
