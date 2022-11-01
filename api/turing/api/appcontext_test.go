package api

import (
	"testing"
	"time"

	"bou.ke/monkey"
	merlin "github.com/gojek/merlin/client"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/gojek/mlp/api/pkg/vault"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/resource"

	batchensembling "github.com/caraml-dev/turing/api/turing/batch/ensembling"
	batchrunner "github.com/caraml-dev/turing/api/turing/batch/runner"
	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/config"
	openapi "github.com/caraml-dev/turing/api/turing/generated"
	"github.com/caraml-dev/turing/api/turing/imagebuilder"
	"github.com/caraml-dev/turing/api/turing/middleware"
	"github.com/caraml-dev/turing/api/turing/service"
	svcmocks "github.com/caraml-dev/turing/api/turing/service/mocks"
)

// MockVaultClient satisfies the vault.VaultClient interface
type MockVaultClient struct{}

func (c *MockVaultClient) GetClusterSecret(clusterName string) (*vault.ClusterSecret, error) {
	return nil, nil
}

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
				DestinationRegistry: "ghcr.io",
				BaseImageRef: map[string]string{
					"3.7.*": "ghcr.io/caraml-dev/turing/pyfunc-ensembler-job:0.0.0-build.1-98b071d",
				},
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
			DefaultEnvironment: defaultEnvironment,
			ImageBuildingConfig: &config.ImageBuildingConfig{
				DestinationRegistry: "ghcr.io",
				BaseImageRef: map[string]string{
					"3.7.*": "ghcr.io/caraml-dev/turing/pyfunc-ensembler-service:0.0.0-build.1-98b071d",
				},
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
		AuthConfig: &config.AuthorizationConfig{
			Enabled: true,
			URL:     "test-auth-url",
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
			VaultConfig: &config.VaultConfig{
				Address: "vault-addr",
				Token:   "vault-token",
			},
		},
		MLPConfig: &config.MLPConfig{
			MerlinURL:        "http://mlp.example.com/api/merlin/v1",
			MLPURL:           "http://mlp.example.com/api/mlp/v1",
			MLPEncryptionKey: "key",
		},
		TuringEncryptionKey: "turing-key",
		AlertConfig: &config.AlertConfig{
			Enabled: true,
			GitLab: &config.GitlabConfig{
				Branch:     "master",
				PathPrefix: "turing",
			},
		},
	}
	// Create test auth enforcer and Vault client
	testAuthorizer := &middleware.Authorizer{}
	testVaultClient := &MockVaultClient{}
	// Create mock MLP Service
	mlpSvc := &svcmocks.MLPService{}
	mlpSvc.On("GetEnvironments").Return([]merlin.Environment{
		{
			Id:         1,
			Name:       "N1",
			Cluster:    "C1",
			GcpProject: "gcp-project",
		},
		{
			Id:         2,
			Name:       "N2",
			Cluster:    "C2",
			GcpProject: "gcp-project-2",
		},
	}, nil)

	// Patch the functions from other packages
	defer monkey.UnpatchAll()
	monkey.Patch(service.NewExperimentsService,
		func(experimentConfig map[string]config.EngineConfig) (service.ExperimentsService, error) {
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
		func(mlpBasePath string, mlpEncryptionKey string, merlinBasePath string,
		) (service.MLPService, error) {
			assert.Equal(t, testCfg.MLPConfig.MLPURL, mlpBasePath)
			assert.Equal(t, testCfg.MLPConfig.MLPEncryptionKey, mlpEncryptionKey)
			assert.Equal(t, testCfg.MLPConfig.MerlinURL, merlinBasePath)
			return mlpSvc, nil
		},
	)
	monkey.Patch(cluster.InitClusterControllers,
		func(
			cfg *config.Config,
			environmentClusterMap map[string]string,
			vaultClient vault.VaultClient,
		) (map[string]cluster.Controller, error) {
			assert.Equal(t, testCfg, cfg)
			assert.Equal(t, map[string]string{
				"N1": "C1",
				"N2": "C2",
			}, environmentClusterMap)
			assert.Equal(t, testVaultClient, vaultClient)
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
		func(db *gorm.DB, config config.AlertConfig) (service.AlertService, error) {
			assert.Equal(t, *testCfg.AlertConfig, config)
			return nil, nil
		},
	)

	// Create expected components
	mlpService, err := service.NewMLPService(testCfg.MLPConfig.MLPURL,
		testCfg.MLPConfig.MLPEncryptionKey, testCfg.MLPConfig.MerlinURL)
	assert.NoError(t, err)
	experimentService, err := service.NewExperimentsService(testCfg.Experiment)
	assert.NoError(t, err)

	// Validate
	appCtx, err := NewAppContext(nil, testCfg, testAuthorizer, testVaultClient)
	assert.NoError(t, err)

	alertService, err := service.NewGitlabOpsAlertService(nil, *testCfg.AlertConfig)
	assert.NoError(t, err)

	ensemblingImageBuilder, err := imagebuilder.NewEnsemblerJobImageBuilder(
		nil,
		*testCfg.BatchEnsemblingConfig.ImageBuildingConfig,
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
	)

	assert.NoError(t, err)

	assert.Equal(t, &AppContext{
		Authorizer: testAuthorizer,
		DeploymentService: service.NewDeploymentService(
			testCfg,
			map[string]cluster.Controller{
				defaultEnvironment: nil,
			},
			ensemblerImageBuilder,
		),
		RoutersService:        service.NewRoutersService(nil, mlpSvc, testCfg.RouterDefaults.MonitoringURLFormat),
		EnsemblersService:     ensemblersService,
		EnsemblingJobService:  ensemblingJobService,
		RouterVersionsService: service.NewRouterVersionsService(nil, mlpSvc, testCfg.RouterDefaults.MonitoringURLFormat),
		EventService:          service.NewEventService(nil),
		RouterDefaults:        testCfg.RouterDefaults,
		CryptoService:         service.NewCryptoService(testCfg.TuringEncryptionKey),
		MLPService:            mlpService,
		ExperimentsService:    experimentService,
		PodLogService: service.NewPodLogService(
			map[string]cluster.Controller{
				defaultEnvironment: nil,
			},
		),
		AlertService: alertService,
		BatchRunners: []batchrunner.BatchJobRunner{batchEnsemblingJobRunner},
	}, appCtx)
}
