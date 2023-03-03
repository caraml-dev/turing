package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	assertgotest "gotest.tools/assert"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service/mocks"
	"github.com/caraml-dev/turing/engines/experiment/manager"
	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
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

var defaultRouteID = "default"
var validRouterConfig = RouterConfig{
	Routes: []*models.Route{
		{
			ID:       "default",
			Type:     "PROXY",
			Endpoint: "endpoint",
			Timeout:  "6s",
		},
	},
	DefaultRouteID: &defaultRouteID,
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
	AutoscalingPolicy: &models.AutoscalingPolicy{
		Metric: models.AutoscalingMetricCPU,
		Target: "80",
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
		AutoscalingPolicy: &models.AutoscalingPolicy{
			Metric: models.AutoscalingMetricRPS,
			Target: "100",
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
			AutoscalingPolicy: &models.AutoscalingPolicy{
				Metric: models.AutoscalingMetricRPS,
				Target: "200",
			},
			Timeout: "5s",
		},
	},
}

var invalidRouterConfig = RouterConfig{
	Routes: []*models.Route{
		{
			ID:       "default",
			Type:     "PROXY",
			Endpoint: "endpoint",
			Timeout:  "6s",
		},
	},
	DefaultRouteID: &defaultRouteID,
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
		Type: "pyfunc",
		PyfuncConfig: &models.EnsemblerPyfuncConfig{
			ProjectID:   models.NewID(11),
			EnsemblerID: models.NewID(12),
			ResourceRequest: &models.ResourceRequest{
				CPURequest:    resource.Quantity{Format: "500m"},
				MemoryRequest: resource.Quantity{Format: "1Gi"},
			},
			Timeout: "5s",
		},
	},
}

var createOrUpdateRequest = CreateOrUpdateRouterRequest{
	Environment: "env",
	Name:        "router",
	Config:      &validRouterConfig,
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

	projectName := "test-project"
	projectID := models.ID(1)
	routerDefault := config.RouterDefaults{
		KafkaConfig: &config.KafkaConfig{
			MaxMessageBytes: 1110,
			CompressionType: "gzip",
		},
		UPIConfig: &config.UPIConfig{KafkaBrokers: "broker"},
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
				},
				BigQueryConfig: nil,
			},
		},
		{
			testName: "Test UPI Logger",
			logConfig: &LogConfig{
				ResultLoggerType: "upi",
			},
			expectedLogConfig: &models.LogConfig{
				LogLevel:             "",
				CustomMetricsEnabled: false,
				FiberDebugLogEnabled: false,
				JaegerEnabled:        false,
				ResultLoggerType:     "upi",
				KafkaConfig: &models.KafkaConfig{
					Brokers:             routerDefault.UPIConfig.KafkaBrokers,
					Topic:               fmt.Sprintf("caraml-%s-%s-router-log", projectName, baseRequest.Name),
					SerializationFormat: "protobuf",
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

			// Set up mock Ensembler service
			ensemblerSvc := &mocks.EnsemblersService{}

			got, err := baseRequest.Config.BuildRouterVersion(
				projectName, router, &routerDefault, cryptoSvc, expSvc, ensemblerSvc)
			assert.NoError(t, err)
			assert.Equal(t, got.LogConfig, tt.expectedLogConfig)
		})
	}
}

