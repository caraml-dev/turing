package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/xanzy/go-gitlab"

	"github.com/caraml-dev/mlp/api/pkg/client/mlflow"

	//nolint:all
	"bou.ke/monkey"
	merlin "github.com/caraml-dev/merlin/client"
	mlpcluster "github.com/caraml-dev/mlp/api/pkg/cluster"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/resource"

	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"

	batchensembling "github.com/caraml-dev/turing/api/turing/batch/ensembling"
	batchrunner "github.com/caraml-dev/turing/api/turing/batch/runner"
	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/config"
	openapi "github.com/caraml-dev/turing/api/turing/generated"
	"github.com/caraml-dev/turing/api/turing/imagebuilder"
	"github.com/caraml-dev/turing/api/turing/service"
	svcmocks "github.com/caraml-dev/turing/api/turing/service/mocks"
)

func TestNewAppContext(t *testing.T) {
	// Create test config
	tolerationName := "batch-job"
	timeout, _ := time.ParseDuration("10s")
	delTimeout, _ := time.ParseDuration("1s")
	defaultEnvironment := "dev"

	driverCPURequest := "1"
	driverMemoryRequest := "1Gi"
	var executorReplica int32 = 2
	executorCPURequest := "1"
	executorMemoryRequest := "1Gi"

	routerMonitoringURLFormat := "http://www.example.com"

	testCfg := &config.Config{
		Port: 8080,
		BatchEnsemblingConfig: config.BatchEnsemblingConfig{
			Enabled: true,
			JobConfig: &config.JobConfig{
				DefaultEnvironment: defaultEnvironment,
				DefaultConfigurations: config.DefaultEnsemblingJobConfigurations{
					BatchEnsemblingJobResources: openapi.EnsemblingResources{
						DriverCpuRequest:      &driverCPURequest,
						DriverMemoryRequest:   &driverMemoryRequest,
						ExecutorReplica:       &executorReplica,
						ExecutorCpuRequest:    &executorCPURequest,
						ExecutorMemoryRequest: &executorMemoryRequest,
					},
					SparkConfigAnnotations: map[string]string{
						"spark/spark.sql.execution.arrow.pyspark.enabled": "true",
					},
				},
			},
			RunnerConfig: &config.RunnerConfig{
				TimeInterval:                   3 * time.Minute,
				RecordsToProcessInOneIteration: 10,
				MaxRetryCount:                  3,
			},
			ImageBuildingConfig: &config.ImageBuildingConfig{
				DestinationRegistry:  "ghcr.io",
				BaseImage:            "ghcr.io/caraml-dev/turing/pyfunc-ensembler-job:0.0.0-build.1-98b071d",
				BuildNamespace:       "default",
				BuildTimeoutDuration: 10 * time.Minute,
				KanikoConfig: config.KanikoConfig{
					BuildContextURI:    "git://github.com/caraml-dev/turing.git#refs/heads/master",
					DockerfileFilePath: "engines/pyfunc-ensembler-job/app.Dockerfile",
					Image:              "gcr.io/kaniko-project/executor",
					ImageVersion:       "v1.5.2",
					ResourceRequestsLimits: config.ResourceRequestsLimits{
						Requests: config.Resource{
							CPU:    "500m",
							Memory: "1Gi",
						},
						Limits: config.Resource{
							CPU:    "500m",
							Memory: "1Gi",
						},
					},
				},
			},
		},
		EnsemblerServiceBuilderConfig: config.EnsemblerServiceBuilderConfig{
			ClusterName: defaultEnvironment,
			ImageBuildingConfig: &config.ImageBuildingConfig{
				DestinationRegistry:  "ghcr.io",
				BaseImage:            "ghcr.io/caraml-dev/turing/pyfunc-ensembler-service:0.0.0-build.1-98b071d",
				BuildNamespace:       "default",
				BuildTimeoutDuration: 10 * time.Minute,
				KanikoConfig: config.KanikoConfig{
					BuildContextURI:    "git://github.com/caraml-dev/turing.git#refs/heads/master",
					DockerfileFilePath: "engines/pyfunc-ensembler-service/app.Dockerfile",
					Image:              "gcr.io/kaniko-project/executor",
					ImageVersion:       "v1.5.2",
					ResourceRequestsLimits: config.ResourceRequestsLimits{
						Requests: config.Resource{
							CPU:    "500m",
							Memory: "1Gi",
						},
						Limits: config.Resource{
							CPU:    "500m",
							Memory: "1Gi",
						},
					},
				},
			},
		},
		DbConfig: &config.DatabaseConfig{
			Host:     "turing-db-host",
			Port:     5432,
			User:     "turing-db-user",
			Password: "turing-db-pass",
			Database: "turing-db-name",
		},
		DeployConfig: &config.DeploymentConfig{
			EnvironmentType: defaultEnvironment,
			Timeout:         timeout,
			DeletionTimeout: delTimeout,
			MaxCPU:          config.Quantity(resource.MustParse("200m")),
			MaxMemory:       config.Quantity(resource.MustParse("100Mi")),
		},
		SparkAppConfig: &config.SparkAppConfig{
			NodeSelector: map[string]string{
				"node-workload-type": "batch",
			},
			CorePerCPURequest:              1.5,
			CPURequestToCPULimit:           1.25,
			SparkVersion:                   "2.4.5",
			TolerationName:                 &tolerationName,
			SubmissionFailureRetries:       3,
			SubmissionFailureRetryInterval: 10,
			FailureRetries:                 3,
			FailureRetryInterval:           10,
			PythonVersion:                  "3",
			TTLSecond:                      86400,
		},
		RouterDefaults: &config.RouterDefaults{
			Image:                   "asia.gcr.io/gcp-project-id/turing-router:1.0.0",
			FiberDebugLogEnabled:    true,
			CustomMetricsEnabled:    true,
			JaegerEnabled:           true,
			JaegerCollectorEndpoint: "jaeger-endpoint",
			LogLevel:                "INFO",
			FluentdConfig: &config.FluentdConfig{
				Image:                "image",
				Tag:                  "turing-result.log",
				FlushIntervalSeconds: 90,
			},
			MonitoringURLFormat: &routerMonitoringURLFormat,
		},
		KubernetesLabelConfigs: &config.KubernetesLabelConfigs{
			Environment: "dev",
		},
		Sentry: sentry.Config{
			Enabled: false,
			DSN:     "",
			Labels:  nil,
		},
		ClusterConfig: config.ClusterConfig{
			InClusterConfig: false,
			EnsemblingServiceK8sConfig: &mlpcluster.K8sConfig{
				Name: defaultEnvironment,
				Cluster: &clientcmdapiv1.Cluster{
					Server: "k8s.api.server",
				},
				AuthInfo: &clientcmdapiv1.AuthInfo{},
			},
			EnvironmentConfigs: []*config.EnvironmentConfig{
				{
					Name: "N1",
					K8sConfig: &mlpcluster.K8sConfig{
						Name: "C1",
						Cluster: &clientcmdapiv1.Cluster{
							Server: "http://k8s.c1.api.server",
						},
					},
				},
				{
					Name: "N2",
					K8sConfig: &mlpcluster.K8sConfig{
						Name: "C2",
						Cluster: &clientcmdapiv1.Cluster{
							Server: "http://k8s.c2.api.server",
						},
					},
				},
			},
			EnvironmentConfigPath: "path-to-env-file.yaml",
		},
		MLPConfig: &config.MLPConfig{
			MerlinURL: "http://mlp.example.com/api/merlin/v1",
			MLPURL:    "http://mlp.example.com/api/mlp/v1",
		},
		TuringEncryptionKey: "turing-key",
		AlertConfig: &config.AlertConfig{
			Enabled: true,
			GitLab: &config.GitlabConfig{
				Branch:     "master",
				PathPrefix: "turing",
			},
		},
		MlflowConfig: &config.MlflowConfig{
			TrackingURL:         "",
			ArtifactServiceType: "nop",
		},
	}
	// Create mock MLP Service
	mockEnvironmentID1 := int32(1)
	mockGcpProject1 := "gcp-project"
	mockEnvironmentID2 := int32(2)
	mockGcpProject2 := "gcp-project-2"

	mlpSvc := &svcmocks.MLPService{}
	mlpSvc.On("GetEnvironments").Return([]merlin.Environment{
		{
			Id:         &mockEnvironmentID1,
			Name:       "N1",
			Cluster:    "C1",
			GcpProject: &mockGcpProject1,
		},
		{
			Id:         &mockEnvironmentID2,
			Name:       "N2",
			Cluster:    "C2",
			GcpProject: &mockGcpProject2,
		},
	}, nil)

	// Patch the functions from other packages
	defer monkey.UnpatchAll()
	monkey.Patch(service.NewExperimentsService,
		func(_ map[string]config.EngineConfig) (service.ExperimentsService, error) {
			return nil, nil
		},
	)
	monkey.Patch(service.NewCryptoService,
		func(encryptionKey string) service.CryptoService {
			assert.Equal(t, testCfg.TuringEncryptionKey, encryptionKey)
			return nil
		},
	)
	monkey.Patch(service.NewMLPService,
		func(mlpBasePath string, merlinBasePath string,
		) (service.MLPService, error) {
			assert.Equal(t, testCfg.MLPConfig.MLPURL, mlpBasePath)
			assert.Equal(t, testCfg.MLPConfig.MerlinURL, merlinBasePath)
			return mlpSvc, nil
		},
	)
	monkey.Patch(cluster.InitClusterControllers,
		func(
			cfg *config.Config,
			environmentClusterMap map[string]*mlpcluster.K8sConfig,
		) (map[string]cluster.Controller, error) {
			assert.Equal(t, testCfg, cfg)
			assert.Equal(t, map[string]*mlpcluster.K8sConfig{
				"N1":  cfg.ClusterConfig.EnvironmentConfigs[0].K8sConfig,
				"N2":  cfg.ClusterConfig.EnvironmentConfigs[1].K8sConfig,
				"dev": cfg.ClusterConfig.EnsemblingServiceK8sConfig,
			}, environmentClusterMap)
			return map[string]cluster.Controller{
				defaultEnvironment: nil,
			}, nil
		},
	)
	monkey.Patch(gitlab.NewClient,
		func(token string, options ...gitlab.ClientOptionFunc) (*gitlab.Client, error) {
			assert.Equal(t, testCfg.AlertConfig.GitLab.Token, token)
			assert.Equal(t, 1, len(options))
			return nil, nil
		},
	)
	monkey.Patch(service.NewGitlabOpsAlertService,
		func(_ *gorm.DB, config config.AlertConfig) (service.AlertService, error) {
			assert.Equal(t, *testCfg.AlertConfig, config)
			return nil, nil
		},
	)

	// Create expected components
	mlpService, err := service.NewMLPService(testCfg.MLPConfig.MLPURL, testCfg.MLPConfig.MerlinURL)
	assert.NoError(t, err)
	experimentService, err := service.NewExperimentsService(testCfg.Experiment)
	assert.NoError(t, err)

	// Validate
	appCtx, err := NewAppContext(nil, testCfg)
	assert.NoError(t, err)

	alertService, err := service.NewGitlabOpsAlertService(nil, *testCfg.AlertConfig)
	assert.NoError(t, err)

	artifactService, err := initArtifactService(testCfg)
	assert.NoError(t, err)

	ensemblingImageBuilder, err := imagebuilder.NewEnsemblerJobImageBuilder(
		nil,
		*testCfg.BatchEnsemblingConfig.ImageBuildingConfig,
		artifactService,
	)
	assert.Nil(t, err)

	ensemblersService := service.NewEnsemblersService(nil)
	ensemblingJobService := service.NewEnsemblingJobService(
		nil,
		testCfg.BatchEnsemblingConfig.JobConfig.DefaultEnvironment,
		testCfg.BatchEnsemblingConfig.ImageBuildingConfig.BuildNamespace,
		testCfg.BatchEnsemblingConfig.LoggingURLFormat,
		testCfg.BatchEnsemblingConfig.MonitoringURLFormat,
		testCfg.BatchEnsemblingConfig.JobConfig.DefaultConfigurations,
		mlpSvc,
	)
	batchEnsemblingController := batchensembling.NewBatchEnsemblingController(
		nil,
		mlpSvc,
		testCfg.SparkAppConfig,
	)
	batchEnsemblingJobRunner := batchensembling.NewBatchEnsemblingJobRunner(
		batchEnsemblingController,
		ensemblingJobService,
		ensemblersService,
		mlpSvc,
		ensemblingImageBuilder,
		testCfg.BatchEnsemblingConfig.RunnerConfig.RecordsToProcessInOneIteration,
		testCfg.BatchEnsemblingConfig.RunnerConfig.MaxRetryCount,
		testCfg.BatchEnsemblingConfig.ImageBuildingConfig.BuildTimeoutDuration,
		testCfg.BatchEnsemblingConfig.RunnerConfig.TimeInterval,
	)

	ensemblerImageBuilder, err := imagebuilder.NewEnsemblerServiceImageBuilder(
		nil,
		*testCfg.EnsemblerServiceBuilderConfig.ImageBuildingConfig,
		artifactService,
	)

	assert.NoError(t, err)

	mlflowService, err := mlflow.NewMlflowService(http.DefaultClient, mlflow.Config{
		TrackingURL:         testCfg.MlflowConfig.TrackingURL,
		ArtifactServiceType: testCfg.MlflowConfig.ArtifactServiceType,
	})
	assert.NoError(t, err)

	assert.Equal(t, &AppContext{
		DeploymentService: service.NewDeploymentService(
			testCfg,
			map[string]cluster.Controller{
				defaultEnvironment: nil,
			},
			ensemblerImageBuilder,
		),
		RoutersService:         service.NewRoutersService(nil, mlpSvc, testCfg.RouterDefaults.MonitoringURLFormat),
		EnsemblersService:      ensemblersService,
		EnsemblerImagesService: service.NewEnsemblerImagesService(ensemblingImageBuilder, ensemblerImageBuilder),
		EnsemblingJobService:   ensemblingJobService,
		RouterVersionsService:  service.NewRouterVersionsService(nil, mlpSvc, testCfg.RouterDefaults.MonitoringURLFormat),
		EventService:           service.NewEventService(nil),
		RouterDefaults:         testCfg.RouterDefaults,
		CryptoService:          service.NewCryptoService(testCfg.TuringEncryptionKey),
		MLPService:             mlpService,
		ExperimentsService:     experimentService,
		PodLogService: service.NewPodLogService(
			map[string]cluster.Controller{
				defaultEnvironment: nil,
			},
		),
		AlertService:  alertService,
		BatchRunners:  []batchrunner.BatchJobRunner{batchEnsemblingJobRunner},
		MlflowService: mlflowService,
	}, appCtx)
}
