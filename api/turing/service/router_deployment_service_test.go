// +build unit

package service

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/cluster/mocks"
	corev1 "k8s.io/api/core/v1"

	"github.com/gojek/turing/api/turing/config"

	merlin "github.com/gojek/merlin/client"
	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/models"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	tu "github.com/gojek/turing/api/turing/internal/testutils"
)

// mockClusterServiceBuilder implements the servicebuilder.ClusterServiceBuilder interface
type mockClusterServiceBuilder struct {
	rv *models.RouterVersion
}

func (msb *mockClusterServiceBuilder) NewRouterEndpoint(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	envType string,
	versionEndpoint string,
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
	experimentPasskey string,
) *cluster.Secret {
	return &cluster.Secret{
		Name:      fmt.Sprintf("%s-svc-acct-secret-%d", routerVersion.Router.Name, routerVersion.Version),
		Namespace: project.Name,
		Data: map[string]string{
			"SecretKeyNameRouter":     routerServiceAccountKey,
			"SecretKeyNameEnricher":   enricherServiceAccountKey,
			"SecretKeyNameEnsembler":  ensemblerServiceAccountKey,
			"SecretKeyNameExperiment": experimentPasskey,
		},
	}
}

func (msb *mockClusterServiceBuilder) NewEnricherService(
	rv *models.RouterVersion,
	project *mlp.Project,
	envType string,
	secretName string,
	targetConcurrency int,
	queueProxyResourcePercentage int,
	userContainerLimitRequestFactor float64,
) (*cluster.KnativeService, error) {
	if rv != msb.rv {
		return nil, errors.New("Unexpected router version data")
	}
	return &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-enricher-%d", rv.Router.Name, rv.Version),
			Namespace: project.Name,
			Labels: map[string]string{
				"env": envType,
			},
		},
		TargetConcurrency:               targetConcurrency,
		QueueProxyResourcePercentage:    queueProxyResourcePercentage,
		UserContainerLimitRequestFactor: userContainerLimitRequestFactor,
	}, nil
}

func (msb *mockClusterServiceBuilder) NewEnsemblerService(
	rv *models.RouterVersion,
	project *mlp.Project,
	envType string,
	secretName string,
	targetConcurrency int,
	queueProxyResourcePercentage int,
	userContainerLimitRequestFactor float64,
) (*cluster.KnativeService, error) {
	if rv != msb.rv {
		return nil, errors.New("Unexpected router version data")
	}
	return &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-ensembler-%d", rv.Router.Name, rv.Version),
			Namespace: project.Name,
			Labels: map[string]string{
				"env": envType,
			},
		},
		TargetConcurrency:               targetConcurrency,
		QueueProxyResourcePercentage:    queueProxyResourcePercentage,
		UserContainerLimitRequestFactor: userContainerLimitRequestFactor,
	}, nil
}

func (msb *mockClusterServiceBuilder) NewRouterService(
	rv *models.RouterVersion,
	project *mlp.Project,
	envType string,
	secretName string,
	expConfig json.RawMessage,
	fluentTag string,
	jaegerEndpoint string,
	sentryEnabled bool,
	sentryDSN string,
	targetConcurrency int,
	queueProxyResourcePercentage int,
	userContainerLimitRequestFactor float64,
) (*cluster.KnativeService, error) {
	if rv != msb.rv {
		return nil, errors.New("Unexpected router version data")
	}
	return &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-router-%d", rv.Router.Name, rv.Version),
			Namespace: project.Name,
			Envs: []corev1.EnvVar{
				{Name: "JAEGER_EP", Value: jaegerEndpoint},
				{Name: "FLUENTD_TAG", Value: fluentTag},
				{Name: "ENVIRONMENT", Value: envType},
				{Name: "SENTRY_ENABLED", Value: strconv.FormatBool(sentryEnabled)},
				{Name: "SENTRY_DSN", Value: sentryDSN},
			},
			Labels: map[string]string{
				"env": envType,
			},
			ConfigMap: &cluster.ConfigMap{
				Name: fmt.Sprintf("%s-fiber-config-%d", rv.Router.Name, rv.Version),
				Data: string(expConfig),
			},
		},
		TargetConcurrency:               targetConcurrency,
		QueueProxyResourcePercentage:    queueProxyResourcePercentage,
		UserContainerLimitRequestFactor: userContainerLimitRequestFactor,
	}, nil
}

func (msb *mockClusterServiceBuilder) NewPluginsServerService(
	rv *models.RouterVersion,
	project *mlp.Project,
	envType string,
) *cluster.KubernetesService {
	return nil
}

