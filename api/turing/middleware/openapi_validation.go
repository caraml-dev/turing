package middleware

import (
	"context"
	"fmt"
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
	routers []*openapi3filter.Router
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

// NewOpenAPIValidation creates open api validations with multiple swagger files
func NewOpenAPIValidation(
	swaggerYamlFiles []SwaggerYamlFile,
	options OpenAPIValidationOptions,
) (*OpenAPIValidation, error) {

	var routers []*openapi3filter.Router
	for _, file := range swaggerYamlFiles {
		var swagger *openapi3.Swagger
		if file.Type == SwaggerV2Type {
			data, err := ioutil.ReadFile(file.File)
			if err != nil {
				return nil, err
			}

			v2Swagger := &openapi2.Swagger{}
			if err := yaml.Unmarshal(data, v2Swagger); err != nil {
				return nil, err
			}

			// The "kin-openapi" library we're using to perform validation expects OpenAPI v3 to perform validation.
			// So we need to convert the source spec from OpenAPI v2 to v3.
			swagger, err = openapi2conv.ToV3Swagger(v2Swagger)
			if err != nil {
				return nil, err
			}
		} else {
			swaggerLoader := &openapi3.SwaggerLoader{
				IsExternalRefsAllowed: true,
			}
			var err error
			swagger, err = swaggerLoader.LoadSwaggerFromFile(file.File)
			if err != nil {
				return nil, err
			}
		}

		if options.IgnoreServers {
			swagger.Servers = nil
		}
		router := openapi3filter.NewRouter().WithSwagger(swagger)
		routers = append(routers, router)
	}

	opts := &openapi3filter.Options{}
	if options.IgnoreAuthentication {
		// when IgnoreAuthentication is true, the authentication function always succeed i.e returns nil error.
		opts.AuthenticationFunc = func(_ context.Context, _ *openapi3filter.AuthenticationInput) error {
			return nil
		}
	}

	return &OpenAPIValidation{routers: routers, options: opts}, nil
}

// Validate the request against the OpenAPI spec
func (openapi *OpenAPIValidation) Validate(r *http.Request) error {
	for idx, router := range openapi.routers {
		route, pathParams, _ := router.FindRoute(r.Method, r.URL)
		if route == nil && idx == len(openapi.routers)-1 {
			// endpoint is not described
			return fmt.Errorf("Route `%s %s` is not described in openapi spec", r.Method, r.URL)
		}

		if route != nil {
			input := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
				Options:    openapi.options,
			}
			return openapi3filter.ValidateRequest(context.Background(), input)
		}
	}
	return nil
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
