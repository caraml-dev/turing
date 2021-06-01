package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
)

// OpenAPIValidation middleware validates HTTP requests against OpenAPI spec.
type OpenAPIValidation struct {
	router  *openapi3filter.Router
	options *openapi3filter.Options
}

type OpenAPIValidationOptions struct {
	// If true, ignore "security" in openapi.yaml.
	IgnoreAuthentication bool
	// If true, ignore "server" declarations in openapi.yaml when validating requests paths. Only consider the paths
	// relative to the server url versus checking the full paths (which include the server URL) in the requests.
	IgnoreServers bool
}

const (
	// SwaggerV2Type is Swagger V2 file type
	SwaggerV2Type = SwaggerType(iota)
	// SwaggerV3Type is Swagger V3 file type
	SwaggerV3Type
)

// SwaggerType is the enum value of swagger version types
type SwaggerType int

// SwaggerYamlFile Stores the type of swagger file.
type SwaggerYamlFile struct {
	Type SwaggerType
	File string
}

// NewOpenAPIValidation creates OpenAPIValidation object from OAS3 spec file
// and provided options
func NewOpenAPIValidation(
	openapiYamlFile string,
	options OpenAPIValidationOptions,
) (*OpenAPIValidation, error) {
	loader := &openapi3.SwaggerLoader{
		IsExternalRefsAllowed: true,
	}
	swagger, err := loader.LoadSwaggerFromFile(openapiYamlFile)
	if err != nil {
		return nil, err
	}

	if options.IgnoreServers {
		swagger.Servers = nil
	}
	router := openapi3filter.NewRouter().WithSwagger(swagger)

	opts := &openapi3filter.Options{}

	if options.IgnoreAuthentication {
		// when IgnoreAuthentication is true, the authentication function always succeed i.e returns nil error.
		opts.AuthenticationFunc = func(_ context.Context, _ *openapi3filter.AuthenticationInput) error {
			return nil
		}
	}

	return &OpenAPIValidation{router: router, options: opts}, nil
}

// Validate the request against the OpenAPI spec
func (openapi *OpenAPIValidation) Validate(r *http.Request) error {
	route, pathParams, _ := openapi.router.FindRoute(r.Method, r.URL)
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