func (msb *mockClusterServiceBuilder) NewFluentdService(
	rv *models.RouterVersion,
	project *mlp.Project,
	envType string,
	serviceAccountSecretName string,
	cfg *config.FluentdConfig,
) *cluster.KubernetesService {
	return &cluster.KubernetesService{
		BaseService: &cluster.BaseService{
			Name:                  fmt.Sprintf("%s-fluentd-logger-%d", rv.Router.Name, rv.Version),
			Namespace:             project.Name,
			PersistentVolumeClaim: &cluster.PersistentVolumeClaim{Name: "pvc"},
		},
	}
}

func (msb *mockClusterServiceBuilder) GetRouterServiceName(ver *models.RouterVersion) string {
	return "test-router-svc"
}

func TestDeployEndpoint(t *testing.T) {
	testEnv := "test-env"
	testNamespace := "test-namespace"
	envType := "staging"

	// Create test router version
	filePath := filepath.Join("..", "testdata", "cluster",
		"servicebuilder", "router_version_success.json")
	routerVersion := tu.GetRouterVersion(t, filePath)

	// Create mock controller
	controller := &mocks.Controller{}
	controller.On("DeployKnativeService", mock.Anything, mock.Anything).Return(nil)
	controller.On("GetKnativeServiceURL", mock.Anything, mock.Anything).Return("test-endpoint")
	controller.On("DeployKubernetesService", mock.Anything, mock.Anything).Return(nil)
	controller.On("CreateNamespace", mock.Anything).Return(nil)
	controller.On("ApplyConfigMap", mock.Anything, mock.Anything).Return(nil)
	controller.On("CreateSecret", mock.Anything, mock.Anything).Return(nil)
	controller.On("ApplyPersistentVolumeClaim", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	controller.On("ApplyIstioVirtualService", mock.Anything, mock.Anything).Return(nil)

	// Create mock service builder
	svcBuilder := &mockClusterServiceBuilder{routerVersion}

	// Create test endpoint service with mock controller and service builder
	ds := &deploymentService{
		fluentdConfig: &config.FluentdConfig{
			Tag: "fluentd-tag",
		},
		jaegerCollectorEndpoint:   "jaeger-endpoint",
		deploymentTimeout:         time.Second * 5,
		deploymentDeletionTimeout: time.Second * 5,
		environmentType:           envType,
		sentryEnabled:             true,
		sentryDSN:                 "test:dsn",
		knativeServiceConfig: &config.KnativeServiceDefaults{
			TargetConcurrency:               1,
			QueueProxyResourcePercentage:    20,
			UserContainerLimitRequestFactor: 1.75,
		},
		clusterControllers: map[string]cluster.Controller{
			testEnv: controller,
		},
		svcBuilder: svcBuilder,
	}

	eventsCh := models.NewEventChannel()
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
		routerVersion,
		"router-service-account-key",
		"enricher-service-account-key",
		"ensembler-service-account-key",
		nil,
		"experiment-passkey",
		eventsCh,
	)

	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("http://%s-router.models.example.com", routerVersion.Router.Name), endpoint)
	controller.AssertCalled(t, "CreateNamespace", testNamespace)
	controller.AssertCalled(t, "ApplyPersistentVolumeClaim", mock.Anything,
		testNamespace, &cluster.PersistentVolumeClaim{Name: "pvc"})
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
			"SecretKeyNameRouter":     "router-service-account-key",
			"SecretKeyNameEnricher":   "enricher-service-account-key",
			"SecretKeyNameEnsembler":  "ensembler-service-account-key",
			"SecretKeyNameExperiment": "experiment-passkey",
		},
	})
	controller.AssertCalled(t, "DeployKnativeService", mock.Anything, &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-enricher-%d", routerVersion.Router.Name, routerVersion.Version),
			Namespace: testNamespace,
			Labels: map[string]string{
				"env": envType,
			},
		},
		TargetConcurrency:               1,
		QueueProxyResourcePercentage:    20,
		UserContainerLimitRequestFactor: 1.75,
	})
	controller.AssertCalled(t, "DeployKnativeService", mock.Anything, &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-ensembler-%d", routerVersion.Router.Name, routerVersion.Version),
			Namespace: testNamespace,
			Labels: map[string]string{
				"env": envType,
			},
		},
		TargetConcurrency:               1,
		QueueProxyResourcePercentage:    20,
		UserContainerLimitRequestFactor: 1.75,
	})
	controller.AssertCalled(t, "ApplyConfigMap", testNamespace,
		&cluster.ConfigMap{Name: fmt.Sprintf("%s-fiber-config-%d", routerVersion.Router.Name, routerVersion.Version)})
	controller.AssertCalled(t, "DeployKnativeService", mock.Anything, &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:      fmt.Sprintf("%s-router-%d", routerVersion.Router.Name, routerVersion.Version),
			Namespace: testNamespace,
			Envs: []corev1.EnvVar{
				{Name: "JAEGER_EP", Value: ds.jaegerCollectorEndpoint},
				{Name: "FLUENTD_TAG", Value: ds.fluentdConfig.Tag},
				{Name: "ENVIRONMENT", Value: envType},
				{Name: "SENTRY_ENABLED", Value: "true"},
				{Name: "SENTRY_DSN", Value: "test:dsn"},
			},
			Labels: map[string]string{
				"env": envType,
			},
			ConfigMap: &cluster.ConfigMap{
				Name: fmt.Sprintf("%s-fiber-config-%d",
					routerVersion.Router.Name,
					routerVersion.Version,
				),
			},
		},
		TargetConcurrency:               1,
		QueueProxyResourcePercentage:    20,
		UserContainerLimitRequestFactor: 1.75,
	})
	controller.AssertCalled(t, "CreateSecret", mock.Anything, &cluster.Secret{
		Name:      fmt.Sprintf("%s-svc-acct-secret-%d", routerVersion.Router.Name, routerVersion.Version),
		Namespace: testNamespace,
		Data: map[string]string{
			"SecretKeyNameRouter":     "router-service-account-key",
			"SecretKeyNameEnricher":   "enricher-service-account-key",
			"SecretKeyNameEnsembler":  "ensembler-service-account-key",
			"SecretKeyNameExperiment": "experiment-passkey",
		},
	})
	controller.AssertNumberOfCalls(t, "DeployKnativeService", 3)
	controller.AssertCalled(t, "GetKnativeServiceURL", "test-router-svc", testNamespace)
	controller.AssertCalled(t, "ApplyIstioVirtualService", mock.Anything, &cluster.VirtualService{
		Name:      "test-svc-turing-router",
		Namespace: "test-namespace",
		Labels: map[string]string{
			"key": "value",
		},
		Endpoint: "test-svc-router.models.example.com",
	})
}

