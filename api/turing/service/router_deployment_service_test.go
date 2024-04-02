package service

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	merlin "github.com/caraml-dev/merlin/client"
	mlp "github.com/caraml-dev/mlp/api/client"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/models"
	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"

	"github.com/caraml-dev/turing/api/turing/cluster/mocks"
	"github.com/caraml-dev/turing/api/turing/cluster/servicebuilder"
	mockImgBuilder "github.com/caraml-dev/turing/api/turing/imagebuilder/mocks"
	tu "github.com/caraml-dev/turing/api/turing/internal/testutils"
)

// mockClusterServiceBuilder implements the servicebuilder.ClusterServiceBuilder interface
type mockClusterServiceBuilder struct {
	rv *models.RouterVersion
}

func (msb *mockClusterServiceBuilder) NewRouterEndpoint(
	_ *models.RouterVersion,
	_ *mlp.Project,
	_ string,
) (*cluster.VirtualService, error) {
	return &cluster.VirtualService{
		Name:      "test-svc-turing-router",
		Namespace: "test-namespace",
		Labels: map[string]string{
			"key": "value",
		},
		Endpoint: "test-svc-router.models.example.com",
	}, nil
}

func (msb *mockClusterServiceBuilder) NewSecret(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	routerServiceAccountKey string,
	enricherServiceAccountKey string,
	ensemblerServiceAccountKey string,
	expEngineServiceAccountKey string,
) *cluster.Secret {
	return &cluster.Secret{
		Name:      fmt.Sprintf("%s-svc-acct-secret-%d", routerVersion.Router.Name, routerVersion.Version),
		Namespace: project.Name,
		Data: map[string]string{
			"SecretKeyNameRouter":    routerServiceAccountKey,
			"SecretKeyNameEnricher":  enricherServiceAccountKey,
			"SecretKeyNameEnsembler": ensemblerServiceAccountKey,
			"SecretKeyNameExpEngine": expEngineServiceAccountKey,
		},
	}
}

func (msb *mockClusterServiceBuilder) NewEnricherService(
	rv *models.RouterVersion,
	project *mlp.Project,
	_ string,
	queueProxyResourcePercentage int,
	userContainerCPULimitRequestFactor float64,
	userContainerMemoryLimitRequestFactor float64,
	initialScale *int,
) (*cluster.KnativeService, error) {
	if rv != msb.rv {
		return nil, errors.New("Unexpected router version data")
	}
	return &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-enricher-%d", rv.Router.Name, rv.Version),
			Namespace: project.Name,
		},
		QueueProxyResourcePercentage:          queueProxyResourcePercentage,
		UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
		UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
	}, nil
}

func (msb *mockClusterServiceBuilder) NewEnsemblerService(
	rv *models.RouterVersion,
	project *mlp.Project,
	_ string,
	queueProxyResourcePercentage int,
	userContainerCPULimitRequestFactor float64,
	userContainerMemoryLimitRequestFactor float64,
	initialScale *int,
) (*cluster.KnativeService, error) {
	if rv != msb.rv {
		return nil, errors.New("Unexpected router version data")
	}
	return &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-ensembler-%d", rv.Router.Name, rv.Version),
			Namespace: project.Name,
		},
		QueueProxyResourcePercentage:          queueProxyResourcePercentage,
		UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
		UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
	}, nil
}

