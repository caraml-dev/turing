package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/gojek/turing/api/turing/api/request"
	"github.com/gojek/turing/api/turing/models"

	val "github.com/go-playground/validator/v10"
	"github.com/gojek/turing/api/turing/validation"
	"github.com/gorilla/mux"

	"github.com/gojek/mlp/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/pkg/instrumentation/sentry"
)

// Handler is a function that returns a Response given the request.
type Handler func(r *http.Request, vars map[string]string, body interface{}) *Response

// Route is a http route for the API.
type Route struct {
	method  string
	path    string
	body    interface{}
	handler Handler
	name    string
}

// HandlerFunc returns the HandlerFunc for this route, which validates the request and
// executes the route's Handler on the request, returning its response.
func (route Route) HandlerFunc(validator *val.Validate) http.HandlerFunc {
	var bodyType reflect.Type
	if route.body != nil {
		bodyType = reflect.TypeOf(route.body)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		for k, v := range r.URL.Query() {
			if len(v) > 0 {
				vars[k] = v[0]
			}
		}

		response := func() *Response {
			var body interface{} = nil

			if bodyType != nil {
				body = reflect.New(bodyType).Interface()
				err := json.NewDecoder(r.Body).Decode(body)
				if err == io.EOF {
					// empty body
					return route.handler(r, vars, body)
				}

				if err != nil {
					return BadRequest("invalid request body",
						fmt.Sprintf("Failed to deserialize request body: %s", err.Error()))
				}

				if err := validator.Struct(body); err != nil {
					return BadRequest("invalid request body", err.Error())
				}
			}
			return route.handler(r, vars, body)
		}()

		response.WriteTo(w)
	}
}

// NewRouter instantiates a mux.Router for this application.
func NewRouter(appCtx *AppContext) *mux.Router {
	validator, _ := validation.NewValidator(appCtx.ExperimentsService)
	baseController := &baseController{appCtx}
	deploymentController := &routerDeploymentController{baseController}
	routersController := RoutersController{deploymentController}
	routerVersionsController := RouterVersionsController{deploymentController}
	alertsController := &AlertsController{baseController}
	podLogController := PodLogController{baseController}
	experimentsController := &ExperimentsController{appCtx}

	routes := []Route{
		{
			name:    "ListRouters",
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers",
			handler: routersController.ListRouters,
		},
		{
			name:    "GetRouter",
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}",
			handler: routersController.GetRouter,
		},
		{
			name:    "CreateRouter",
			method:  http.MethodPost,
			path:    "/projects/{project_id}/routers",
			body:    request.CreateOrUpdateRouterRequest{},
			handler: routersController.CreateRouter,
		},
		{
			name:    "UpdateRouter",
			method:  http.MethodPut,
			path:    "/projects/{project_id}/routers/{router_id}",
			body:    request.CreateOrUpdateRouterRequest{},
			handler: routersController.UpdateRouter,
		},
		{
			name:    "DeleteRouter",
			method:  http.MethodDelete,
			path:    "/projects/{project_id}/routers/{router_id}",
			handler: routersController.DeleteRouter,
		},
		{
			name:    "ListRouterVersions",
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/versions",
			handler: routerVersionsController.ListRouterVersions,
		},
		{
			name:    "GetRouterVersion",
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/versions/{version}",
			handler: routerVersionsController.GetRouterVersion,
		},
		{
			name:    "DeleteRouterVersion",
			method:  http.MethodDelete,
			path:    "/projects/{project_id}/routers/{router_id}/versions/{version}",
			handler: routerVersionsController.DeleteRouterVersion,
		},
		// Deploy / Undeploy router version
		{
			name:    "DeployRouter",
			method:  http.MethodPost,
			path:    "/projects/{project_id}/routers/{router_id}/deploy",
			handler: routersController.DeployRouter,
		},
		{
			name:    "DeployRouterVersion",
			method:  http.MethodPost,
			path:    "/projects/{project_id}/routers/{router_id}/versions/{version}/deploy",
			handler: routerVersionsController.DeployRouterVersion,
		},
		{
			name:    "DeployRouter",
			method:  http.MethodPost,
			path:    "/projects/{project_id}/routers/{router_id}/undeploy",
			handler: routersController.UndeployRouter,
		},
		// Router Events
		{
			name:    "ListRouterEvents",
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/events",
			handler: routersController.ListRouterEvents,
		},
		// CRUD operations router alerts
		{
			name:    "CreateAlert",
			method:  http.MethodPost,
			path:    "/projects/{project_id}/routers/{router_id}/alerts",
			body:    models.Alert{},
			handler: alertsController.CreateAlert,
		},
		{
			name:    "ListAlerts",
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/alerts",
			handler: alertsController.ListAlerts,
		},
		{
			name:    "UpdateAlert",
			method:  http.MethodPut,
			path:    "/projects/{project_id}/routers/{router_id}/alerts/{alert_id}",
			body:    models.Alert{},
			handler: alertsController.UpdateAlert,
		},
		{
			name:    "GetAlert",
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/alerts/{alert_id}",
			body:    models.Alert{},
			handler: alertsController.GetAlert,
		},
		{
			name:    "DeleteAlert",
			method:  http.MethodDelete,
			path:    "/projects/{project_id}/routers/{router_id}/alerts/{alert_id}",
			handler: alertsController.DeleteAlert,
		},
		{
			name:    "ListPodLogs",
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/logs",
			handler: podLogController.ListPodLogs,
		},
		{
			name:    "ListPodLogs",
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/versions/{version}/logs",
			handler: podLogController.ListPodLogs,
		},
		// Experiments APIs
		{
			name:    "ListExperimentEngines",
			method:  http.MethodGet,
			path:    "/experiment-engines",
			handler: experimentsController.ListExperimentEngines,
		},
		{
			name:    "ListExperimentEngineClients",
			method:  http.MethodGet,
			path:    "/experiment-engines/{engine}/clients",
			handler: experimentsController.ListExperimentEngineClients,
		},
		{
			name:    "ListExperimentEngineExperiments",
			method:  http.MethodGet,
			path:    "/experiment-engines/{engine}/experiments",
			handler: experimentsController.ListExperimentEngineExperiments,
		},
		{
			name:    "ListExperimentEngineVariables",
			method:  http.MethodGet,
			path:    "/experiment-engines/{engine}/variables",
			handler: experimentsController.ListExperimentEngineVariables,
		},
	}

	router := mux.NewRouter().StrictSlash(true)

	for _, r := range routes {
		_, handler := newrelic.WrapHandle(r.name, r.HandlerFunc(validator))

		// Wrap with authz handler, if provided
		if appCtx.Authorizer != nil {
			handler = appCtx.Authorizer.Middleware(handler)
		}

		router.Name(r.name).
			Methods(r.method).
			Path(r.path).
			Handler(handler)
	}

	router.Use(appCtx.OpenAPIValidation.Middleware)
	router.Use(recoveryHandler)
	return router
}

func recoveryHandler(next http.Handler) http.Handler {
	return sentry.Recoverer(next)
}
