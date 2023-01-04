package api

import (
	"fmt"
	"text/template"

	"github.com/gojek/mlp/api/pkg/vault"
	"gorm.io/gorm"

	batchensembling "github.com/caraml-dev/turing/api/turing/batch/ensembling"
	batchrunner "github.com/caraml-dev/turing/api/turing/batch/runner"
	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/cluster/labeller"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/imagebuilder"
	"github.com/caraml-dev/turing/api/turing/middleware"
	"github.com/caraml-dev/turing/api/turing/repository"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
)

// AppContext stores the entities relating to the application's context
type AppContext struct {
	Authorizer   *middleware.Authorizer
	BatchRunners []batchrunner.BatchJobRunner
	Services     service.Services

	// Default configuration for routers
	RouterDefaults *config.RouterDefaults
}

// NewAppContext is a creator for the app context
func NewAppContext(
	db *gorm.DB,
	cfg *config.Config,
	authorizer *middleware.Authorizer,
	vaultClient vault.VaultClient,
) (*AppContext, error) {
	// Init Services
	var allServices service.Services

	// Init Experiments Service
	expSvc, err := service.NewExperimentsService(cfg.Experiment)
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
	envClusterMap, err := getEnvironmentClusterMap(mlpSvc, []string{cfg.EnsemblerServiceBuilderConfig.ClusterName})
	if err != nil {
		return nil, errors.Wrapf(err, "Error obtaining environment info from MLP Service")
	}

	if cfg.ClusterConfig.InClusterConfig && len(envClusterMap) > 1 {
		return nil, fmt.Errorf("There should only be one cluster if in cluster credentials are used")
	}

	// Initialise cluster controllers
	clusterControllers, err := cluster.InitClusterControllers(cfg, envClusterMap, vaultClient)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed initializing cluster controllers")
	}

	// Initialise Labeller
	labeller.InitKubernetesLabeller(
		cfg.KubernetesLabelConfigs.LabelPrefix,
		cfg.KubernetesLabelConfigs.Environment,
	)

	// Initialise Batch components
	// Since there is only the default environment, we will not create multiple batch runners.
	var batchJobRunners []batchrunner.BatchJobRunner
	var ensemblingJobService service.EnsemblingJobService

	// Init ensemblers service
	ensemblersService := service.NewEnsemblersService(db)

	if cfg.BatchEnsemblingConfig.Enabled {
		if cfg.BatchEnsemblingConfig.JobConfig == nil {
			return nil, errors.Wrapf(err, "BatchEnsemblingConfig.JobConfig was not set")
		}
		if cfg.BatchEnsemblingConfig.RunnerConfig == nil {
			return nil, errors.Wrapf(err, "BatchEnsemblingConfig.RunnerConfig was not set")
		}
		if cfg.BatchEnsemblingConfig.ImageBuildingConfig == nil {
			return nil, errors.Wrapf(err, "BatchEnsemblingConfig.ImageBuildingConfig was not set")
		}

		// Initialise Ensembling Job Service
		ensemblingJobService = service.NewEnsemblingJobService(
			db,
			cfg.BatchEnsemblingConfig.JobConfig.DefaultEnvironment,
			cfg.BatchEnsemblingConfig.ImageBuildingConfig.BuildNamespace,
			cfg.BatchEnsemblingConfig.LoggingURLFormat,
			cfg.BatchEnsemblingConfig.MonitoringURLFormat,
			cfg.BatchEnsemblingConfig.JobConfig.DefaultConfigurations,
			mlpSvc,
		)

		imageBuildingController, ok := clusterControllers[cfg.EnsemblerServiceBuilderConfig.ClusterName]
		if !ok {
			return nil, errors.Wrapf(err, "Failed getting the image building controller")
		}
		ensemblingImageBuilder, err := imagebuilder.NewEnsemblerJobImageBuilder(
			imageBuildingController,
			*cfg.BatchEnsemblingConfig.ImageBuildingConfig,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed initializing ensembling image builder")
		}

		batchClusterController, ok := clusterControllers[cfg.BatchEnsemblingConfig.JobConfig.DefaultEnvironment]
		if !ok {
			return nil, errors.Wrapf(err, "Failed getting the batch ensembling job controller")
		}
		batchEnsemblingController := batchensembling.NewBatchEnsemblingController(
			batchClusterController,
			mlpSvc,
			cfg.SparkAppConfig,
		)

		batchEnsemblingJobRunner := batchensembling.NewBatchEnsemblingJobRunner(
			batchEnsemblingController,
			ensemblingJobService,
			ensemblersService,
			mlpSvc,
			ensemblingImageBuilder,
			cfg.BatchEnsemblingConfig.RunnerConfig.RecordsToProcessInOneIteration,
			cfg.BatchEnsemblingConfig.RunnerConfig.MaxRetryCount,
			cfg.BatchEnsemblingConfig.ImageBuildingConfig.BuildTimeoutDuration,
			cfg.BatchEnsemblingConfig.RunnerConfig.TimeInterval,
		)
		batchJobRunners = append(batchJobRunners, batchEnsemblingJobRunner)
	}

	// Initialise EnsemblerServiceImageBuilder
	ensemblerServiceImageBuilder, err := imagebuilder.NewEnsemblerServiceImageBuilder(
		clusterControllers[cfg.EnsemblerServiceBuilderConfig.ClusterName],
		*cfg.EnsemblerServiceBuilderConfig.ImageBuildingConfig,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed initializing ensembler service builder")
	}

	// Parse monitoring URL
	var monitoringURLTemplate *template.Template
	if cfg.RouterDefaults.MonitoringURLFormat != nil {
		var err error
		monitoringURLTemplate, err = template.New("monitoringURLTemplate").Parse(*cfg.RouterDefaults.MonitoringURLFormat)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing monitoring url template")
		}
	}

	allServices = service.Services{
		RoutersService:       service.NewRoutersService(db, mlpSvc, cfg.RouterDefaults.MonitoringURLFormat),
		EnsemblersService:    ensemblersService,
		EnsemblingJobService: ensemblingJobService,
		RouterVersionsService: service.NewRouterVersionsService(
			repository.NewRoutersRepository(db),
			repository.NewRouterVersionsRepository(db),
			&allServices,
		),
		RouterDeploymentService: service.NewDeploymentService(
			cfg,
			clusterControllers,
			ensemblerServiceImageBuilder,
			&allServices,
		),
		RouterMonitoringService: service.NewRouterMonitoringService(mlpSvc, monitoringURLTemplate),
		EventService:            service.NewEventService(db),
		CryptoService:           cryptoService,
		MLPService:              mlpSvc,
		ExperimentsService:      expSvc,
		PodLogService: service.NewPodLogService(
			clusterControllers,
		),
	}
	if cfg.AlertConfig.Enabled && cfg.AlertConfig.GitLab != nil {
		allServices.AlertService, err = service.NewGitlabOpsAlertService(db, *cfg.AlertConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to initialize AlertService")
		}
	}

	return &AppContext{
		Authorizer:     authorizer,
		Services:       allServices,
		BatchRunners:   batchJobRunners,
		RouterDefaults: cfg.RouterDefaults,
	}, nil
}

// getEnvironmentClusterMap creates a map of the environment name to the kubernetes cluster. Additionally,
// clusters that are not a part of the deployment environments can be registered using the clusterNames
// parameter (in such cases, the environment name will be saved to be the same as the cluster name).
func getEnvironmentClusterMap(mlpSvc service.MLPService, clusterNames []string) (map[string]string, error) {
	envClusterMap := map[string]string{}
	// Get all environments
	environments, err := mlpSvc.GetEnvironments()
	if err != nil {
		return envClusterMap, err
	}
	// Create a map of the environment name to cluster id
	for _, environment := range environments {
		envClusterMap[environment.Name] = environment.Cluster
	}
	// Add other required clusters that are not a part of the environments
	for _, clusterName := range clusterNames {
		envClusterMap[clusterName] = clusterName
	}
	return envClusterMap, nil
}