func (msb *mockClusterServiceBuilder) NewRouterService(
	rv *models.RouterVersion,
	project *mlp.Project,
	envType string,
	_ string,
	expConfig json.RawMessage,
	routerDefaults *config.RouterDefaults,
	sentryEnabled bool,
	sentryDSN string,
	queueProxyResourcePercentage int,
	userContainerCPULimitRequestFactor float64,
	userContainerMemoryLimitRequestFactor float64,
	initialScale *int,
) (*cluster.KnativeService, error) {
	if rv != msb.rv {
		return nil, errors.New("Unexpected router version data")
	}
	return &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-router-%d", rv.Router.Name, rv.Version),
			Namespace: project.Name,
			Envs: []corev1.EnvVar{
				{Name: "JAEGER_EP", Value: routerDefaults.JaegerCollectorEndpoint},
				{Name: "FLUENTD_TAG", Value: routerDefaults.FluentdConfig.Tag},
				{Name: "ENVIRONMENT", Value: envType},
				{Name: "SENTRY_ENABLED", Value: strconv.FormatBool(sentryEnabled)},
				{Name: "SENTRY_DSN", Value: sentryDSN},
			},
			ConfigMap: &cluster.ConfigMap{
				Name: fmt.Sprintf("%s-fiber-config-%d", rv.Router.Name, rv.Version),
				Data: string(expConfig),
			},
		},
		QueueProxyResourcePercentage:          queueProxyResourcePercentage,
		UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
		UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
	}, nil
}

func (msb *mockClusterServiceBuilder) NewFluentdService(
	rv *models.RouterVersion,
	project *mlp.Project,
	_ string,
	_ *config.FluentdConfig,
) *cluster.KubernetesService {
	return &cluster.KubernetesService{
		BaseService: &cluster.BaseService{
			Name:                  fmt.Sprintf("%s-fluentd-logger-%d", rv.Router.Name, rv.Version),
			Namespace:             project.Name,
			PersistentVolumeClaim: &cluster.PersistentVolumeClaim{Name: "pvc"},
		},
	}
}

func (msb *mockClusterServiceBuilder) NewPodDisruptionBudget(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	componentType string,
	pdbConfig config.PodDisruptionBudgetConfig,
) *cluster.PodDisruptionBudget {
	return &cluster.PodDisruptionBudget{
		Name: fmt.Sprintf(
			"%s-%s",
			servicebuilder.GetComponentName(routerVersion, componentType),
			servicebuilder.ComponentTypes.PDB,
		),
		Namespace:                project.Name,
		MaxUnavailablePercentage: pdbConfig.MaxUnavailablePercentage,
		MinAvailablePercentage:   pdbConfig.MinAvailablePercentage,
		Selector: &apimetav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": fmt.Sprintf(
					"%s-0",
					servicebuilder.GetComponentName(routerVersion, componentType),
				),
			},
		},
	}
}

func (msb *mockClusterServiceBuilder) GetRouterServiceName(_ *models.RouterVersion) string {
	return "test-router-svc"
}

