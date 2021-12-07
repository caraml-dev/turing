// +build unit

package request

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	assertgotest "gotest.tools/assert"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/gojek/turing/api/turing/config"
	tu "github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service/mocks"
	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/pkg/request"
)

var expEngineConfig = manager.TuringExperimentConfig{
	Client: manager.Client{
		ID:       "1",
		Username: "client",
		Passkey:  "dummy_passkey",
	},
	Experiments: []manager.Experiment{
		{
			ID:   "2",
			Name: "test-exp",
		},
	},
	Variables: manager.Variables{
		ClientVariables: []manager.Variable{
			{
				Name:     "app_version",
				Required: false,
			},
		},
		ExperimentVariables: map[string][]manager.Variable{
			"2": {
				{
					Name:     "customer",
					Required: true,
				},
			},
		},
		Config: []manager.VariableConfig{
			{
				Name:        "customer",
				Required:    true,
				Field:       "customer_id",
				FieldSource: request.HeaderFieldSource,
			},
			{
				Name:        "app_version",
				Required:    false,
				Field:       "test_field",
				FieldSource: request.HeaderFieldSource,
			},
		},
	},
}

var createOrUpdateRequest = CreateOrUpdateRouterRequest{
	Environment: "env",
	Name:        "router",
	Config: &RouterConfig{
		Routes: []*models.Route{
			{
				ID:       "default",
				Type:     "PROXY",
				Endpoint: "endpoint",
				Timeout:  "6s",
			},
		},
		DefaultRouteID: "default",
		ExperimentEngine: &ExperimentEngineConfig{
			Type:   "litmus",
			Config: &expEngineConfig,
		},
		ResourceRequest: &models.ResourceRequest{
			MinReplica: 0,
			MaxReplica: 5,
			CPURequest: resource.Quantity{
				Format: "500M",
			},
			MemoryRequest: resource.Quantity{
				Format: "1G",
			},
		},
		Timeout: "10s",
		LogConfig: &LogConfig{
			ResultLoggerType: "bigquery",
			BigQueryConfig: &BigQueryConfig{
				Table:                "project.dataset.table",
				ServiceAccountSecret: "service_account",
			},
		},
		Enricher: &EnricherEnsemblerConfig{
			Image: "lala",
			ResourceRequest: &models.ResourceRequest{
				MinReplica: 0,
				MaxReplica: 5,
				CPURequest: resource.Quantity{
					Format: "500M",
				},
				MemoryRequest: resource.Quantity{
					Format: "1G",
				},
			},
			Endpoint: "endpoint",
			Timeout:  "6s",
			Port:     8080,
			Env: []*models.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
			},
		},
		Ensembler: &models.Ensembler{
			Type: "docker",
			DockerConfig: &models.EnsemblerDockerConfig{
				Image: "nginx",
				ResourceRequest: &models.ResourceRequest{
					CPURequest:    resource.Quantity{Format: "500m"},
					MemoryRequest: resource.Quantity{Format: "1Gi"},
				},
				Timeout: "5s",
			},
		},
	},
}

func TestRequestBuildRouter(t *testing.T) {
	projectID := models.ID(1)
	expected := &models.Router{
		ProjectID:       projectID,
		EnvironmentName: "env",
		Name:            "router",
		Status:          "pending",
	}
	got := createOrUpdateRequest.BuildRouter(projectID)
	expected.Model = got.Model
	assert.Equal(t, *expected, *got)
}

