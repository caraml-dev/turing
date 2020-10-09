// +build unit

package config

import (
	"os"
	"testing"
	"time"

	"github.com/gojek/mlp/pkg/instrumentation/sentry"
	tu "github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

var requiredEnvs = map[string]string{
	"TURING_DATABASE_HOST":                    "turing-db-host",
	"TURING_DATABASE_USER":                    "turing-db-user",
	"TURING_DATABASE_PASSWORD":                "turing-db-pass",
	"TURING_DATABASE_NAME":                    "turing-db-name",
	"TURING_ROUTER_IMAGE":                     "asia.gcr.io/gcp-project-id/turing-router:1.0.0",
	"TURING_ROUTER_JAEGER_COLLECTOR_ENDPOINT": "lala",
	"TURING_ROUTER_FLUENTD_IMAGE":             "image",
	"TURING_ROUTER_LITMUS_GRPC_ENDPOINT":      "grpc://litmus",
	"TURING_ROUTER_XP_HTTP_ENDPOINT":          "http://xp",
	"DEPLOYMENT_ENVIRONMENTS":                 "env1,env2",
	"DEPLOYMENT_ENVIRONMENT_TYPE":             "dev",
	"DEPLOYMENT_GCP_PROJECT":                  "gcp-project",
	"DEPLOYMENT_TIMEOUT":                      "10s",
	"DEPLOYMENT_DELETION_TIMEOUT":             "1s",
	"DEPLOYMENT_MAX_CPU":                      "200m",
	"DEPLOYMENT_MAX_MEMORY":                   "100Mi",
	"AUTHORIZATION_URL":                       "test-auth-url",
	"VAULT_ADDRESS":                           "vault-addr",
	"VAULT_TOKEN":                             "vault-token",
	"MERLIN_URL":                              "http://mlp.example.com/api/merlin/v1",
	"MLP_URL":                                 "http://mlp.example.com/api/mlp/v1",
	"MLP_ENCRYPTION_KEY":                      "key",
	"TURING_ENCRYPTION_KEY":                   "turing-key",
	"LITMUS_CLIENT_ID":                        "client",
	"LITMUS_PASSKEY":                          "pk",
	"LITMUS_CAS_TOKEN":                        "cas",
	"LITMUS_BASE_URL":                         "test_url",
	"XP_CLIENT_ID":                            "xp-client",
	"XP_PASSKEY":                              "xp-passkey",
	"XP_BASE_URL":                             "xp-host",
	"XP_USE_MOCK_DATA":                        "false",
}

var optionalEnvs = map[string]string{
	"TURING_PORT":                                  "5000",
	"TURING_DATABASE_PORT":                         "5433",
	"TURING_ROUTER_FIBER_DEBUG_LOG_ENABLED":        "false",
	"TURING_ROUTER_CUSTOM_METRICS_ENABLED":         "false",
	"TURING_ROUTER_FLUENTD_FLUSH_INTERVAL_SECONDS": "10",
	"TURING_ROUTER_JAEGER_ENABLED":                 "false",
	"TURING_ROUTER_LOG_LEVEL":                      "DEBUG",
	"TURING_ROUTER_FLUENTD_TAG":                    "turing-result2.log",
	"TURING_ROUTER_LITMUS_TIMEOUT":                 "50ms",
	"TURING_ROUTER_XP_TIMEOUT":                     "70ms",
	"AUTHORIZATION_ENABLED":                        "false",
	"ALERT_ENABLED":                                "true",
	"ALERT_GITLAB_BASEURL":                         "https://gitlab.com",
	"ALERT_GITLAB_TOKEN":                           "apitoken",
	"ALERT_GITLAB_PROJECTID":                       "12345",
	"ALERT_GITLAB_BRANCH":                          "dev",
	"ALERT_GITLAB_PATHPREFIX":                      "prefix",
	"SENTRY_ENABLED":                               "true",
	"SENTRY_DSN":                                   "sentry-test-dsn",
	"SENTRY_LABELS":                                "sentry_key1:value1,sentry_key2:value2",
	"TURING_UI_APP_DIRECTORY":                      "./build",
	"TURING_UI_HOMEPAGE":                           "/turing-ui",
	"SWAGGER_FILE":                                 "/custom/path/swagger.yaml",
	"XP_ENABLED":                                   "true",
}

func TestMissingRequiredEnvs(t *testing.T) {
	// Setup
	setupNewEnv()
	_, err := FromEnv()
	if err == nil {
		t.Error("Expected init config to fail, but it succeeded.")
	}
}

func TestInitConfigDefaultEnvs(t *testing.T) {
	timeout, _ := time.ParseDuration("10s")
	delTimeout, _ := time.ParseDuration("1s")
	expected := Config{
		Port: 8080,
		AuthConfig: &AuthorizationConfig{
			Enabled: true,
			URL:     "test-auth-url",
		},
		DbConfig: &DatabaseConfig{
			Host:     "turing-db-host",
			Port:     5432,
			User:     "turing-db-user",
			Password: "turing-db-pass",
			Database: "turing-db-name",
		},
		DeployConfig: &DeploymentConfig{
			EnvironmentType: "dev",
			GcpProject:      "gcp-project",
			Timeout:         timeout,
			DeletionTimeout: delTimeout,
			MaxCPU:          Quantity(resource.MustParse("200m")),
			MaxMemory:       Quantity(resource.MustParse("100Mi")),
		},
		RouterDefaults: &RouterDefaults{
			Image:                   "asia.gcr.io/gcp-project-id/turing-router:1.0.0",
			FiberDebugLogEnabled:    true,
			CustomMetricsEnabled:    true,
			JaegerEnabled:           true,
			JaegerCollectorEndpoint: "lala",
			LogLevel:                "INFO",
			FluentdConfig: &FluentdConfig{
				Image:                "image",
				Tag:                  "turing-result.log",
				FlushIntervalSeconds: 90,
			},
			LitmusGRPCEndpoint: "grpc://litmus",
			LitmusTimeout:      "60ms",
			XpHTTPEndpoint:     "http://xp",
			XpTimeout:          "60ms",
		},
		Sentry: sentry.Config{
			Enabled: false,
			DSN:     "",
			Labels:  nil,
		},
		VaultConfig: &VaultConfig{
			Address: "vault-addr",
			Token:   "vault-token",
		},
		MLPConfig: &MLPConfig{
			MerlinURL:        "http://mlp.example.com/api/merlin/v1",
			MLPURL:           "http://mlp.example.com/api/mlp/v1",
			MLPEncryptionKey: "key",
		},
		TuringEncryptionKey: "turing-key",
		AlertConfig: &AlertConfig{
			Enabled: false,
			GitLab: &GitlabConfig{
				Branch:     "master",
				PathPrefix: "turing",
			},
		},
		LitmusConfig: &LitmusConfig{
			ClientID: "client",
			Passkey:  "pk",
			CasToken: "cas",
			BaseURL:  "test_url",
		},
		XPConfig: &XPConfig{
			Enabled:     false,
			UseMockData: false,
			ClientID:    "xp-client",
			Passkey:     "xp-passkey",
			BaseURL:     "xp-host",
		},
		TuringUIConfig: &TuringUIConfig{
			Homepage: "/turing",
		},
		SwaggerFile: "swagger.yaml",
	}

	// Setup
	setupNewEnv(requiredEnvs)
	cfg, err := FromEnv()
	tu.FailOnError(t, err)

	// Test
	assert.Equal(t, cfg, &expected)
}

func TestInitConfigEnv(t *testing.T) {
	timeout, _ := time.ParseDuration("10s")
	delTimeout, _ := time.ParseDuration("1s")
	expected := Config{
		Port: 5000,
		AuthConfig: &AuthorizationConfig{
			Enabled: false,
			URL:     "test-auth-url",
		},
		DbConfig: &DatabaseConfig{
			Host:     "turing-db-host",
			Port:     5433,
			User:     "turing-db-user",
			Password: "turing-db-pass",
			Database: "turing-db-name",
		},
		RouterDefaults: &RouterDefaults{
			Image:                   "asia.gcr.io/gcp-project-id/turing-router:1.0.0",
			FiberDebugLogEnabled:    false,
			CustomMetricsEnabled:    false,
			JaegerEnabled:           false,
			JaegerCollectorEndpoint: "lala",
			LogLevel:                "DEBUG",
			FluentdConfig: &FluentdConfig{
				Image:                "image",
				Tag:                  "turing-result2.log",
				FlushIntervalSeconds: 10,
			},
			LitmusGRPCEndpoint: "grpc://litmus",
			LitmusTimeout:      "50ms",
			XpHTTPEndpoint:     "http://xp",
			XpTimeout:          "70ms",
		},
		DeployConfig: &DeploymentConfig{
			EnvironmentType: "dev",
			GcpProject:      "gcp-project",
			Timeout:         timeout,
			DeletionTimeout: delTimeout,
			MaxCPU:          Quantity(resource.MustParse("200m")),
			MaxMemory:       Quantity(resource.MustParse("100Mi")),
		},
		Sentry: sentry.Config{
			Enabled: true,
			DSN:     "sentry-test-dsn",
			Labels: map[string]string{
				"sentry_key1": "value1",
				"sentry_key2": "value2",
			},
		},
		VaultConfig: &VaultConfig{
			Address: "vault-addr",
			Token:   "vault-token",
		},
		MLPConfig: &MLPConfig{
			MerlinURL:        "http://mlp.example.com/api/merlin/v1",
			MLPURL:           "http://mlp.example.com/api/mlp/v1",
			MLPEncryptionKey: "key",
		},
		TuringEncryptionKey: "turing-key",
		AlertConfig: &AlertConfig{
			Enabled: true,
			GitLab: &GitlabConfig{
				BaseURL:    "https://gitlab.com",
				Token:      "apitoken",
				ProjectID:  "12345",
				Branch:     "dev",
				PathPrefix: "prefix",
			},
		},
		LitmusConfig: &LitmusConfig{
			ClientID: "client",
			Passkey:  "pk",
			CasToken: "cas",
			BaseURL:  "test_url",
		},
		XPConfig: &XPConfig{
			Enabled:     true,
			UseMockData: false,
			ClientID:    "xp-client",
			Passkey:     "xp-passkey",
			BaseURL:     "xp-host",
		},
		TuringUIConfig: &TuringUIConfig{
			AppDirectory: "./build",
			Homepage:     "/turing-ui",
		},
		SwaggerFile: "/custom/path/swagger.yaml",
	}

	// Setup
	setupNewEnv(requiredEnvs, optionalEnvs)
	cfg, err := FromEnv()
	tu.FailOnError(t, err)

	// Test
	assert.Equal(t, cfg, &expected)
}

func TestDecodeQuantity(t *testing.T) {
	var tests = map[string]struct {
		value    string
		success  bool
		expected resource.Quantity
	}{
		"success": {
			value:    "300m",
			success:  true,
			expected: resource.MustParse("300m"),
		},
		"empty value": {
			value:   "",
			success: false,
		},
		"bad value": {
			value:   "abc",
			success: false,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Run and validate
			var qty Quantity
			err := qty.Decode(data.value)
			if data.success {
				assert.NoError(t, err)
				assert.Equal(t, Quantity(data.expected), qty)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGetters(t *testing.T) {
	cfg := Config{
		Port: 5000,
	}
	assert.Equal(t, ":5000", cfg.ListenAddress())
}

func TestAuthConfigValidation(t *testing.T) {
	tests := map[string]struct {
		cfg     AuthorizationConfig
		success bool
	}{
		"success auth disabled": {
			cfg: AuthorizationConfig{
				Enabled: false,
			},
			success: true,
		},
		"success auth enabled": {
			cfg: AuthorizationConfig{
				Enabled: true,
				URL:     "url",
			},
			success: true,
		},
		"failure auth enabled no url": {
			cfg: AuthorizationConfig{
				Enabled: true,
			},
			success: false,
		},
	}

	validate, err := newConfigValidator()
	tu.FailOnError(t, err)

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			err := validate.Struct(data.cfg)
			if data.success {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func setupNewEnv(envMaps ...map[string]string) {
	os.Clearenv()

	for _, envMap := range envMaps {
		for key, val := range envMap {
			os.Setenv(key, val)
		}
	}
}