func TestDeployEndpoint(t *testing.T) {
	testEnv := "test-env"
	testNamespace := "test-namespace"
	envType := "staging"
	defaultMinAvailablePercentage := 20

	// Create test router version
	filePath := filepath.Join("..", "testdata", "cluster",
		"servicebuilder", "router_version_success.json")
	routerVersion := tu.GetRouterVersion(t, filePath)

	// Create mock controller
	controller := &mocks.Controller{}
	controller.On("DeployKnativeService", mock.Anything, mock.Anything).Return(nil)
	controller.On("GetKnativeServiceURL", mock.Anything, mock.Anything, mock.Anything).Return("test-endpoint")
	controller.On("DeployKubernetesService", mock.Anything, mock.Anything).Return(nil)
	controller.On("CreateNamespace", mock.Anything, mock.Anything).Return(nil)
	controller.On("ApplyConfigMap", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	controller.On("CreateSecret", mock.Anything, mock.Anything).Return(nil)
	controller.On("ApplyIstioVirtualService", mock.Anything, mock.Anything).Return(nil)
	controller.On("ApplyPodDisruptionBudget", mock.Anything, mock.Anything, mock.Anything).
		Return(&policyv1.PodDisruptionBudget{}, nil)

	// Create mock service builder
	svcBuilder := &mockClusterServiceBuilder{routerVersion}

	// Create test endpoint service with mock controller and service builder
	ds := &deploymentService{
		routerDefaults: &config.RouterDefaults{
			JaegerCollectorEndpoint: "jaeger-endpoint",
			FluentdConfig:           &config.FluentdConfig{Tag: "fluentd-tag"},
		},
		deploymentTimeout:         time.Second * 5,
		deploymentDeletionTimeout: time.Second * 5,
		environmentType:           envType,
		sentryEnabled:             true,
		sentryDSN:                 "test:dsn",
		knativeServiceConfig: &config.KnativeServiceDefaults{
			QueueProxyResourcePercentage:          20,
			UserContainerCPULimitRequestFactor:    1.75,
			UserContainerMemoryLimitRequestFactor: 1.75,
		},
		clusterControllers: map[string]cluster.Controller{
			testEnv: controller,
		},
		svcBuilder: svcBuilder,
		pdbConfig: config.PodDisruptionBudgetConfig{
			Enabled:                true,
			MinAvailablePercentage: &defaultMinAvailablePercentage,
		},
	}

	eventsCh := NewEventChannel()
	go func() {
		for {
			_, done := eventsCh.Read()
			if done {
				return
			}
		}
	}()
	defer eventsCh.Close()

	// Run test method and validate
	endpoint, err := ds.DeployRouterVersion(
		&mlp.Project{Name: testNamespace},
		&merlin.Environment{Name: testEnv},
		nil,
		routerVersion,
		"router-service-account-key",
		"enricher-service-account-key",
		"ensembler-service-account-key",
		"exp-engine-service-account-key",
		nil,
		nil,
		eventsCh,
	)

	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("http://%s-router.models.example.com", routerVersion.Router.Name), endpoint)
	controller.AssertCalled(t, "CreateNamespace", mock.Anything, testNamespace)
	controller.AssertCalled(t, "DeployKubernetesService", mock.Anything, &cluster.KubernetesService{
		BaseService: &cluster.BaseService{
			Name:                  fmt.Sprintf("%s-fluentd-logger-%d", routerVersion.Router.Name, routerVersion.Version),
			Namespace:             testNamespace,
			PersistentVolumeClaim: &cluster.PersistentVolumeClaim{Name: "pvc"},
		},
	})
	controller.AssertCalled(t, "CreateSecret", mock.Anything, &cluster.Secret{
		Name:      fmt.Sprintf("%s-svc-acct-secret-%d", routerVersion.Router.Name, routerVersion.Version),
		Namespace: testNamespace,
		Data: map[string]string{
			"SecretKeyNameRouter":    "router-service-account-key",
			"SecretKeyNameEnricher":  "enricher-service-account-key",
			"SecretKeyNameEnsembler": "ensembler-service-account-key",
			"SecretKeyNameExpEngine": "exp-engine-service-account-key",
		},
	})
	controller.AssertCalled(t, "DeployKnativeService", mock.Anything, &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-enricher-%d", routerVersion.Router.Name, routerVersion.Version),
			Namespace: testNamespace,
		},
		QueueProxyResourcePercentage:          20,
		UserContainerCPULimitRequestFactor:    1.75,
		UserContainerMemoryLimitRequestFactor: 1.75,
	})
	controller.AssertCalled(t, "DeployKnativeService", mock.Anything, &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-ensembler-%d", routerVersion.Router.Name, routerVersion.Version),
			Namespace: testNamespace,
		},
		QueueProxyResourcePercentage:          20,
		UserContainerCPULimitRequestFactor:    1.75,
		UserContainerMemoryLimitRequestFactor: 1.75,
	})
	controller.AssertCalled(t, "ApplyConfigMap", mock.Anything, testNamespace,
		&cluster.ConfigMap{Name: fmt.Sprintf("%s-fiber-config-%d", routerVersion.Router.Name, routerVersion.Version)})
	controller.AssertCalled(t, "DeployKnativeService", mock.Anything, &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-router-%d", routerVersion.Router.Name, routerVersion.Version),
			Namespace: testNamespace,
			Envs: []corev1.EnvVar{
				{Name: "JAEGER_EP", Value: ds.routerDefaults.JaegerCollectorEndpoint},
				{Name: "FLUENTD_TAG", Value: ds.routerDefaults.FluentdConfig.Tag},
				{Name: "ENVIRONMENT", Value: envType},
				{Name: "SENTRY_ENABLED", Value: "true"},
				{Name: "SENTRY_DSN", Value: "test:dsn"},
			},
			ConfigMap: &cluster.ConfigMap{
				Name: fmt.Sprintf("%s-fiber-config-%d",
					routerVersion.Router.Name,
					routerVersion.Version,
				),
			},
		},
		QueueProxyResourcePercentage:          20,
		UserContainerCPULimitRequestFactor:    1.75,
		UserContainerMemoryLimitRequestFactor: 1.75,
	})
	controller.AssertCalled(t, "CreateSecret", mock.Anything, &cluster.Secret{
		Name:      fmt.Sprintf("%s-svc-acct-secret-%d", routerVersion.Router.Name, routerVersion.Version),
		Namespace: testNamespace,
		Data: map[string]string{
			"SecretKeyNameRouter":    "router-service-account-key",
			"SecretKeyNameEnricher":  "enricher-service-account-key",
			"SecretKeyNameEnsembler": "ensembler-service-account-key",
			"SecretKeyNameExpEngine": "exp-engine-service-account-key",
		},
	})
	controller.AssertNumberOfCalls(t, "DeployKnativeService", 3)
	controller.AssertCalled(t, "GetKnativeServiceURL", mock.Anything, "test-router-svc", testNamespace)
	controller.AssertCalled(t, "ApplyIstioVirtualService", mock.Anything, &cluster.VirtualService{
		Name:      "test-svc-turing-router",
		Namespace: "test-namespace",
		Labels: map[string]string{
			"key": "value",
		},
		Endpoint: "test-svc-router.models.example.com",
	})
	controller.AssertCalled(t, "ApplyPodDisruptionBudget", mock.Anything, testNamespace, cluster.PodDisruptionBudget{
		Name:                   "test-svc-turing-router-1-pdb",
		Namespace:              testNamespace,
		MinAvailablePercentage: &defaultMinAvailablePercentage,
		Selector: &apimetav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "test-svc-turing-router-1-0",
			},
		},
	})
	controller.AssertCalled(t, "ApplyPodDisruptionBudget", mock.Anything, testNamespace, cluster.PodDisruptionBudget{
		Name:                   "test-svc-turing-ensembler-1-pdb",
		Namespace:              testNamespace,
		MinAvailablePercentage: &defaultMinAvailablePercentage,
		Selector: &apimetav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "test-svc-turing-ensembler-1-0",
			},
		},
	})
	controller.AssertNumberOfCalls(t, "ApplyPodDisruptionBudget", 3)

	// Verify endpoint for upi routers
	routerVersion.Protocol = routerConfig.UPI
	endpoint, err = ds.DeployRouterVersion(
		&mlp.Project{Name: testNamespace},
		&merlin.Environment{Name: testEnv},
		nil,
		routerVersion,
		"router-service-account-key",
		"enricher-service-account-key",
		"ensembler-service-account-key",
		"exp-engine-service-account-key",
		nil,
		nil,
		eventsCh,
	)
	assert.Equal(t, fmt.Sprintf("%s-router.models.example.com:80", routerVersion.Router.Name), endpoint)
	assert.NoError(t, err)
}

