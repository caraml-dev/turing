// +build unit

package config

import (
	"github.com/mitchellh/mapstructure"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gojek/mlp/pkg/instrumentation/sentry"
	tu "github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)


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

func TestFromFiles(t *testing.T) {
	tests := map[string]struct {
		filepaths []string
		env       map[string]string
		want      *Config
		wantErr   bool
	}{
		"default": {
			want: &Config{
				Port:           8080,
				AllowedOrigins: []string{"*"},
				AuthConfig:     &AuthorizationConfig{},
				DbConfig: &DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					Database: "turing",
				},
				DeployConfig: &DeploymentConfig{
					Timeout:         3 * time.Minute,
					DeletionTimeout: 1 * time.Minute,
					MaxCPU:          Quantity(resource.MustParse("4")),
					MaxMemory:       Quantity(resource.MustParse("8Gi")),
				},
				RouterDefaults: &RouterDefaults{
					LogLevel: "INFO",
					FluentdConfig: &FluentdConfig{
						Tag:                  "turing-result.log",
						FlushIntervalSeconds: 90,
					},
				},
				Sentry: sentry.Config{},
				VaultConfig: &VaultConfig{
					Address: "http://localhost:8200",
				},
				AlertConfig: &AlertConfig{
					GitLab: &GitlabConfig{
						BaseURL:    "https://gitlab.com",
						Branch:     "master",
						PathPrefix: "turing",
					},
				},
				MLPConfig: &MLPConfig{},
				TuringUIConfig: &TuringUIConfig{
					Homepage: "/turing",
				},
				SwaggerFile: "swagger.yaml",
			},
		},
		"single file": {
			filepaths: []string{"testdata/config-1.yaml"},
			want: &Config{
				Port:           9999,
				AllowedOrigins: []string{"http://foo.com", "http://bar.com"},
				AuthConfig: &AuthorizationConfig{
					Enabled: true,
					URL:     "http://example.com",
				},
				DbConfig: &DatabaseConfig{
					Host:     "127.0.0.1",
					Port:     5432,
					User:     "dbuser",
					Password: "dbpassword",
					Database: "turing",
				},
				DeployConfig: &DeploymentConfig{
					EnvironmentType: "dev",
					GcpProject:      "gcp-001",
					Timeout:         5 * time.Minute,
					DeletionTimeout: 1 * time.Minute,
					MaxCPU:          Quantity(resource.MustParse("500m")),
					MaxMemory:       Quantity(resource.MustParse("4000Mi")),
				},
				RouterDefaults: &RouterDefaults{
					LogLevel: "INFO",
					FluentdConfig: &FluentdConfig{
						Tag:                  "turing-result.log",
						FlushIntervalSeconds: 60,
					},
					Experiment: map[string]interface{}{
						"foo": map[string]interface{}{
							"fookey1": "fooval1",
							"fookey2": map[string]interface{}{
								"fookey2-1": "fooval2-1",
								"fookey2-2": "fooval2-2",
							},
						},
						"bar": map[string]interface{}{
							"barkey1": 8,
						},
					},
				},
				Sentry: sentry.Config{
					Enabled: true,
					Labels:  map[string]string{"foo": "bar"},
				},
				VaultConfig: &VaultConfig{
					Address: "http://localhost:8200",
					Token:   "root",
				},
				AlertConfig: &AlertConfig{
					GitLab: &GitlabConfig{
						BaseURL:    "https://gitlab.com",
						Branch:     "master",
						PathPrefix: "turing",
					},
				},
				MLPConfig: &MLPConfig{},
				TuringUIConfig: &TuringUIConfig{
					Homepage: "/turing",
				},
				SwaggerFile: "swagger.yaml",
				Experiment: map[string]interface{}{
					"qux": map[string]interface{}{
						"quxkey1": "quxval1",
						"quxkey2": map[string]interface{}{
							"quxkey2-1": "quxval2-1",
							"quxkey2-2": "quxval2-2",
						},
					},
					"quux": map[string]interface{}{
						"quuxkey1": "quuxval1",
					},
				},
			},
		},
		"multiple files": {
			filepaths: []string{"testdata/config-1.yaml", "testdata/config-2.yaml"},
			want: &Config{
				Port:           10000,
				AllowedOrigins: []string{"http://foo2.com"},
				AuthConfig: &AuthorizationConfig{
					Enabled: false,
					URL:     "http://example.com",
				},
				DbConfig: &DatabaseConfig{
					Host:     "127.0.0.1",
					Port:     5432,
					User:     "dbuser",
					Password: "newpassword",
					Database: "turing",
				},
				DeployConfig: &DeploymentConfig{
					EnvironmentType: "dev",
					GcpProject:      "gcp-001",
					Timeout:         5 * time.Minute,
					DeletionTimeout: 1 * time.Minute,
					MaxCPU:          Quantity(resource.MustParse("500m")),
					MaxMemory:       Quantity(resource.MustParse("12Gi")),
				},
				RouterDefaults: &RouterDefaults{
					LogLevel: "INFO",
					FluentdConfig: &FluentdConfig{
						Tag:                  "turing-result.log",
						FlushIntervalSeconds: 90,
					},
					Experiment: map[string]interface{}{
						"foo": map[string]interface{}{
							"fookey1": "",
							"fookey2": map[string]interface{}{
								"fookey2-1": "fooval2-1",
								"fookey2-2": "fooval2-2-override",
							},
						},
						"bar": map[string]interface{}{
							"barkey1": 8,
						},
						"qux": map[string]interface{}{
							"quux": "quuxval",
						},
					},
				},
				Sentry: sentry.Config{
					Enabled: true,
					Labels:  map[string]string{"foo": "bar"},
				},
				VaultConfig: &VaultConfig{
					Address: "http://localhost:8200",
					Token:   "root",
				},
				AlertConfig: &AlertConfig{
					GitLab: &GitlabConfig{
						BaseURL:    "https://gitlab.com",
						Branch:     "master",
						PathPrefix: "turing",
					},
				},
				MLPConfig: &MLPConfig{},
				TuringUIConfig: &TuringUIConfig{
					Homepage: "/turing",
				},
				SwaggerFile: "swagger.yaml",
				Experiment: map[string]interface{}{
					"qux": map[string]interface{}{
						"quxkey1": "quxval1-override",
						"quxkey2": map[string]interface{}{
							"quxkey2-1": "quxval2-1-override",
							"quxkey2-2": "quxval2-2",
							"quxkey2-3": "quxval2-3-add",
						},
					},
					"quux": map[string]interface{}{
						"quuxkey1": "quuxval1",
					},
				},
			},
		},
		"environment variables": {
			filepaths: []string{"testdata/config-1.yaml", "testdata/config-2.yaml"},
			env: map[string]string{
				"PORT":                                  "5000",
				"ALLOWEDORIGINS":                        "http://baz.com,http://qux.com",
				"AUTHCONFIG_ENABLED":                    "true",
				"AUTHCONFIG_URL":                        "http://env.example.com",
				"DBCONFIG_USER":                         "dbuser-env",
				"DBCONFIG_PASSWORD":                     "dbpassword-env",
				"DEPLOYCONFIG_TIMEOUT":                  "10m",
				"DEPLOYCONFIG_MAXMEMORY":                "4500Mi",
				"ROUTERDEFAULTS_EXPERIMENT_FOO_FOOKEY1": "fooval1-env",
				"ROUTERDEFAULTS_EXPERIMENT_QUX_QUUX":    "quuxval-env",
				"TURINGUICONFIG_APPDIRECTORY":           "appdir-env",
				"TURINGUICONFIG_HOMEPAGE":               "/turing-env",
				"EXPERIMENT_QUX_QUXKEY1":                "quxval1-env",
				"EXPERIMENT_QUX_QUXKEY2_QUXKEY2-1":      "quxval2-1-env",
			},
			want: &Config{
				Port:           5000,
				AllowedOrigins: []string{"http://baz.com", "http://qux.com"},
				AuthConfig: &AuthorizationConfig{
					Enabled: true,
					URL:     "http://env.example.com",
				},
				DbConfig: &DatabaseConfig{
					Host:     "127.0.0.1",
					Port:     5432,
					User:     "dbuser-env",
					Password: "dbpassword-env",
					Database: "turing",
				},
				DeployConfig: &DeploymentConfig{
					EnvironmentType: "dev",
					GcpProject:      "gcp-001",
					Timeout:         10 * time.Minute,
					DeletionTimeout: 1 * time.Minute,
					MaxCPU:          Quantity(resource.MustParse("500m")),
					MaxMemory:       Quantity(resource.MustParse("4500Mi")),
				},
				RouterDefaults: &RouterDefaults{
					LogLevel: "INFO",
					FluentdConfig: &FluentdConfig{
						Tag:                  "turing-result.log",
						FlushIntervalSeconds: 90,
					},
					Experiment: map[string]interface{}{
						"foo": map[string]interface{}{
							"fookey1": "fooval1-env",
							"fookey2": map[string]interface{}{
								"fookey2-1": "fooval2-1",
								"fookey2-2": "fooval2-2-override",
							},
						},
						"bar": map[string]interface{}{
							"barkey1": 8,
						},
						"qux": map[string]interface{}{
							"quux": "quuxval-env",
						},
					},
				},
				Sentry: sentry.Config{
					Enabled: true,
					Labels:  map[string]string{"foo": "bar"},
				},
				VaultConfig: &VaultConfig{
					Address: "http://localhost:8200",
					Token:   "root",
				},
				AlertConfig: &AlertConfig{
					GitLab: &GitlabConfig{
						BaseURL:    "https://gitlab.com",
						Branch:     "master",
						PathPrefix: "turing",
					},
				},
				MLPConfig: &MLPConfig{},
				TuringUIConfig: &TuringUIConfig{
					AppDirectory: "appdir-env",
					Homepage:     "/turing-env",
				},
				SwaggerFile: "swagger.yaml",
				Experiment: map[string]interface{}{
					"qux": map[string]interface{}{
						"quxkey1": "quxval1-env",
						"quxkey2": map[string]interface{}{
							"quxkey2-1": "quxval2-1-env",
							"quxkey2-2": "quxval2-2",
							"quxkey2-3": "quxval2-3-add",
						},
					},
					"quux": map[string]interface{}{
						"quuxkey1": "quuxval1",
					},
				},
			},
		},
		"missing file": {
			filepaths: []string{"this-file-should-not-exists.yaml"},
			wantErr: true,
		},
		"invalid duration format": {
			filepaths: []string{"invalid-duration-format.yaml"},
			wantErr: true,
		},
		"invalid quantity format": {
			filepaths: []string{"invalid-quantity-format.yaml"},
			wantErr: true,
		},
		"invalid type": {
			filepaths: []string{"invalid-type.yaml"},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			setupNewEnv(tt.env)
			got, err := FromFiles(tt.filepaths...)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// Reference:
// https://github.com/mitchellh/mapstructure/blob/ce2ff0c13ce509e36e9254c08ea0bca90ed5af6c/decode_hooks_test.go#L128
func TestStringToQuantityHookFunc(t *testing.T) {
	hookFunc := StringToQuantityHookFunc()
	strType := reflect.TypeOf("")
	qtyType := reflect.TypeOf(Quantity{})

	tests := []struct {
		name     string
		from, to reflect.Type
		data     interface{}
		want     interface{}
		wantErr  bool
	}{
		{
			name: "digit",
			from: strType,
			to:   qtyType,
			data: "5",
			want: Quantity(resource.MustParse("5")),
		},
		{
			name: "digit with suffix",
			from: strType,
			to:   qtyType,
			data: "5Gi",
			want: Quantity(resource.MustParse("5Gi")),
		},
		{
			name:    "empty",
			from:    strType,
			to:      qtyType,
			data:    "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			from:    strType,
			to:      qtyType,
			data:    "5GGi",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapstructure.DecodeHookExec(hookFunc, tt.from, tt.to, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToQuantityHookFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}