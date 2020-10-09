package api

import (
	"os"

	"github.com/gojek/mlp/pkg/authz/enforcer"
	"github.com/gojek/mlp/pkg/vault"
	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/middleware"
	"github.com/gojek/turing/api/turing/service"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/jinzhu/gorm"
	"github.com/xanzy/go-gitlab"
)

// AppContext stores the entities relating to the application's context
type AppContext struct {
	Authorizer        *middleware.Authorizer
	OpenAPIValidation *middleware.OpenAPIValidation

	// DAO
	DeploymentService     service.DeploymentService
	RoutersService        service.RoutersService
	RouterVersionsService service.RouterVersionsService
	EventService          service.EventService
	AlertService          service.AlertService

	// Default configuration for routers
	RouterDefaults *config.RouterDefaults

	CryptoService      service.CryptoService
	MLPService         service.MLPService
	ExperimentsService service.ExperimentsService
	PodLogService      service.PodLogService
}

// NewAppContext is a creator for the app context
func NewAppContext(
	db *gorm.DB,
	cfg *config.Config,
	authEnforcer *enforcer.Enforcer,
	vaultClient vault.VaultClient,
) (*AppContext, error) {
	// Init Authorizer
	var authorizer *middleware.Authorizer
	var err error
	if authEnforcer != nil {
		authorizer, err = middleware.NewAuthorizer(*authEnforcer)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed initializing Authorizer")
		}
	}

	// Init Experiments Service
	expSvc, err := service.NewExperimentsService(cfg.LitmusConfig, cfg.XPConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed initializing Experiments Service")
	}

	// Init Crypto Service
	cryptoService := service.NewCryptoService(cfg.TuringEncryptionKey)

	// Init MLP service
	mlpSvc, err := service.NewMLPService(cfg.MLPConfig.MLPURL, cfg.MLPConfig.MLPEncryptionKey,
		cfg.MLPConfig.MerlinURL)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed initializing MLP Service")
	}

	// Create a map of env name to cluster name for each supported deployment environment
	envClusterMap, err := getEnvironmentClusterMap(mlpSvc, cfg.DeployConfig.GcpProject)
	if err != nil {
		return nil, errors.Wrapf(err, "Error obtaining environment info from MLP Service")
	}

	// Initialise cluster controllers
	clusterControllers, err := cluster.InitClusterControllers(cfg, envClusterMap, vaultClient)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed initializing cluster controllers")
	}

	appContext := &AppContext{
		Authorizer:            authorizer,
		DeploymentService:     service.NewDeploymentService(cfg, clusterControllers),
		RoutersService:        service.NewRoutersService(db),
		RouterVersionsService: service.NewRouterVersionsService(db),
		EventService:          service.NewEventService(db),
		RouterDefaults:        cfg.RouterDefaults,
		CryptoService:         cryptoService,
		MLPService:            mlpSvc,
		ExperimentsService:    expSvc,
		PodLogService:         service.NewPodLogService(clusterControllers),
	}

	if cfg.AlertConfig.Enabled {
		gitlabClient, err := gitlab.NewClient(
			cfg.AlertConfig.GitLab.Token,
			gitlab.WithBaseURL(cfg.AlertConfig.GitLab.BaseURL),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed initializing GitLab client. GitLab config: %v",
				cfg.AlertConfig.GitLab)
		}
		appContext.AlertService = service.NewGitlabOpsAlertService(db, gitlabClient, cfg.AlertConfig.GitLab.ProjectID,
			cfg.AlertConfig.GitLab.Branch, cfg.AlertConfig.GitLab.PathPrefix)
	}

	// Initialize OpenAPI validation middleware
	if _, err = os.Stat(cfg.SwaggerFile); os.IsExist(err) {
		return nil, errors.Wrapf(err, "Swagger spec file not found")
	}
	appContext.OpenAPIValidation, err = middleware.NewOpenAPIV2Validation(
		cfg.SwaggerFile,
		middleware.OpenAPIValidationOptions{
			// Authentication is ignored because it is handled by another middleware
			IgnoreAuthentication: true,
			// Servers declaration (e.g. validating the Host value in http request) in Swagger is
			// ignored so that the configuration is simpler (since this server value can change depends on
			// where Turing API is deployed, localhost or staging/production environment).
			//
			// Validating path parameters, request and response body is the most useful in typical cases.
			IgnoreServers: true,
		},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to initialize OpenAPI Validation middleware")
	}

	return appContext, nil
}

// getEnvironmentClusterMap creates a map of the environment name to the
// kubernetes cluster, that belong to the configured GCP project.
func getEnvironmentClusterMap(
	mlpSvc service.MLPService,
	gcpProject string,
) (map[string]string, error) {
	envClusterMap := map[string]string{}
	// Get all environments
	environments, err := mlpSvc.GetEnvironments()
	if err != nil {
		return envClusterMap, err
	}
	// Create a map of the environment name to cluster id
	for _, environment := range environments {
		if environment.GcpProject == gcpProject {
			envClusterMap[environment.Name] = environment.Cluster
		}
	}
	return envClusterMap, nil
}