func TestDeleteEndpoint(t *testing.T) {
	testEnv := "test-env"
	testNs := "test-namespace"
	timeout := time.Second * 5

	// Create mock controller
	controller := &mocks.Controller{}
	controller.On("DeleteKnativeService", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(nil)
	controller.On("DeleteKubernetesService", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(nil)
	controller.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	controller.On("DeleteConfigMap", mock.Anything, mock.Anything).Return(nil)
	controller.On("DeletePersistentVolumeClaim", mock.Anything, mock.Anything).Return(nil)

	// Create test router version
	filePath := filepath.Join("..", "testdata", "cluster",
		"servicebuilder", "router_version_success.json")
	routerVersion := tu.GetRouterVersion(t, filePath)

	// Create mock service builder
	svcBuilder := &mockClusterServiceBuilder{routerVersion}

	// Create test endpoint service with mock controller and service builder
	ds := &deploymentService{
		fluentdConfig: &config.FluentdConfig{
			Tag: "fluentd-tag",
		},
		jaegerCollectorEndpoint:   "jaeger-endpoint",
		deploymentTimeout:         timeout,
		deploymentDeletionTimeout: timeout,
		knativeServiceConfig: &config.KnativeServiceDefaults{
			TargetConcurrency:               1,
			QueueProxyResourcePercentage:    20,
			UserContainerLimitRequestFactor: 1.75,
		},
		clusterControllers: map[string]cluster.Controller{
			testEnv: controller,
		},
		svcBuilder: svcBuilder,
	}

	eventsCh := models.NewEventChannel()
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
	)
	assert.NoError(t, err)
	controller.AssertCalled(t, "DeleteKubernetesService", "test-svc-fluentd-logger-1", testNs, timeout)
	controller.AssertCalled(t, "DeleteConfigMap", "test-svc-fiber-config-1", testNs)
	controller.AssertCalled(t, "DeleteKnativeService", "test-svc-enricher-1", testNs, timeout)
	controller.AssertCalled(t, "DeleteKnativeService", "test-svc-ensembler-1", testNs, timeout)
	controller.AssertCalled(t, "DeleteKnativeService", "test-svc-router-1", testNs, timeout)
	controller.AssertCalled(t, "DeleteSecret", "test-svc-svc-acct-secret-1", testNs)
	controller.AssertCalled(t, "DeletePersistentVolumeClaim", "pvc", testNs)
	controller.AssertNumberOfCalls(t, "DeleteKnativeService", 3)
}