func TestDeleteEndpoint(t *testing.T) {
	testEnv := "test-env"
	testNs := "test-namespace"
	timeout := time.Second * 5
	defaultMinAvailablePercentage := 10

	// Create mock controller
	controller := &mocks.Controller{}
	controller.On("DeleteKnativeService", mock.Anything, mock.Anything,
		mock.Anything, false).Return(nil)
	controller.On("DeleteKubernetesStatefulSet", mock.Anything, mock.Anything,
		mock.Anything, false).Return(nil)
	controller.On("DeleteKubernetesService", mock.Anything, mock.Anything,
		mock.Anything, false).Return(nil)
	controller.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
	controller.On("DeleteConfigMap", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
	controller.On("DeletePVCs", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
	controller.On("DeletePodDisruptionBudget", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create test router version
	filePath := filepath.Join("..", "testdata", "cluster",
		"servicebuilder", "router_version_success.json")
	routerVersion := tu.GetRouterVersion(t, filePath)

	// Create mock service builder
	svcBuilder := &mockClusterServiceBuilder{routerVersion}

	// Create test endpoint service with mock controller and service builder
	ds := &deploymentService{
		routerDefaults: &config.RouterDefaults{
			JaegerCollectorEndpoint: "jaeger-endpoint",
			FluentdConfig:           &config.FluentdConfig{Tag: "fluentd-tag"},
		},
		deploymentTimeout:         timeout,
		deploymentDeletionTimeout: timeout,
		knativeServiceConfig: &config.KnativeServiceDefaults{
			QueueProxyResourcePercentage:          20,
			UserContainerCPULimitRequestFactor:    1.75,
			UserContainerMemoryLimitRequestFactor: 1.75,
		},
		clusterControllers: map[string]cluster.Controller{
			testEnv: controller,
		},
		svcBuilder: svcBuilder,
		pdbConfig: config.PodDisruptionBudgetConfig{
			Enabled:                true,
			MinAvailablePercentage: &defaultMinAvailablePercentage,
		},
	}

	eventsCh := NewEventChannel()
	go func() {
		for {
			_, done := eventsCh.Read()
			if done {
				return
			}
		}
	}()
	defer eventsCh.Close()

	// Run test method and validate
	err := ds.UndeployRouterVersion(
		&mlp.Project{Name: testNs},
		&merlin.Environment{Name: testEnv},
		routerVersion,
		eventsCh,
		false,
	)
	assert.NoError(t, err)
	controller.AssertCalled(t, "DeleteKubernetesService",
		mock.Anything, "test-svc-fluentd-logger-1", testNs, false)
	controller.AssertCalled(t, "DeleteConfigMap", mock.Anything, "test-svc-fiber-config-1", testNs, false)
	controller.AssertCalled(t, "DeleteKnativeService", mock.Anything, "test-svc-enricher-1", testNs, false)
	controller.AssertCalled(t, "DeleteKnativeService", mock.Anything, "test-svc-ensembler-1", testNs, false)
	controller.AssertCalled(t, "DeleteKnativeService", mock.Anything, "test-svc-router-1", testNs, false)
	controller.AssertCalled(t, "DeleteSecret", mock.Anything, "test-svc-svc-acct-secret-1", testNs, false)
	controller.AssertCalled(t, "DeletePVCs", mock.Anything, mock.Anything, testNs, false)
	controller.AssertCalled(t, "DeletePodDisruptionBudget", mock.Anything, testNs, mock.Anything)
	controller.AssertNumberOfCalls(t, "DeleteKnativeService", 3)
	controller.AssertNumberOfCalls(t, "DeletePodDisruptionBudget", 3)
}

func TestBuildEnsemblerServiceImage(t *testing.T) {
	ensembler := &models.PyFuncEnsembler{GenericEnsembler: &models.GenericEnsembler{Name: "test-ensembler"}}
	project := &mlp.Project{}
	id := models.ID(1)
	routerVersion := &models.RouterVersion{
		Ensembler: &models.Ensembler{
			PyfuncConfig: &models.EnsemblerPyfuncConfig{
				EnsemblerID: &id,
				ProjectID:   &id,
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 1,
					MaxReplica: 2,
					CPURequest: resource.Quantity{
						Format: "500M",
					},
					MemoryRequest: resource.Quantity{
						Format: "1G",
					},
				},
				AutoscalingPolicy: &models.AutoscalingPolicy{
					Metric: models.AutoscalingMetricConcurrency,
					Target: "10",
				},
				Timeout: "5s",
				Env: []*models.EnvVar{
					{
						Name:  "key",
						Value: "value",
					},
				},
			},
		},
	}

	eventsCh := NewEventChannel()
	go func() {
		for {
			_, done := eventsCh.Read()
			if done {
				return
			}
		}
	}()
	defer eventsCh.Close()

	// Set up mock services
	imageBuilder := &mockImgBuilder.ImageBuilder{}
	imageBuilder.On("BuildImage", mock.Anything).Return("test-image", nil)
	ds := &deploymentService{
		ensemblerServiceImageBuilder: imageBuilder,
	}

	// Call test function
	_ = ds.buildEnsemblerServiceImage(ensembler, project, routerVersion, eventsCh)

	// Test that the docker config is set correctly
	assert.Equal(t, routerVersion.Ensembler.DockerConfig, &models.EnsemblerDockerConfig{
		Image:             "test-image",
		ResourceRequest:   routerVersion.Ensembler.PyfuncConfig.ResourceRequest,
		AutoscalingPolicy: routerVersion.Ensembler.PyfuncConfig.AutoscalingPolicy,
		Timeout:           routerVersion.Ensembler.PyfuncConfig.Timeout,
		Endpoint:          "/ensemble",
		Port:              8083,
		Env:               routerVersion.Ensembler.PyfuncConfig.Env,
	})
}

func TestCreatePodDisruptionBudgets(t *testing.T) {
	twenty, eighty := 20, 80
	testRouterLabels := map[string]string{
		"app":          "test",
		"environment":  "",
		"orchestrator": "turing",
		"stream":       "",
		"team":         "",
	}

	tests := map[string]struct {
		rv        *models.RouterVersion
		pdbConfig config.PodDisruptionBudgetConfig
		expected  []*cluster.PodDisruptionBudget
	}{
		"bad pdb config": {
			rv: &models.RouterVersion{
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 5,
				},
			},
			pdbConfig: config.PodDisruptionBudgetConfig{
				Enabled: true,
			},
			expected: []*cluster.PodDisruptionBudget{},
		},
		"all pdbs | minAvailablePercentage": {
			rv: &models.RouterVersion{
				Router: &models.Router{
					Name: "test",
				},
				Version: 3,
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 5,
				},
				Enricher: &models.Enricher{
					ResourceRequest: &models.ResourceRequest{
						MinReplica: 3,
					},
				},
				Ensembler: &models.Ensembler{
					DockerConfig: &models.EnsemblerDockerConfig{
						ResourceRequest: &models.ResourceRequest{
							MinReplica: 2,
						},
					},
				},
				LogConfig: &models.LogConfig{
					ResultLoggerType: models.BigQueryLogger,
				},
			},
			pdbConfig: config.PodDisruptionBudgetConfig{
				Enabled:                true,
				MinAvailablePercentage: &twenty,
			},
			expected: []*cluster.PodDisruptionBudget{
				{
					Name:                   "test-turing-enricher-3-pdb",
					Namespace:              "ns",
					Labels:                 testRouterLabels,
					MinAvailablePercentage: &twenty,
					Selector: &apimetav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-turing-enricher-3-0",
						},
					},
				},
				{
					Name:                   "test-turing-ensembler-3-pdb",
					Namespace:              "ns",
					Labels:                 testRouterLabels,
					MinAvailablePercentage: &twenty,
					Selector: &apimetav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-turing-ensembler-3-0",
						},
					},
				},
				{
					Name:                   "test-turing-router-3-pdb",
					Namespace:              "ns",
					Labels:                 testRouterLabels,
					MinAvailablePercentage: &twenty,
					Selector: &apimetav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-turing-router-3-0",
						},
					},
				},
				{
					Name:                   "test-turing-fluentd-logger-3-pdb",
					Namespace:              "ns",
					Labels:                 testRouterLabels,
					MinAvailablePercentage: &twenty,
					Selector: &apimetav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-turing-fluentd-logger-3",
						},
					},
				},
			},
		},
		"all pdbs | maxUnavailablePercentage": {
			rv: &models.RouterVersion{
				Router: &models.Router{
					Name: "test",
				},
				Version: 3,
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 5,
				},
				Enricher: &models.Enricher{
					ResourceRequest: &models.ResourceRequest{
						MinReplica: 3,
					},
				},
				Ensembler: &models.Ensembler{
					DockerConfig: &models.EnsemblerDockerConfig{
						ResourceRequest: &models.ResourceRequest{
							MinReplica: 2,
						},
					},
				},
				LogConfig: &models.LogConfig{
					ResultLoggerType: models.BigQueryLogger,
				},
			},
			pdbConfig: config.PodDisruptionBudgetConfig{
				Enabled:                  true,
				MaxUnavailablePercentage: &eighty,
			},
			expected: []*cluster.PodDisruptionBudget{
				{
					Name:                     "test-turing-enricher-3-pdb",
					Namespace:                "ns",
					Labels:                   testRouterLabels,
					MaxUnavailablePercentage: &eighty,
					Selector: &apimetav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-turing-enricher-3-0",
						},
					},
				},
				{
					Name:                     "test-turing-ensembler-3-pdb",
					Namespace:                "ns",
					Labels:                   testRouterLabels,
					MaxUnavailablePercentage: &eighty,
					Selector: &apimetav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-turing-ensembler-3-0",
						},
					},
				},
				{
					Name:                     "test-turing-router-3-pdb",
					Namespace:                "ns",
					Labels:                   testRouterLabels,
					MaxUnavailablePercentage: &eighty,
					Selector: &apimetav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-turing-router-3-0",
						},
					},
				},
				{
					Name:                     "test-turing-fluentd-logger-3-pdb",
					Namespace:                "ns",
					Labels:                   testRouterLabels,
					MaxUnavailablePercentage: &eighty,
					Selector: &apimetav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-turing-fluentd-logger-3",
						},
					},
				},
			},
		},
		"pyfunc ensembler": {
			rv: &models.RouterVersion{
				Router: &models.Router{
					Name: "test",
				},
				Version: 3,
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 1,
				},
				Ensembler: &models.Ensembler{
					PyfuncConfig: &models.EnsemblerPyfuncConfig{
						ResourceRequest: &models.ResourceRequest{
							MinReplica: 10,
						},
					},
				},
				LogConfig: &models.LogConfig{
					ResultLoggerType: models.NopLogger,
				},
			},
			pdbConfig: config.PodDisruptionBudgetConfig{
				Enabled:                true,
				MinAvailablePercentage: &twenty,
			},
			expected: []*cluster.PodDisruptionBudget{
				{
					Name:                   "test-turing-ensembler-3-pdb",
					Namespace:              "ns",
					Labels:                 testRouterLabels,
					MinAvailablePercentage: &twenty,
					Selector: &apimetav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test-turing-ensembler-3-0",
						},
					},
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ds := &deploymentService{
				pdbConfig: tt.pdbConfig,
				svcBuilder: servicebuilder.NewClusterServiceBuilder(
					resource.MustParse("200m"),
					resource.MustParse("200Mi"),
					10,
					[]corev1.TopologySpreadConstraint{},
				),
			}
			pdbs := ds.createPodDisruptionBudgets(tt.rv, &mlp.Project{Name: "ns"})
			assert.Equal(t, tt.expected, pdbs)
		})
	}
}
