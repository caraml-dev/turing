package api

import (
	"fmt"
	"net/http"
	"strings"
)

// ExperimentsController implements the handlers for experiment related APIs
type ExperimentsController struct {
	BaseController
}

// ListExperimentEngines returns a list of available experiment engines
func (c ExperimentsController) ListExperimentEngines(
	_ *http.Request,
	_ RequestVars,
	_ interface{},
) *Response {
	return Ok(c.ExperimentsService.ListEngines())
}

// ListExperimentEngineClients returns a list of clients on the given experiment engine
func (c ExperimentsController) ListExperimentEngineClients(
	r *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	engine, ok := vars.get("engine")
	if !ok {
		return BadRequest("invalid experiment engine", "key engine not found in vars")
	}

	clients, err := c.ExperimentsService.ListClients(engine)
	if err != nil {
		return InternalServerError(fmt.Sprintf("error when querying %s clients", engine), err.Error())
	}

	return Ok(clients)
}

// ListExperimentEngineExperiments returns a list of experiments on the given experiment engine,
// optionally tied to the given client id
func (c ExperimentsController) ListExperimentEngineExperiments(
	r *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	engine, ok := vars.get("engine")
	if !ok {
		return BadRequest("invalid experiment engine", "key engine not found in vars")
	}

	// Get client ID, if supplied
	clientID, _ := vars.get("client_id")
	// Get experiments (optionally, tied to the client)
	experiments, err := c.ExperimentsService.ListExperiments(engine, clientID)
	if err != nil {
		return InternalServerError(fmt.Sprintf("error when querying %s experiments", engine), err.Error())
	}

	return Ok(experiments)
}

// ListExperimentEngineVariables returns a list of variables for the given client and/or experiments
func (c ExperimentsController) ListExperimentEngineVariables(
	r *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	engine, ok := vars.get("engine")
	if !ok {
		return BadRequest("invalid experiment engine", "key engine not found in vars")
	}

	// Get client ID, if supplied
	clientID, _ := vars.get("client_id")
	// Get experiment IDs, if supplied
	experimentIDStr, _ := vars.get("experiment_id")
	var experimentIDs []string
	if len(experimentIDStr) > 0 {
		experimentIDs = strings.Split(experimentIDStr, ",")
	}
	// Get variables
	variables, err := c.ExperimentsService.ListVariables(engine, clientID, experimentIDs)
	if err != nil {
		return InternalServerError(fmt.Sprintf("error when querying %s variables", engine), err.Error())
	}

	return Ok(variables)
}

func (c ExperimentsController) Routes() []Route {
	return []Route{
		{
			method:  http.MethodGet,
			path:    "/experiment-engines",
			handler: c.ListExperimentEngines,
		},
		{
			method:  http.MethodGet,
			path:    "/experiment-engines/{engine}/clients",
			handler: c.ListExperimentEngineClients,
		},
		{
			method:  http.MethodGet,
			path:    "/experiment-engines/{engine}/experiments",
			handler: c.ListExperimentEngineExperiments,
		},
		{
			method:  http.MethodGet,
			path:    "/experiment-engines/{engine}/variables",
			handler: c.ListExperimentEngineVariables,
		},
	}
}
