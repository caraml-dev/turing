package request

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	assertgotest "gotest.tools/assert"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service/mocks"
	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/pkg/request"
)

func makeTuringExperimentConfig(clientPasskey string) json.RawMessage {
	expEngineConfig, _ := json.Marshal(manager.TuringExperimentConfig{
		Client: manager.Client{
			ID:       "1",
			Username: "client",
			Passkey:  clientPasskey,
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
	})
	return expEngineConfig
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
			Type:   "standard",
			Config: makeTuringExperimentConfig("dummy_passkey"),
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
				EnsemblerContainerRuntimeConfig: &models.EnsemblerContainerRuntimeConfig{
					ResourceRequest: &models.ResourceRequest{
						CPURequest:    resource.Quantity{Format: "500m"},
						MemoryRequest: resource.Quantity{Format: "1Gi"},
					},
					Timeout: "5s",
				},
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

func TestRequestBuildRouterVersionLoggerConfiguration(t *testing.T) {

	baseRequest := CreateOrUpdateRouterRequest{
		Environment: "env",
		Name:        "router",
		Config: &RouterConfig{
			ExperimentEngine: &ExperimentEngineConfig{
				Type: "nop",
			},
		},
	}

	projectID := models.ID(1)
	routerDefault := config.RouterDefaults{
		KafkaConfig: &config.KafkaConfig{
			MaxMessageBytes: 1110,
			CompressionType: "gzip",
		},
	}

	tests := []struct {
		testName          string
		logConfig         *LogConfig
		expectedLogConfig *models.LogConfig
	}{
		{
			testName: "Test Kafka Logger",
			logConfig: &LogConfig{
				ResultLoggerType: "kafka",
				KafkaConfig: &KafkaConfig{
					Brokers:             "10:11",
					Topic:               "2222",
					SerializationFormat: "json",
				},
			},
			expectedLogConfig: &models.LogConfig{
				LogLevel:             "",
				CustomMetricsEnabled: false,
				FiberDebugLogEnabled: false,
				JaegerEnabled:        false,
				ResultLoggerType:     "kafka",
				KafkaConfig: &models.KafkaConfig{
					Brokers:             "10:11",
					Topic:               "2222",
					SerializationFormat: "json",
					MaxMessageBytes:     1110,
					CompressionType:     "gzip",
				},
				BigQueryConfig: nil,
			},
		},
		{
			testName: "Test BQ Logger",
			logConfig: &LogConfig{
				ResultLoggerType: "bigquery",
				BigQueryConfig: &BigQueryConfig{
					Table:                "project.dataset.table",
					ServiceAccountSecret: "service_account",
				},
			},
			expectedLogConfig: &models.LogConfig{
				LogLevel:             "",
				CustomMetricsEnabled: false,
				FiberDebugLogEnabled: false,
				JaegerEnabled:        false,
				ResultLoggerType:     "bigquery",
				KafkaConfig:          nil,
				BigQueryConfig: &models.BigQueryConfig{
					Table:                "project.dataset.table",
					ServiceAccountSecret: "service_account",
					BatchLoad:            true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			baseRequest.Config.LogConfig = tt.logConfig
			router := baseRequest.BuildRouter(projectID)
			// Set up mock Crypto service
			cryptoSvc := &mocks.CryptoService{}
			cryptoSvc.On("Encrypt", "dummy_passkey").Return("enc_passkey", nil)

			// Set up mock Experiment service
			expSvc := &mocks.ExperimentsService{}
			expSvc.On("IsStandardExperimentManager", mock.Anything).Return(true)

			got, err := baseRequest.BuildRouterVersion(router, &routerDefault, cryptoSvc, expSvc)
			assert.NoError(t, err)
			assert.Equal(t, got.LogConfig, tt.expectedLogConfig)
		})
	}
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
			Type:   "standard",
			Config: makeTuringExperimentConfig("enc_passkey"),
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
				EnsemblerContainerRuntimeConfig: &models.EnsemblerContainerRuntimeConfig{
					ResourceRequest: &models.ResourceRequest{
						CPURequest:    resource.Quantity{Format: "500m"},
						MemoryRequest: resource.Quantity{Format: "1Gi"},
					},
					Timeout: "5s",
				},
			},
		},
	}

	// Set up mock Crypto service
	cryptoSvc := &mocks.CryptoService{}
	cryptoSvc.On("Encrypt", "dummy_passkey").Return("enc_passkey", nil)

	// Set up mock Experiment service
	expSvc := &mocks.ExperimentsService{}
	expSvc.On("IsStandardExperimentManager", mock.Anything).Return(true)

	got, err := createOrUpdateRequest.BuildRouterVersion(router, &defaults, cryptoSvc, expSvc)
	require.NoError(t, err)
	expected.Model = got.Model
	assertgotest.DeepEqual(t, expected, *got)
}

