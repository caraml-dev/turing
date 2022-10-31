package server

import (
	"net/http"

	"github.com/gojek/mlp/api/pkg/instrumentation/newrelic"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/cors"

	"github.com/caraml-dev/turing/api/turing/api"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/middleware"
	"github.com/caraml-dev/turing/api/turing/validation"
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

func openapiValidationMiddleware(
	apiPath string,
	cfg *config.Config,
) (mux.MiddlewareFunc, error) {
	if cfg.OpenapiConfig != nil && cfg.OpenapiConfig.ValidationEnabled {
		spec, specErr := cfg.OpenapiConfig.SpecData()
		if specErr != nil {
			return nil, errors.Wrapf(specErr, "Failed to initialize OpenAPI Validation middleware")
		}
		openapiValidation, err := middleware.NewOpenAPIValidation(
			spec,
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
