package middleware

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/ghodss/yaml"
)

// OpenAPIValidation middleware validates HTTP requests against OpenAPI spec.
type OpenAPIValidation struct {
	swagger *openapi3.Swagger
	router  *openapi3filter.Router
	options *openapi3filter.Options
}

type OpenAPIValidationOptions struct {
	// If true, ignore "securityDefinitions" in swagger.yaml.
	IgnoreAuthentication bool
	// If true, ignore "server" declarations in swagger.yaml when validating requests paths. Only consider the paths
	// relative to the server url versus checking the full paths (which include the server URL) in the requests.
	IgnoreServers bool
}

// Create OpenAPIValidation object from swagger.yaml file with OpenAPI v2
func NewOpenAPIV2Validation(swaggerYamlFile string, options OpenAPIValidationOptions) (*OpenAPIValidation, error) {
	data, err := ioutil.ReadFile(swaggerYamlFile)
	if err != nil {
		return nil, err
	}

	v2Swagger := &openapi2.Swagger{}
	if err := yaml.Unmarshal(data, v2Swagger); err != nil {
		return nil, err
	}

	// The "kin-openapi" library we're using to perform validation expects OpenAPI v3 to perform validation.
	// So we need to convert the source spec from OpenAPI v2 to v3.
	v3Swagger, err := openapi2conv.ToV3Swagger(v2Swagger)
	if err != nil {
		return nil, err
	}

	router := openapi3filter.NewRouter().WithSwagger(v3Swagger)
	opts := &openapi3filter.Options{}

	if options.IgnoreAuthentication {
		// when IgnoreAuthentication is true, the authentication function always succeed i.e returns nil error.
		opts.AuthenticationFunc = func(_ context.Context, _ *openapi3filter.AuthenticationInput) error {
			return nil
		}
	}

	if options.IgnoreServers {
		v3Swagger.Servers = nil
	}

	return &OpenAPIValidation{swagger: v3Swagger, router: router, options: opts}, nil
}

// Validate the request against the OpenAPI spec
func (openapi *OpenAPIValidation) Validate(r *http.Request) error {
	route, pathParams, _ := openapi.router.FindRoute(r.Method, r.URL)
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
