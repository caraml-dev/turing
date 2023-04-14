package api

import (
	"net/http"

	"github.com/caraml-dev/mlp/api/pkg/authz/enforcer"
)

type ProjectsController struct {
	BaseController
}

func (c ProjectsController) ListProjects(
	r *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	projectName, _ := vars.get("name")
	projects, err := c.MLPService.GetProjects(projectName)
	if err != nil {
		return InternalServerError("failed to fetch projects", err.Error())
	}

	if c.Authorizer != nil {
		user := r.Header.Get("User-Email")
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