func TestBuildExperimentEngineConfig(t *testing.T) {
	// Set up mock Crypto service
	cs := &mocks.CryptoService{}
	cs.On("Encrypt", "passkey-bad").
		Return("", errors.New("test-encrypt-error"))
	cs.On("Encrypt", "passkey").
		Return("passkey-enc", nil)

	es := &mocks.ExperimentsService{}
	es.On("IsStandardExperimentManager", "standard-manager").Return(true)
	es.On("IsStandardExperimentManager", "custom-manager").Return(false)

	// Define tests
	tests := map[string]struct {
		req      CreateOrUpdateRouterRequest
		router   *models.Router
		expected json.RawMessage
		err      string
	}{
		"failure | std engine | missing curr version passkey": {
			req: CreateOrUpdateRouterRequest{
				Config: &RouterConfig{
					ExperimentEngine: &ExperimentEngineConfig{
						Type:   "standard-manager",
						Config: json.RawMessage(`{"client": {"username": "client-name"}}`),
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
						Type:   "standard-manager",
						Config: json.RawMessage(`{"client": {"username": "client-name", "passkey": "passkey-bad"}}`),
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
						Type:   "standard-manager",
						Config: json.RawMessage(`{"client": {"username": "client-name"}}`),
					},
				},
			},
			router: &models.Router{
				CurrRouterVersion: &models.RouterVersion{
					ExperimentEngine: &models.ExperimentEngine{
						Type:   "standard-manager",
						Config: json.RawMessage(`{"client": {"username": "client-name", "passkey": "passkey"}}`),
					},
				},
			},
			expected: json.RawMessage(`{
				"client":{"id":"","username":"client-name","passkey":"passkey"},
				"experiments":null,
				"variables":{"client_variables":null,"experiment_variables":null,"config":null}
			}`),
		},
		"success | std engine | use new passkey": {
			req: CreateOrUpdateRouterRequest{
				Config: &RouterConfig{
					ExperimentEngine: &ExperimentEngineConfig{
						Type:   "standard-manager",
						Config: json.RawMessage(`{"client": {"username": "client-name", "passkey": "passkey"}}`),
					},
				},
			},
			router: &models.Router{},
			expected: json.RawMessage(`{
				"client":{"id":"","username":"client-name","passkey":"passkey-enc"},
				"experiments":null,
				"variables":{"client_variables":null,"experiment_variables":null,"config":null}
			}`),
		},
		"success | custom engine": {
			req: CreateOrUpdateRouterRequest{
				Config: &RouterConfig{
					ExperimentEngine: &ExperimentEngineConfig{
						Type:   "custom-manager",
						Config: json.RawMessage("[1, 2]"),
					},
				},
			},
			router:   &models.Router{},
			expected: json.RawMessage("[1, 2]"),
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := data.req.BuildExperimentEngineConfig(data.router, cs, es)
			if data.err == "" {
				assert.JSONEq(t, string(data.expected), string(result))
			} else {
				assert.EqualError(t, err, data.err)
			}
		})
	}
}