func TestRequestBuildRouterVersionWithDefaultConfig(t *testing.T) {
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
		AutoscalingPolicy: &models.AutoscalingPolicy{
			Metric: models.AutoscalingMetricCPU,
			Target: "80",
		},
		Timeout:  "10s",
		Protocol: routerConfig.HTTP,
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
			AutoscalingPolicy: &models.AutoscalingPolicy{
				Metric: models.AutoscalingMetricRPS,
				Target: "100",
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
				AutoscalingPolicy: &models.AutoscalingPolicy{
					Metric: models.AutoscalingMetricRPS,
					Target: "200",
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
	expSvc.On("IsClientSelectionEnabled", mock.Anything).Return(true, nil)

	// Set up mock Ensembler service
	ensemblerSvc := &mocks.EnsemblersService{}

	got, err := validRouterConfig.BuildRouterVersion("", router, &defaults, cryptoSvc, expSvc, ensemblerSvc)
	require.NoError(t, err)
	expected.Model = got.Model
	assertgotest.DeepEqual(t, expected, *got)
}

func TestRequestBuildRouterVersionWithUnavailablePyFuncEnsembler(t *testing.T) {
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

	// Get a default working router
	router := createOrUpdateRequest.BuildRouter(projectID)

	// Set up mock Crypto service
	cryptoSvc := &mocks.CryptoService{}
	cryptoSvc.On("Encrypt", "dummy_passkey").Return("enc_passkey", nil)

	// Set up mock Experiment service
	expSvc := &mocks.ExperimentsService{}
	expSvc.On("IsStandardExperimentManager", mock.Anything).Return(true)

	// Set up mock Ensembler service
	ensemblerSvc := &mocks.EnsemblersService{}
	ensemblerSvc.On("FindByID", mock.Anything, mock.Anything).Return(nil, errors.New("record not found"))

	// Update the router with an invalid request
	got, err := invalidRouterConfig.BuildRouterVersion("", router, &defaults, cryptoSvc, expSvc, ensemblerSvc)

	assert.EqualError(t, err, "failed to find specified ensembler: record not found")
	assert.Nil(t, got)
}

func TestRequestBuildRouterVersionNoDefaultRoute(t *testing.T) {
	defaults := &config.RouterDefaults{
		Image: "routerimage",
	}
	router := &models.Router{
		ProjectID:       models.ID(1),
		EnvironmentName: "env",
		Name:            "router",
		Status:          "pending",
	}
	cfg := RouterConfig{
		Routes: []*models.Route{
			{
				ID:       "default",
				Type:     "PROXY",
				Endpoint: "endpoint",
				Timeout:  "6s",
			},
		},
		Timeout:          "10s",
		ExperimentEngine: &ExperimentEngineConfig{Type: "nop"},
		LogConfig:        &LogConfig{ResultLoggerType: "nop"},
	}

	rv, err := cfg.BuildRouterVersion("", router, defaults, nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "", rv.DefaultRouteID)
}

func TestBuildExperimentEngineConfig(t *testing.T) {
	// Set up mock Crypto service
	cs := &mocks.CryptoService{}
	cs.On("Encrypt", "passkey-bad").
		Return("", errors.New("test-encrypt-error"))
	cs.On("Encrypt", "passkey").
		Return("passkey-enc", nil)

	es := &mocks.ExperimentsService{}
	es.On("IsClientSelectionEnabled", "standard-manager").Return(true, nil)
	es.On("IsClientSelectionEnabled", "custom-manager").Return(false, nil)

	// Define tests
	tests := map[string]struct {
		req      RouterConfig
		router   *models.Router
		expected json.RawMessage
		err      string
	}{
		"failure | std engine | missing curr version passkey": {
			req: RouterConfig{
				ExperimentEngine: &ExperimentEngineConfig{
					Type:   "standard-manager",
					Config: json.RawMessage(`{"client": {"username": "client-name"}}`),
				},
			},
			router: &models.Router{},
			err:    "Passkey must be configured",
		},
		"failure | std engine | encrypt passkey error": {
			req: RouterConfig{
				ExperimentEngine: &ExperimentEngineConfig{
					Type:   "standard-manager",
					Config: json.RawMessage(`{"client": {"username": "client-name", "passkey": "passkey-bad"}}`),
				},
			},
			router: &models.Router{},
			err:    "Passkey could not be encrypted: test-encrypt-error",
		},
		"success | std engine | use curr version passkey": {
			req: RouterConfig{
				ExperimentEngine: &ExperimentEngineConfig{
					Type:   "standard-manager",
					Config: json.RawMessage(`{"client": {"username": "client-name"}}`),
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
			req: RouterConfig{
				ExperimentEngine: &ExperimentEngineConfig{
					Type:   "standard-manager",
					Config: json.RawMessage(`{"client": {"username": "client-name", "passkey": "passkey"}}`),
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
			req: RouterConfig{
				ExperimentEngine: &ExperimentEngineConfig{
					Type:   "custom-manager",
					Config: json.RawMessage("[1, 2]"),
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

func TestDefaultAutoscalingPolicy(t *testing.T) {
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

	// Get a default working router
	projectID := models.ID(1)
	router := createOrUpdateRequest.BuildRouter(projectID)

	// Set up mock Crypto service
	cryptoSvc := &mocks.CryptoService{}
	cryptoSvc.On("Encrypt", "dummy_passkey").Return("enc_passkey", nil)

	// Set up mock Experiment service
	expSvc := &mocks.ExperimentsService{}
	expSvc.On("IsClientSelectionEnabled", mock.Anything).Return(true, nil)

	// Set up mock Ensembler service
	ensemblerSvc := &mocks.EnsemblersService{}

	var routerConfig = validRouterConfig
	routerConfig.AutoscalingPolicy = nil
	routerConfig.Enricher.AutoscalingPolicy = nil
	routerConfig.Ensembler.DockerConfig.AutoscalingPolicy = nil

	expectedAutoscalingPolicy := &models.AutoscalingPolicy{
		Metric: models.AutoscalingMetricConcurrency,
		Target: "1",
	}

	got, err := routerConfig.BuildRouterVersion("", router, &defaults, cryptoSvc, expSvc, ensemblerSvc)
	assert.NoError(t, err)
	assert.Equal(t, expectedAutoscalingPolicy, got.AutoscalingPolicy)
	assert.Equal(t, expectedAutoscalingPolicy, got.Enricher.AutoscalingPolicy)
	assert.Equal(t, expectedAutoscalingPolicy, got.Ensembler.DockerConfig.AutoscalingPolicy)
}
