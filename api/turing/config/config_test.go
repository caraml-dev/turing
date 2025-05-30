package config_test

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"
	"github.com/mitchellh/copystructure"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mlpcluster "github.com/caraml-dev/mlp/api/pkg/cluster"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"

	"github.com/caraml-dev/turing/api/turing/config"
	openapi "github.com/caraml-dev/turing/api/turing/generated"
)

func TestDecodeQuantity(t *testing.T) {
	tests := map[string]struct {
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
			var qty config.Quantity
			err := qty.Decode(data.value)
			if data.success {
				assert.NoError(t, err)
				assert.Equal(t, config.Quantity(data.expected), qty)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGetters(t *testing.T) {
	cfg := config.Config{
		Port: 5000,
	}
	assert.Equal(t, ":5000", cfg.ListenAddress())
}

func TestPDBConfigValidation(t *testing.T) {
	defaultInt := 20

	tests := map[string]struct {
		cfg     config.PodDisruptionBudgetConfig
		success bool
	}{
		"success pdb disabled": {
			cfg: config.PodDisruptionBudgetConfig{
				Enabled: false,
			},
			success: true,
		},
		"success pdb enabled, max unavailable exist": {
			cfg: config.PodDisruptionBudgetConfig{
				Enabled:                  true,
				MaxUnavailablePercentage: &defaultInt,
			},
			success: true,
		},
		"success pdb enabled, min available exist": {
			cfg: config.PodDisruptionBudgetConfig{
				Enabled:                true,
				MinAvailablePercentage: &defaultInt,
			},
			success: true,
		},
		"failure pdb enabled no max unavailable and min available": {
			cfg: config.PodDisruptionBudgetConfig{
				Enabled: true,
			},
			success: false,
		},
		"failure pdb enabled, both max unavailable and min available exist": {
			cfg: config.PodDisruptionBudgetConfig{
				Enabled:                  true,
				MaxUnavailablePercentage: &defaultInt,
				MinAvailablePercentage:   &defaultInt,
			},
			success: false,
		},
	}

	validate, err := config.NewConfigValidator()
	require.NoError(t, err)

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

func TestLoad(t *testing.T) {
	blueSvcAcctKeyFilePath := "/etc/plugins/blue/gcp_service_account/service-account.json"
	zeroSecond, _ := time.ParseDuration("0s")
	oneSecond, _ := time.ParseDuration("1s")
	twoSecond, _ := time.ParseDuration("2s")

	defaultMinAvailablePercentage := 20

	tests := map[string]struct {
		filepaths []string
		env       map[string]string
		want      *config.Config
		wantErr   bool
	}{
		"default": {
			want: &config.Config{
				Port:           8080,
				AllowedOrigins: []string{"*"},
				DbConfig: &config.DatabaseConfig{
					Host:             "localhost",
					Port:             5432,
					Database:         "turing",
					MigrationsFolder: "db-migrations/",
					ConnMaxIdleTime:  zeroSecond,
					ConnMaxLifetime:  zeroSecond,
					MaxIdleConns:     0,
					MaxOpenConns:     0,
				},
				DeployConfig: &config.DeploymentConfig{
					Timeout:           3 * time.Minute,
					DeletionTimeout:   1 * time.Minute,
					MaxCPU:            config.Quantity(resource.MustParse("4")),
					MaxMemory:         config.Quantity(resource.MustParse("8Gi")),
					MaxAllowedReplica: 20,
				},
				KnativeServiceDefaults: &config.KnativeServiceDefaults{
					QueueProxyResourcePercentage:          30,
					UserContainerCPULimitRequestFactor:    0,
					UserContainerMemoryLimitRequestFactor: 1,
				},
				RouterDefaults: &config.RouterDefaults{
					LogLevel: "INFO",
					FluentdConfig: &config.FluentdConfig{
						Tag:                  "turing-result.log",
						FlushIntervalSeconds: 90,
						WorkerCount:          1,
					},
					KafkaConfig: &config.KafkaConfig{
						MaxMessageBytes: 1048588,
						CompressionType: "none",
					},
				},
				Sentry: sentry.Config{},
				ClusterConfig: config.ClusterConfig{
					InClusterConfig: false,
				},
				AlertConfig: &config.AlertConfig{
					GitLab: &config.GitlabConfig{
						BaseURL:    "https://gitlab.com",
						Branch:     "master",
						PathPrefix: "turing",
					},
				},
				MLPConfig: &config.MLPConfig{},
				TuringUIConfig: &config.SinglePageApplicationConfig{
					ServingPath: "/turing",
				},
				MlflowConfig: &config.MlflowConfig{
					TrackingURL:         "",
					ArtifactServiceType: "nop",
				},
				OpenapiConfig: &config.OpenapiConfig{
					ValidationEnabled: true,
					SpecFile:          "api/openapi.bundle.yaml",
					MergedSpecFile:    "api/swagger-ui-dist/openapi.bundle.yaml",
					SwaggerUIConfig: &config.SinglePageApplicationConfig{
						ServingDirectory: "",
						ServingPath:      "/api-docs/",
					},
				},
			},
		},
		"single file": {
			filepaths: []string{"testdata/config-1.yaml"},
			want: &config.Config{
				Port:           9999,
				AllowedOrigins: []string{"http://foo.com", "http://bar.com"},
				DbConfig: &config.DatabaseConfig{
					Host:             "127.0.0.1",
					Port:             5432,
					User:             "dbuser",
					Password:         "dbpassword",
					Database:         "turing",
					MigrationsFolder: "db-migrations/",
					ConnMaxIdleTime:  oneSecond,
					ConnMaxLifetime:  twoSecond,
					MaxIdleConns:     3,
					MaxOpenConns:     4,
				},
				DeployConfig: &config.DeploymentConfig{
					EnvironmentType:   "dev",
					Timeout:           5 * time.Minute,
					DeletionTimeout:   1 * time.Minute,
					MaxCPU:            config.Quantity(resource.MustParse("500m")),
					MaxMemory:         config.Quantity(resource.MustParse("4000Mi")),
					MaxAllowedReplica: 20,
					TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
						{
							MaxSkew:           1,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.ScheduleAnyway,
						},
						{
							MaxSkew:           2,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.DoNotSchedule,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app-label": "spread",
								},
							},
						},
						{
							MaxSkew:           3,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.DoNotSchedule,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app-label": "spread",
								},
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "app-expression",
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{"1"},
									},
								},
							},
						},
					},
					PodDisruptionBudget: config.PodDisruptionBudgetConfig{
						Enabled:                true,
						MinAvailablePercentage: &defaultMinAvailablePercentage,
					},
				},
				KnativeServiceDefaults: &config.KnativeServiceDefaults{
					QueueProxyResourcePercentage:          20,
					UserContainerCPULimitRequestFactor:    0,
					UserContainerMemoryLimitRequestFactor: 1.25,
					DefaultEnvVarsWithoutCPULimits: []corev1.EnvVar{
						{
							Name:  "foo",
							Value: "bar",
						},
					},
				},
				RouterDefaults: &config.RouterDefaults{
					LogLevel: "INFO",
					FluentdConfig: &config.FluentdConfig{
						Tag:                  "turing-result.log",
						FlushIntervalSeconds: 60,
						WorkerCount:          2,
					},
					KafkaConfig: &config.KafkaConfig{
						MaxMessageBytes: 1048588,
						CompressionType: "none",
					},
				},
				Sentry: sentry.Config{
					Enabled: true,
					Labels:  map[string]string{"foo": "bar"},
				},
				ClusterConfig: config.ClusterConfig{
					InClusterConfig:       false,
					EnvironmentConfigPath: "path_to_env.yaml",
					EnsemblingServiceK8sConfig: &mlpcluster.K8sConfig{
						Name: "dev-server",
						Cluster: &clientcmdapiv1.Cluster{
							Server:                   "https://127.0.0.1",
							CertificateAuthorityData: []byte("some_string"),
						},
						AuthInfo: &clientcmdapiv1.AuthInfo{
							Exec: &clientcmdapiv1.ExecConfig{
								APIVersion:         "some_api_version",
								Command:            "some_command",
								InteractiveMode:    clientcmdapiv1.IfAvailableExecInteractiveMode,
								ProvideClusterInfo: true,
							},
						},
					},
				},
				AlertConfig: &config.AlertConfig{
					GitLab: &config.GitlabConfig{
						BaseURL:    "https://gitlab.com",
						Branch:     "master",
						PathPrefix: "turing",
					},
				},
				MLPConfig: &config.MLPConfig{},
				TuringUIConfig: &config.SinglePageApplicationConfig{
					ServingPath: "/turing",
				},
				OpenapiConfig: &config.OpenapiConfig{
					ValidationEnabled: true,
					SpecFile:          "api/openapi.bundle.yaml",
					MergedSpecFile:    "api/swagger-ui-dist/openapi.bundle.yaml",
					SwaggerUIConfig: &config.SinglePageApplicationConfig{
						ServingDirectory: "",
						ServingPath:      "/api-docs/",
					},
				},
				MlflowConfig: &config.MlflowConfig{
					TrackingURL:         "",
					ArtifactServiceType: "nop",
				},
				Experiment: map[string]config.EngineConfig{
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
			want: &config.Config{
				Port:           10000,
				LogLevel:       "DEBUG",
				AllowedOrigins: []string{"http://foo2.com"},
				DbConfig: &config.DatabaseConfig{
					Host:             "127.0.0.1",
					Port:             5432,
					User:             "dbuser",
					Password:         "newpassword",
					Database:         "turing",
					MigrationsFolder: "db-migrations/",
					ConnMaxIdleTime:  oneSecond,
					ConnMaxLifetime:  twoSecond,
					MaxIdleConns:     3,
					MaxOpenConns:     4,
				},
				DeployConfig: &config.DeploymentConfig{
					EnvironmentType:   "dev",
					Timeout:           5 * time.Minute,
					DeletionTimeout:   1 * time.Minute,
					MaxCPU:            config.Quantity(resource.MustParse("500m")),
					MaxMemory:         config.Quantity(resource.MustParse("12Gi")),
					MaxAllowedReplica: 30,
					TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
						{
							MaxSkew:           1,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.ScheduleAnyway,
						},
						{
							MaxSkew:           2,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.DoNotSchedule,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app-label": "spread",
								},
							},
						},
						{
							MaxSkew:           3,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.DoNotSchedule,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app-label": "spread",
								},
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "app-expression",
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{"1"},
									},
								},
							},
						},
					},
					PodDisruptionBudget: config.PodDisruptionBudgetConfig{
						Enabled:                true,
						MinAvailablePercentage: &defaultMinAvailablePercentage,
					},
				},
				KnativeServiceDefaults: &config.KnativeServiceDefaults{
					QueueProxyResourcePercentage:          20,
					UserContainerCPULimitRequestFactor:    0,
					UserContainerMemoryLimitRequestFactor: 1.25,
					DefaultEnvVarsWithoutCPULimits: []corev1.EnvVar{
						{
							Name:  "foo",
							Value: "bar",
						},
					},
				},
				RouterDefaults: &config.RouterDefaults{
					LogLevel: "INFO",
					FluentdConfig: &config.FluentdConfig{
						Tag:                  "turing-result.log",
						FlushIntervalSeconds: 90,
						WorkerCount:          2,
					},
					ExperimentEnginePlugins: map[string]*config.ExperimentEngineConfig{
						"red": {
							PluginConfig: &config.ExperimentEnginePluginConfig{
								Image:                 "ghcr.io/myproject/red-exp-engine-plugin:v0.0.1",
								LivenessPeriodSeconds: 5,
							},
						},
						"blue": {
							PluginConfig: &config.ExperimentEnginePluginConfig{
								Image:                 "ghcr.io/myproject/blue-exp-engine-plugin:latest",
								LivenessPeriodSeconds: 10,
							},
							ServiceAccountKeyFilePath: &blueSvcAcctKeyFilePath,
						},
					},
					KafkaConfig: &config.KafkaConfig{
						MaxMessageBytes: 1234567,
						CompressionType: "snappy",
					},
				},
				Sentry: sentry.Config{
					Enabled: true,
					Labels:  map[string]string{"foo": "bar"},
				},
				ClusterConfig: config.ClusterConfig{
					InClusterConfig:       false,
					EnvironmentConfigPath: "path_to_env.yaml",
					EnsemblingServiceK8sConfig: &mlpcluster.K8sConfig{
						Name: "dev-server",
						Cluster: &clientcmdapiv1.Cluster{
							Server:                   "https://127.0.0.1",
							CertificateAuthorityData: []byte("some_string"),
						},
						AuthInfo: &clientcmdapiv1.AuthInfo{
							Exec: &clientcmdapiv1.ExecConfig{
								APIVersion:         "some_api_version",
								Command:            "some_command",
								InteractiveMode:    clientcmdapiv1.IfAvailableExecInteractiveMode,
								ProvideClusterInfo: true,
							},
						},
					},
				},
				AlertConfig: &config.AlertConfig{
					GitLab: &config.GitlabConfig{
						BaseURL:    "https://gitlab.com",
						Branch:     "master",
						PathPrefix: "turing",
					},
				},
				MLPConfig: &config.MLPConfig{},
				TuringUIConfig: &config.SinglePageApplicationConfig{
					ServingPath: "/turing",
				},
				OpenapiConfig: &config.OpenapiConfig{
					ValidationEnabled: true,
					SpecFile:          "api/openapi.bundle.yaml",
					MergedSpecFile:    "api/swagger-ui-dist/openapi.bundle.yaml",
					SwaggerUIConfig: &config.SinglePageApplicationConfig{
						ServingDirectory: "",
						ServingPath:      "/swagger-ui",
					},
				},
				MlflowConfig: &config.MlflowConfig{
					TrackingURL:         "",
					ArtifactServiceType: "nop",
				},
				Experiment: map[string]config.EngineConfig{
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
		"multiple files and environment variables": {
			filepaths: []string{"testdata/config-1.yaml", "testdata/config-2.yaml"},
			env: map[string]string{
				"PORT":                                           "5000",
				"ALLOWEDORIGINS":                                 "http://baz.com,http://qux.com",
				"DBCONFIG_USER":                                  "dbuser-env",
				"DBCONFIG_PASSWORD":                              "dbpassword-env",
				"DEPLOYCONFIG_TIMEOUT":                           "10m",
				"DEPLOYCONFIG_MAXMEMORY":                         "4500Mi",
				"ROUTERDEFAULTS_EXPERIMENT_FOO_FOOKEY1":          "fooval1-env",
				"ROUTERDEFAULTS_EXPERIMENT_QUX_QUUX":             "quuxval-env",
				"TURINGUICONFIG_SERVINGDIRECTORY":                "appdir-env",
				"TURINGUICONFIG_SERVINGPATH":                     "/turing-env",
				"OPENAPICONFIG_SWAGGERUICONFIG_SERVINGDIRECTORY": "static/swagger-ui",
				"OPENAPICONFIG_SWAGGERUICONFIG_SERVINGPATH":      "/swagger-ui",
				"EXPERIMENT_QUX_QUXKEY1":                         "quxval1-env",
				"EXPERIMENT_QUX_QUXKEY2_QUXKEY2-1":               "quxval2-1-env",
				"CLUSTERCONFIG_ENVIRONMENTCONFIGPATH":            "env_var_path_to_env.yaml",
			},
			want: &config.Config{
				Port:           5000,
				LogLevel:       "DEBUG",
				AllowedOrigins: []string{"http://baz.com", "http://qux.com"},
				DbConfig: &config.DatabaseConfig{
					Host:             "127.0.0.1",
					Port:             5432,
					User:             "dbuser-env",
					Password:         "dbpassword-env",
					Database:         "turing",
					MigrationsFolder: "db-migrations/",
					ConnMaxIdleTime:  oneSecond,
					ConnMaxLifetime:  twoSecond,
					MaxIdleConns:     3,
					MaxOpenConns:     4,
				},
				KnativeServiceDefaults: &config.KnativeServiceDefaults{
					QueueProxyResourcePercentage:          20,
					UserContainerCPULimitRequestFactor:    0,
					UserContainerMemoryLimitRequestFactor: 1.25,
					DefaultEnvVarsWithoutCPULimits: []corev1.EnvVar{
						{
							Name:  "foo",
							Value: "bar",
						},
					},
				},
				DeployConfig: &config.DeploymentConfig{
					EnvironmentType:   "dev",
					Timeout:           10 * time.Minute,
					DeletionTimeout:   1 * time.Minute,
					MaxCPU:            config.Quantity(resource.MustParse("500m")),
					MaxMemory:         config.Quantity(resource.MustParse("4500Mi")),
					MaxAllowedReplica: 30,
					TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
						{
							MaxSkew:           1,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.ScheduleAnyway,
						},
						{
							MaxSkew:           2,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.DoNotSchedule,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app-label": "spread",
								},
							},
						},
						{
							MaxSkew:           3,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.DoNotSchedule,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app-label": "spread",
								},
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "app-expression",
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{"1"},
									},
								},
							},
						},
					},
					PodDisruptionBudget: config.PodDisruptionBudgetConfig{
						Enabled:                true,
						MinAvailablePercentage: &defaultMinAvailablePercentage,
					},
				},
				RouterDefaults: &config.RouterDefaults{
					LogLevel: "INFO",
					FluentdConfig: &config.FluentdConfig{
						Tag:                  "turing-result.log",
						FlushIntervalSeconds: 90,
						WorkerCount:          2,
					},
					ExperimentEnginePlugins: map[string]*config.ExperimentEngineConfig{
						"red": {
							PluginConfig: &config.ExperimentEnginePluginConfig{
								Image:                 "ghcr.io/myproject/red-exp-engine-plugin:v0.0.1",
								LivenessPeriodSeconds: 5,
							},
						},
						"blue": {
							PluginConfig: &config.ExperimentEnginePluginConfig{
								Image:                 "ghcr.io/myproject/blue-exp-engine-plugin:latest",
								LivenessPeriodSeconds: 10,
							},
							ServiceAccountKeyFilePath: &blueSvcAcctKeyFilePath,
						},
					},
					KafkaConfig: &config.KafkaConfig{
						MaxMessageBytes: 1234567,
						CompressionType: "snappy",
					},
				},
				Sentry: sentry.Config{
					Enabled: true,
					Labels:  map[string]string{"foo": "bar"},
				},
				ClusterConfig: config.ClusterConfig{
					InClusterConfig:       false,
					EnvironmentConfigPath: "env_var_path_to_env.yaml",
					EnsemblingServiceK8sConfig: &mlpcluster.K8sConfig{
						Name: "dev-server",
						Cluster: &clientcmdapiv1.Cluster{
							Server:                   "https://127.0.0.1",
							CertificateAuthorityData: []byte("some_string"),
						},
						AuthInfo: &clientcmdapiv1.AuthInfo{
							Exec: &clientcmdapiv1.ExecConfig{
								APIVersion:         "some_api_version",
								Command:            "some_command",
								InteractiveMode:    clientcmdapiv1.IfAvailableExecInteractiveMode,
								ProvideClusterInfo: true,
							},
						},
					},
				},
				AlertConfig: &config.AlertConfig{
					GitLab: &config.GitlabConfig{
						BaseURL:    "https://gitlab.com",
						Branch:     "master",
						PathPrefix: "turing",
					},
				},
				MLPConfig: &config.MLPConfig{},
				TuringUIConfig: &config.SinglePageApplicationConfig{
					ServingDirectory: "appdir-env",
					ServingPath:      "/turing-env",
				},
				OpenapiConfig: &config.OpenapiConfig{
					ValidationEnabled: true,
					SpecFile:          "api/openapi.bundle.yaml",
					MergedSpecFile:    "api/swagger-ui-dist/openapi.bundle.yaml",
					SwaggerUIConfig: &config.SinglePageApplicationConfig{
						ServingDirectory: "static/swagger-ui",
						ServingPath:      "/swagger-ui",
					},
				},
				MlflowConfig: &config.MlflowConfig{
					TrackingURL:         "",
					ArtifactServiceType: "nop",
				},
				Experiment: map[string]config.EngineConfig{
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
			wantErr:   true,
		},
		"invalid duration format": {
			filepaths: []string{"invalid-duration-format.yaml"},
			wantErr:   true,
		},
		"invalid quantity format": {
			filepaths: []string{"invalid-quantity-format.yaml"},
			wantErr:   true,
		},
		"invalid type": {
			filepaths: []string{"invalid-type.yaml"},
			wantErr:   true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			setupNewEnv(tt.env)
			got, err := config.Load(tt.filepaths...)
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
	hookFunc := config.StringToQuantityHookFunc()
	qtyType := reflect.TypeOf(config.Quantity{})

	tests := []struct {
		name    string
		from    interface{}
		to      reflect.Type
		want    interface{}
		wantErr bool
	}{
		{
			name: "digit",
			from: "5",
			to:   qtyType,
			want: config.Quantity(resource.MustParse("5")),
		},
		{
			name: "digit with suffix",
			from: "5Gi",
			to:   qtyType,
			want: config.Quantity(resource.MustParse("5Gi")),
		},
		{
			name:    "empty",
			from:    "",
			to:      qtyType,
			wantErr: true,
		},
		{
			name:    "invalid format",
			from:    "5GGi",
			to:      qtyType,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapstructure.DecodeHookExec(hookFunc, reflect.ValueOf(tt.from), reflect.Zero(tt.to))
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToQuantityHookFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfigValidate(t *testing.T) {
	driverCPURequest := "1"
	driverMemoryRequest := "1Gi"
	var executorReplica int32 = 2
	executorCPURequest := "1"
	executorMemoryRequest := "1Gi"
	tolerationName := "batch-job"
	validConfig := config.Config{
		Port: 5000,
		BatchEnsemblingConfig: config.BatchEnsemblingConfig{
			Enabled: true,
			JobConfig: &config.JobConfig{
				DefaultEnvironment: "dev",
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
					PushRegistryType: "docker",
				},
			},
		},
		EnsemblerServiceBuilderConfig: config.EnsemblerServiceBuilderConfig{
			ClusterName: "dev",
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
					PushRegistryType: "docker",
				},
			},
		},
		DbConfig: &config.DatabaseConfig{
			Host:             "localhost",
			Port:             5432,
			User:             "user",
			Password:         "password",
			Database:         "postgres",
			MigrationsFolder: "db-migrations/",
		},
		DeployConfig: &config.DeploymentConfig{
			EnvironmentType:   "dev",
			Timeout:           1 * time.Minute,
			DeletionTimeout:   1 * time.Minute,
			MaxCPU:            config.Quantity(resource.MustParse("2")),
			MaxMemory:         config.Quantity(resource.MustParse("8Gi")),
			MaxAllowedReplica: 30,
		},
		MlflowConfig: &config.MlflowConfig{
			TrackingURL:         "http://localhost:8081",
			ArtifactServiceType: "gcs",
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
			Image:    "turing-router:latest",
			LogLevel: "DEBUG",
		},
		KubernetesLabelConfigs: &config.KubernetesLabelConfigs{
			Environment: "dev",
		},
		Sentry: sentry.Config{},
		NewRelicConfig: newrelic.Config{
			Enabled:           true,
			AppName:           "test",
			License:           "test",
			IgnoreStatusCodes: []int{403, 404},
		},
		ClusterConfig: config.ClusterConfig{
			InClusterConfig:       false,
			EnvironmentConfigPath: "./path/to/env-file.yaml",
			EnsemblingServiceK8sConfig: &mlpcluster.K8sConfig{
				Cluster:  &clientcmdapiv1.Cluster{},
				AuthInfo: &clientcmdapiv1.AuthInfo{},
				Name:     "dev",
			},
		},
		TuringEncryptionKey: "secret",
		AlertConfig:         nil,
		MLPConfig: &config.MLPConfig{
			MerlinURL: "http://merlin.example.com",
			MLPURL:    "http://mlp.example.com",
		},
	}

	// validConfigUpdate returns an updated config from a valid one
	type validConfigUpdate func(validConfig config.Config) config.Config

	tests := map[string]struct {
		validConfigUpdate validConfigUpdate
		wantErr           bool
	}{
		"valid": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				return validConfig
			},
		},
		"missing port": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.Port = 0
				return validConfig
			},
			wantErr: true,
		},
		"missing database password": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.DbConfig.Password = ""
				return validConfig
			},
			wantErr: true,
		},
		"missing deployment timeout": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.DeployConfig.Timeout = 0
				return validConfig
			},
			wantErr: true,
		},
		"missing turing encryption key": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.TuringEncryptionKey = ""
				return validConfig
			},
			wantErr: true,
		},
		"missing MLP URL": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.MLPConfig.MLPURL = ""
				return validConfig
			},
			wantErr: true,
		},
		"missing Merlin URL": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.MLPConfig.MerlinURL = ""
				return validConfig
			},
			wantErr: true,
		},
		"missing ensembling job default environment": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.BatchEnsemblingConfig.JobConfig.DefaultEnvironment = ""
				return validConfig
			},
			wantErr: true,
		},
		"missing spark infra config": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.SparkAppConfig = nil
				return validConfig
			},
			wantErr: true,
		},
		"missing kubernetes label config": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.KubernetesLabelConfigs = nil
				return validConfig
			},
			wantErr: true,
		},
		"missing EnvironmentConfigPath when InClusterConfig is false": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.ClusterConfig.EnvironmentConfigPath = ""
				validConfig.ClusterConfig.InClusterConfig = false
				return validConfig
			},
			wantErr: true,
		},
		"missing EnsemblingServiceK8sConfig when InClusterConfig is false": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.ClusterConfig.EnsemblingServiceK8sConfig = nil
				validConfig.ClusterConfig.InClusterConfig = false
				return validConfig
			},
			wantErr: true,
		},
		"valid in cluster config": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.ClusterConfig.InClusterConfig = true
				return validConfig
			},
			wantErr: false,
		},
		"valid batch ensembling disabled": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.BatchEnsemblingConfig = config.BatchEnsemblingConfig{
					Enabled: false,
				}
				return validConfig
			},
			wantErr: false,
		},
		"batch ensembling enabled but missing settings": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.BatchEnsemblingConfig.JobConfig.DefaultEnvironment = ""
				return validConfig
			},
			wantErr: true,
		},
		"batch ensembling enabled but one whole section missing": {
			validConfigUpdate: func(validConfig config.Config) config.Config {
				validConfig.BatchEnsemblingConfig.JobConfig = nil
				return validConfig
			},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validConfigCopy := copystructure.Must(copystructure.Copy(validConfig))
			c := tt.validConfigUpdate(validConfigCopy.(config.Config))
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessEnvConfigs(t *testing.T) {
	tests := map[string]struct {
		filepath      string
		want          []*config.EnvironmentConfig
		wantErr       bool
		wantInCluster bool
	}{
		"basic": {
			filepath: "testdata/env-config-1.yaml",
			want: []*config.EnvironmentConfig{
				{
					Name: "id-dev",
					K8sConfig: &mlpcluster.K8sConfig{
						Name: "dev-cluster",
						Cluster: &clientcmdapiv1.Cluster{
							Server:                "https://k8s.api.server",
							InsecureSkipTLSVerify: true,
						},
						AuthInfo: &clientcmdapiv1.AuthInfo{
							Exec: &clientcmdapiv1.ExecConfig{
								APIVersion:         "client.authentication.k8s.io/v1beta1",
								Command:            "gke-gcloud-auth-plugin",
								InteractiveMode:    clientcmdapiv1.IfAvailableExecInteractiveMode,
								ProvideClusterInfo: true,
							},
						},
					},
				},
			},
		},
		"error_parsing": {
			filepath: "testdata/env-err.yaml",
			want:     []*config.EnvironmentConfig{},
			wantErr:  true,
		},
		"InClusterConfig": {
			want:          []*config.EnvironmentConfig(nil),
			filepath:      "testdata/env-config-1.yaml",
			wantInCluster: true,
		},
		"env k8sconfig is missing": {
			filepath: "testdata/env-config-nil.yaml",
			want:     []*config.EnvironmentConfig{},
			wantErr:  true,
		},
	}
	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			c := config.ClusterConfig{
				EnvironmentConfigPath: tt.filepath,
			}
			if tt.wantInCluster {
				c.InClusterConfig = true
			}
			if err := c.ProcessEnvConfigs(); (err != nil) != tt.wantErr {
				t.Errorf("ProcessEnvConfigs() error = %v, wantErr %v", err, tt.wantErr)
			} else if !tt.wantErr {
				assert.Equal(t, tt.want, c.EnvironmentConfigs)
			}
		})
	}
}