func TestRequestBuildRouterVersionWithDefaults(t *testing.T) {
	defaults := config.RouterDefaults{
		Image:                   "routerimage",
		FiberDebugLogEnabled:    true,
		CustomMetricsEnabled:    true,
		JaegerEnabled:           true,
		JaegerCollectorEndpoint: "jaegerendpoint",
		LogLevel:                "DEBUG",
		FluentdConfig: &config.FluentdConfig{
			Image: "fluentdimage",
			Tag:   "fluentdtag",
		},
		Experiment: map[string]interface{}{
			"litmus": map[string]interface{}{
				"endpoint": "grpc://test",
				"timeout":  "2s",
			},
		},
	}
	projectID := models.ID(1)
	router := createOrUpdateRequest.BuildRouter(projectID)
	expected := models.RouterVersion{
		Router: router,
		Status: "pending",
		Image:  "routerimage",
		Routes: []*models.Route{
			{
				ID:       "default",
				Type:     "PROXY",
				Endpoint: "endpoint",
				Timeout:  "6s",
			},
		},
		DefaultRouteID: "default",
		ExperimentEngine: &models.ExperimentEngine{
			Type: "litmus",
			Config: &manager.TuringExperimentConfig{
				Client: manager.Client{
					ID:       "1",
					Username: "client",
					Passkey:  "enc_passkey",
				},
				Experiments: []manager.Experiment{
					{
						ID:   "2",
						Name: "test-exp",
					},
				},
				Variables: manager.Variables{
					ClientVariables: []manager.Variable{
						{
							Name:     "app_version",
							Required: false,
						},
					},
					ExperimentVariables: map[string][]manager.Variable{
						"2": {
							{
								Name:     "customer",
								Required: true,
							},
						},
					},
					Config: []manager.VariableConfig{
						{
							Name:        "customer",
							Required:    true,
							Field:       "customer_id",
							FieldSource: request.HeaderFieldSource,
						},
						{
							Name:        "app_version",
							Required:    false,
							Field:       "test_field",
							FieldSource: request.HeaderFieldSource,
						},
					},
				},
			},
		},
		ResourceRequest: &models.ResourceRequest{
			MinReplica: 0,
			MaxReplica: 5,
			CPURequest: resource.Quantity{
				Format: "500M",
			},
			MemoryRequest: resource.Quantity{
				Format: "1G",
			},
		},
		Timeout: "10s",
		LogConfig: &models.LogConfig{
			LogLevel:             "DEBUG",
			CustomMetricsEnabled: true,
			FiberDebugLogEnabled: true,
			JaegerEnabled:        true,
			ResultLoggerType:     models.BigQueryLogger,
			BigQueryConfig: &models.BigQueryConfig{
				Table:                "project.dataset.table",
				ServiceAccountSecret: "service_account",
				BatchLoad:            true,
			},
		},
		Enricher: &models.Enricher{
			Image: "lala",
			ResourceRequest: &models.ResourceRequest{
				MinReplica: 0,
				MaxReplica: 5,
				CPURequest: resource.Quantity{
					Format: "500M",
				},
				MemoryRequest: resource.Quantity{
					Format: "1G",
				},
			},
			Endpoint: "endpoint",
			Timeout:  "6s",
			Port:     8080,
			Env: []*models.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
			},
		},
		Ensembler: &models.Ensembler{
			Type: "docker",
			DockerConfig: &models.EnsemblerDockerConfig{
				Image: "nginx",
				ResourceRequest: &models.ResourceRequest{
					CPURequest:    resource.Quantity{Format: "500m"},
					MemoryRequest: resource.Quantity{Format: "1Gi"},
				},
				Timeout: "5s",
			},
		},
	}

	// Set up mock Crypto service
	cryptoSvc := &mocks.CryptoService{}
	cryptoSvc.On("Encrypt", "dummy_passkey").Return("enc_passkey", nil)

	// Set up mock Experiment service
	expSvc := &mocks.ExperimentsService{}
	expSvc.On("IsStandardExperimentManager", mock.Anything).Return(true)
	expSvc.On("GetStandardExperimentConfig", &expEngineConfig).Return(expEngineConfig, nil)

	got, err := createOrUpdateRequest.BuildRouterVersion(router, &defaults, cryptoSvc, expSvc)
	tu.FailOnError(t, err)
	expected.Model = got.Model
	assertgotest.DeepEqual(t, expected, *got)
}

