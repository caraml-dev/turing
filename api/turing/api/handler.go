package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

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
	controllers := []Controller{
		RoutersController{deploymentController},
		RouterVersionsController{deploymentController},
		EnsemblersControler{baseController},
		AlertsController{baseController},
		PodLogController{baseController},
		ExperimentsController{baseController},
	}

	var routes []Route
	for _, c := range controllers {
		routes = append(routes, c.Routes()...)
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
