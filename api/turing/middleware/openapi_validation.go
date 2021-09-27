package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

// OpenAPIValidation middleware validates HTTP requests against OpenAPI spec.
type OpenAPIValidation struct {
	options *openapi3filter.Options
	router  routers.Router
}

type OpenAPIValidationOptions struct {
	// If true, ignore "security" in openapi.yaml.
	IgnoreAuthentication bool
	// If true, ignore "server" declarations in openapi.yaml when validating requests paths. Only consider the paths
	// relative to the server url versus checking the full paths (which include the server URL) in the requests.
	IgnoreServers bool
}

// NewOpenAPIValidation creates OpenAPIValidation object from OAS3 spec file
// and provided options
func NewOpenAPIValidation(
	openapiYamlFile string,
	options OpenAPIValidationOptions,
) (*OpenAPIValidation, error) {
	// Get Swagger specs
	loader := &openapi3.Loader{
		IsExternalRefsAllowed: true,
	}
	swagger, err := loader.LoadFromFile(openapiYamlFile)
	if err != nil {
		return nil, fmt.Errorf("Error loading swagger spec\n: %s", err)
	}

	// Handle options
	openAPIFilterOpts := &openapi3filter.Options{}
	if options.IgnoreAuthentication {
		openAPIFilterOpts.AuthenticationFunc = openapi3filter.NoopAuthenticationFunc
	}
	if options.IgnoreServers {
		swagger.Servers = nil
	}

	// Create router
	var router routers.Router
	if router, err = legacyrouter.NewRouter(swagger); err != nil {
		return nil, err
	}

	return &OpenAPIValidation{openAPIFilterOpts, router}, nil
}

// Validate the request against the OpenAPI spec
func (openapi *OpenAPIValidation) Validate(r *http.Request) error {
	route, pathParams, _ := openapi.router.FindRoute(r)
	if route == nil {
		return fmt.Errorf("route `%s %s` is not described in openapi spec", r.Method, r.URL)
	}

	input := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
		Options:    openapi.options,
	}
	return openapi3filter.ValidateRequest(context.Background(), input)
}

// Middleware returns a middleware function
func (openapi *OpenAPIValidation) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := openapi.Validate(r); err != nil {
			var errMsg string
			// err string returned can be very lengthy containing a lot of lines but the first line of error
			// is usually descriptive and useful enough, so only the first line is returned for less verbose error msg.
			errs := strings.Split(err.Error(), "\n")
			if len(errs) > 0 {
				errMsg = errs[0]
			} else {
				errMsg = "Request failed OpenAPI validation"
			}
			jsonError(w, errMsg, http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
