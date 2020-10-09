package api

import (
	"fmt"
	"net/http"
	"strings"
)

// ExperimentsController implements the handlers for experiment related APIs
type ExperimentsController struct {
	*AppContext
}

// ListExperimentEngines returns a list of available experiment engines
func (c ExperimentsController) ListExperimentEngines(
	_ *http.Request,
	_ map[string]string,
	_ interface{},
) *Response {
	return Ok(c.ExperimentsService.ListEngines())
}

// ListExperimentEngineClients returns a list of clients on the given experiment engine
func (c ExperimentsController) ListExperimentEngineClients(
	r *http.Request,
	vars map[string]string,
	_ interface{},
) *Response {
	engine, ok := vars["engine"]
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
	vars map[string]string,
	_ interface{},
) *Response {
	engine, ok := vars["engine"]
	if !ok {
		return BadRequest("invalid experiment engine", "key engine not found in vars")
	}

	// Get client ID, if supplied
	clientID := vars["client_id"]
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
	vars map[string]string,
	_ interface{},
) *Response {
	engine, ok := vars["engine"]
	if !ok {
		return BadRequest("invalid experiment engine", "key engine not found in vars")
	}

	// Get client ID, if supplied
	clientID := vars["client_id"]
	// Get experiment IDs, if supplied
	experimentIDs := []string{}
	if len(vars["experiment_id"]) > 0 {
		experimentIDs = strings.Split(vars["experiment_id"], ",")
	}
	// Get variables
	variables, err := c.ExperimentsService.ListVariables(engine, clientID, experimentIDs)
	if err != nil {
		return InternalServerError(fmt.Sprintf("error when querying %s variables", engine), err.Error())
	}

	return Ok(variables)
}
