package server

import (
	"net/http"
	"os"
	"path"

	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/gojek/turing/api/turing/api"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/middleware"
	"github.com/gojek/turing/api/turing/validation"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/cors"
)

func AddAPIRoutesHandler(r *mux.Router, path string, appCtx *api.AppContext, cfg *config.Config) error {
	apiRouter := r.PathPrefix(path).Subrouter().StrictSlash(true)

	// Add Middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: cfg.AllowedOrigins,
	})
	apiRouter.Use(corsHandler.Handler)

	if appCtx.Authorizer != nil {
		apiRouter.Use(appCtx.Authorizer.Middleware)
	}

	openapiMiddleware, err := openapiValidationMiddleware(path, cfg)
	if err != nil {
		return err
	}

	apiRouter.Use(openapiMiddleware, sentry.Recoverer)

	validator, _ := validation.NewValidator(appCtx.ExperimentsService)
	baseController := api.NewBaseController(appCtx, validator)
	deploymentController := api.RouterDeploymentController{BaseController: baseController}
	controllers := []api.Controller{
		api.AlertsController{BaseController: baseController},
		api.EnsemblersController{BaseController: baseController},
		api.ExperimentsController{BaseController: baseController},
		api.PodLogController{BaseController: baseController},
		api.ProjectsController{BaseController: baseController},
		api.RoutersController{RouterDeploymentController: deploymentController},
		api.RouterVersionsController{RouterDeploymentController: deploymentController},
	}

	if cfg.BatchEnsemblingConfig.Enabled {
		controllers = append(controllers, api.EnsemblingJobController{BaseController: baseController})
	}

	for _, c := range controllers {
		for _, route := range c.Routes() {
			// NewRelic handler
			_, handler := newrelic.WrapHandle(route.Name(), route.HandlerFunc(validator))

			apiRouter.Name(route.Name()).
				Methods(route.Method()).
				Path(route.Path()).
				Handler(handler)
		}
	}

	return nil
}

func openapiValidationMiddleware(apiPath string, cfg *config.Config) (mux.MiddlewareFunc, error) {
	if cfg.OpenapiConfig != nil && cfg.OpenapiConfig.ValidationEnabled {
		// Choose between bundled openapi yaml file or development files
		specFile := cfg.OpenapiConfig.SpecFile
		if spaCfg := cfg.OpenapiConfig.SwaggerUIConfig; spaCfg != nil && len(spaCfg.ServingDirectory) > 0 {
			// During build time, the script will generate a file called openapi.bundle.yaml in the serving directory
			specFile = path.Join(spaCfg.ServingDirectory, cfg.OpenapiConfig.OpenapiBundleFileName)
		}

		// Initialize OpenAPI validation middleware
		if _, err := os.Stat(specFile); os.IsExist(err) {
			return nil, errors.Wrapf(err, "Openapi spec file '%s' not found", cfg.OpenapiConfig.SpecFile)
		}

		openapiValidation, err := middleware.NewOpenAPIValidation(
			specFile,
			middleware.OpenAPIValidationOptions{
				// Authentication is ignored because it is handled by another middleware
				IgnoreAuthentication: true,
				// Servers declaration (e.g. validating the Host value in http request) in Swagger is
				// ignored so that the configuration is simpler (since this server value can change depends on
				// where Turing API is deployed, localhost or staging/production environment).
				//
				// Validating path parameters, request and response body is the most useful in typical cases.
				IgnoreServers: true,
				// Turing API is deployed under "/v1/**" path
				APIPrefix: apiPath,
			},
		)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to initialize OpenAPI Validation middleware")
		}

		return openapiValidation.Middleware, nil
	}

	nopMiddleware := func(handler http.Handler) http.Handler {
		return handler
	}
	return nopMiddleware, nil
}
