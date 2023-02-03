package api

import (
	"fmt"

	"gorm.io/gorm"

	mlpcluster "github.com/gojek/mlp/api/pkg/cluster"

	batchensembling "github.com/caraml-dev/turing/api/turing/batch/ensembling"
	batchrunner "github.com/caraml-dev/turing/api/turing/batch/runner"
	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/cluster/labeller"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/imagebuilder"
	"github.com/caraml-dev/turing/api/turing/middleware"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
)

// AppContext stores the entities relating to the application's context
type AppContext struct {
	Authorizer *middleware.Authorizer
	// DAO
	DeploymentService     service.DeploymentService
	RoutersService        service.RoutersService
	RouterVersionsService service.RouterVersionsService
	EventService          service.EventService
	EnsemblersService     service.EnsemblersService
	EnsemblingJobService  service.EnsemblingJobService
	AlertService          service.AlertService

	// Default configuration for routers
	RouterDefaults *config.RouterDefaults

	BatchRunners       []batchrunner.BatchJobRunner
	CryptoService      service.CryptoService
	MLPService         service.MLPService
	ExperimentsService service.ExperimentsService
	PodLogService      service.PodLogService
}

// NewAppContext is a creator for the app context
func NewAppContext(
	db *gorm.DB,
	cfg *config.Config,
	authorizer *middleware.Authorizer,
) (*AppContext, error) {
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

	envClusterMap, err := buildKubeconfigStore(mlpSvc, cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "Error obtaining environment info from MLP Service and constructing kubeconfig store")
	}

	if cfg.ClusterConfig.InClusterConfig && len(envClusterMap) > 1 {
		return nil, fmt.Errorf("There should only be one cluster if in cluster credentials are used")
	}

	// Initialise cluster controllers
	clusterControllers, err := cluster.InitClusterControllers(cfg, envClusterMap)
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
			return nil, fmt.Errorf("Failed getting the batch ensembling job controller")
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

	appContext := &AppContext{
		Authorizer:            authorizer,
		DeploymentService:     service.NewDeploymentService(cfg, clusterControllers, ensemblerServiceImageBuilder),
		RoutersService:        service.NewRoutersService(db, mlpSvc, cfg.RouterDefaults.MonitoringURLFormat),
		EnsemblersService:     ensemblersService,
		EnsemblingJobService:  ensemblingJobService,
		RouterVersionsService: service.NewRouterVersionsService(db, mlpSvc, cfg.RouterDefaults.MonitoringURLFormat),
		EventService:          service.NewEventService(db),
		RouterDefaults:        cfg.RouterDefaults,
		CryptoService:         cryptoService,
		MLPService:            mlpSvc,
		ExperimentsService:    expSvc,
		PodLogService: service.NewPodLogService(
			clusterControllers,
		),
		BatchRunners: batchJobRunners,
	}

	if cfg.AlertConfig.Enabled && cfg.AlertConfig.GitLab != nil {
		appContext.AlertService, err = service.NewGitlabOpsAlertService(db, *cfg.AlertConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to initialize AlertService")
		}
	}

	return appContext, nil
}

// buildKubeconfigStore creates a map of the environment name to the kubernetes cluster.
// It combines the EnsemblerServiceBuilderConfig with a list of environments retrieved from mlpSvc
// into a map. Each environment retrieved from mlpSvc should have a corresponding k8sConfig, else
// an error is returned.
func buildKubeconfigStore(mlpSvc service.MLPService, cfg *config.Config) (map[string]*mlpcluster.K8sConfig, error) {
	// Create a map of env name to cluster name for each supported deployment environment
	k8sConfigStore := make(map[string]*mlpcluster.K8sConfig)
	if !cfg.ClusterConfig.InClusterConfig {
		k8sConfigStore[cfg.EnsemblerServiceBuilderConfig.ClusterName] = cfg.ClusterConfig.EnsemblingServiceK8sConfig
	} else {
		k8sConfigStore[cfg.EnsemblerServiceBuilderConfig.ClusterName] = nil
	}
	for _, envconfig := range cfg.ClusterConfig.EnvironmentConfigs {
		k8sConfigStore[envconfig.Name] = envconfig.K8sConfig
	}

	// Get all environments
	environments, err := mlpSvc.GetEnvironments()
	if err != nil {
		return k8sConfigStore, err
	}
	// check if k8s store contains kubeconfig for envs received from merlin
	for _, environment := range environments {
		// check if clusterConfigs have k8sconfig for environment name
		if _, ok := k8sConfigStore[environment.Name]; !ok {
			return nil, fmt.Errorf("Missing k8sconfig for cluster %s", environment.Cluster)
		}
	}
	return k8sConfigStore, nil
}