func TestBuildExperimentEngineConfig(t *testing.T) {
	// Set up mock Crypto service
	cs := &mocks.CryptoService{}
	cs.On("Encrypt", "xp-passkey-bad").Return("", errors.New("test-encrypt-error"))
	cs.On("Encrypt", "xp-passkey").Return("xp-passkey-enc", nil)

	// Set up mock Experiment service
	cfgWithPasskey := &manager.TuringExperimentConfig{
		Client: manager.Client{
			Username: "client-name",
			Passkey:  "xp-passkey",
		},
	}
	cfgWithoutPasskey := &manager.TuringExperimentConfig{
		Client: manager.Client{
			Username: "client-name",
		},
	}
	cfgWithBadPasskey := &manager.TuringExperimentConfig{
		Client: manager.Client{
			Username: "client-name",
			Passkey:  "xp-passkey-bad",
		},
	}
	es := &mocks.ExperimentsService{}
	es.On("IsStandardExperimentManager", "litmus").Return(true)
	es.On("IsStandardExperimentManager", "xp").Return(false)
	es.On("GetStandardExperimentConfig", cfgWithPasskey).Return(*cfgWithPasskey, nil)
	es.On("GetStandardExperimentConfig", cfgWithoutPasskey).Return(*cfgWithoutPasskey, nil)
	es.On("GetStandardExperimentConfig", cfgWithBadPasskey).Return(*cfgWithBadPasskey, nil)

	// Define tests
	tests := map[string]struct {
		req      CreateOrUpdateRouterRequest
		router   *models.Router
		expected interface{}
		err      string
	}{
		"failure | std engine | missing curr version passkey": {
			req: CreateOrUpdateRouterRequest{
				Config: &RouterConfig{
					ExperimentEngine: &ExperimentEngineConfig{
						Type:   "litmus",
						Config: cfgWithoutPasskey,
					},
				},
			},
			router: &models.Router{},
			err:    "Passkey must be configured",
		},
		"failure | std engine | encrypt passkey error": {
			req: CreateOrUpdateRouterRequest{
				Config: &RouterConfig{
					ExperimentEngine: &ExperimentEngineConfig{
						Type:   "litmus",
						Config: cfgWithBadPasskey,
					},
				},
			},
			router: &models.Router{},
			err:    "Passkey could not be encrypted: test-encrypt-error",
		},
		"success | std engine | use curr version passkey": {
			req: CreateOrUpdateRouterRequest{
				Config: &RouterConfig{
					ExperimentEngine: &ExperimentEngineConfig{
						Type:   "litmus",
						Config: cfgWithoutPasskey,
					},
				},
			},
			router: &models.Router{
				CurrRouterVersion: &models.RouterVersion{
					ExperimentEngine: &models.ExperimentEngine{
						Type:   "litmus",
						Config: cfgWithPasskey,
					},
				},
			},
			expected: cfgWithPasskey,
		},
		"success | std engine | use new passkey": {
			req: CreateOrUpdateRouterRequest{
				Config: &RouterConfig{
					ExperimentEngine: &ExperimentEngineConfig{
						Type:   "litmus",
						Config: cfgWithPasskey,
					},
				},
			},
			router: &models.Router{},
			expected: &manager.TuringExperimentConfig{
				Client: manager.Client{
					Username: "client-name",
					Passkey:  "xp-passkey-enc",
				},
			},
		},
		"success | custom engine": {
			req: CreateOrUpdateRouterRequest{
				Config: &RouterConfig{
					ExperimentEngine: &ExperimentEngineConfig{
						Type:   "xp",
						Config: []int{1, 2},
					},
				},
			},
			router:   &models.Router{},
			expected: []int{1, 2},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := data.req.BuildExperimentEngineConfig(data.router, cs, es)
			assert.Equal(t, data.expected, result)
			assert.Equal(t, data.err == "", err == nil)
			if err != nil {
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}
