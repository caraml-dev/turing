package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// RequestVars is an alias of map[string][]string
// This is done to make it compatible with *http.Request.URL.Query() return type
// and also to make it possible to use it with gorilla/schema Decoder
type RequestVars map[string][]string

func (vars RequestVars) get(key string) (string, bool) {
	if values, ok := vars[key]; ok && len(values) > 0 {
		return values[0], true
	}
	return "", false
}

// Handler is a function that returns a Response given the request.
type Handler func(r *http.Request, vars RequestVars, body interface{}) *Response

// Route is a http route for the API.
type Route struct {
	method  string
	path    string
	body    interface{}
	handler Handler
	name    string
}

// Method returns HTTP method of this route
func (route Route) Method() string {
	return route.method
}

// Path returns path associated with this route
func (route Route) Path() string {
	return route.path
}

// Name returns the name of the route by either using Route's property `name`
// or if it's empty – then by inferring it from the Route's `handler` function name
func (route Route) Name() string {
	if len(route.name) > 0 {
		return route.name
	}
	v := reflect.ValueOf(route.handler)
	return runtime.FuncForPC(v.Pointer()).Name()
}

// HandlerFunc returns the HandlerFunc for this route, which validates the request and
// executes the route's Handler on the request, returning its response.
func (route Route) HandlerFunc(validator *validator.Validate) http.HandlerFunc {
	var bodyType reflect.Type
	if route.body != nil {
		bodyType = reflect.TypeOf(route.body)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Query()
		vars := RequestVars(r.URL.Query())

		for k, v := range mux.Vars(r) {
			vars[k] = []string{v}
		}

		response := func() *Response {
			var body interface{}

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
