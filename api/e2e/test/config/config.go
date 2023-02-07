package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-playground/validator/v10"
	mlpcluster "github.com/gojek/mlp/api/pkg/cluster"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	sigyaml "sigs.k8s.io/yaml"
)

type Config struct {
	TestID      string `yaml:"test_id" validate:"required" mapstructure:"test_id"`
	APIBasePath string `yaml:"api_base_path" validate:"required" mapstructure:"api_base_path"`

	Project ProjectConfig `validate:"required,dive"`
	Cluster ClusterConfig `validate:"required,dive"`

	Mockserver             MockserverConfig `yaml:"mockserver" validate:"required,dive"`
	MockControlUPIServer   MockserverConfig `yaml:"mockControlUPIServer" validate:"required,dive"`
	MockTreatmentUPIServer MockserverConfig `yaml:"mockTreatmentUPIServer" validate:"required,dive"`
	Echoserver             EchoserverConfig `validate:"required,dive"`

	PythonVersions []string `yaml:"python_versions" validate:"required" mapstructure:"python_versions"`

	// Dynamically computed
	Ensemblers EnsemblerConfig `yaml:"ensemblers" validate:"required,dive"`

	// KubeconfigUseLocal specifies whether the test helper should use local Kube config to
	// authenticate to the cluster. The Kube config is assumed to be available at $HOME/.kube/config.
	// If false, the helper will use the cluster credentials config.yaml.
	KubeconfigUseLocal bool   `yaml:"kubeconfig_use_local" default:"false" mapstructure:"kubeconfig_use_local"`
	KubeconfigFilePath string `yaml:"kubeconfig_file_path" default:"" mapstructure:"kubeconfig_file_path"`
}

func (c Config) String() string {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(b)
}

func DefaultConfig() *Config {
	return &Config{
		TestID: fmt.Sprint(time.Now().Unix()),
	}
}

type ProjectConfig struct {
	ID   int
	Name string `validate:"required"`
}

type K8sConfig mlpcluster.K8sConfig

// MarshalYAML implements the Marshal interface,
// so k8sConfig fields can be properly marshalled
func (k *K8sConfig) MarshalYAML() (interface{}, error) {
	output, err := sigyaml.Marshal(k)
	if err != nil {
		return nil, err
	}
	mapStrFormat := make(map[string]interface{})
	if err = sigyaml.Unmarshal(output, &mapStrFormat); err != nil {
		return nil, err
	}
	return mapStrFormat, nil
}

// UnmarshalYAML implements Unmarshal interface
// Since K8sConfig fields only have json tags, sigyaml.Unmarshal needs to be used
// to unmarshal all the fields. This method reads K8sConfig into a map[string]interface{},
// marshals it into a byte for, before passing to sigyaml.Unmarshal
func (k *K8sConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var kubeconfig map[string]interface{}
	// Unmarshal into map[string]interface{}
	if err := unmarshal(&kubeconfig); err != nil {
		return err
	}
	// convert back to byte string
	byteForm, err := yaml.Marshal(kubeconfig)
	if err != nil {
		return err
	}
	// use sigyaml.Unmarshal to convert to json object then unmarshal
	if err := sigyaml.Unmarshal(byteForm, k); err != nil {
		return err
	}
	return nil
}

type ClusterConfig struct {
	Name        string     `validate:"required"`
	Credentials *K8sConfig `validate:"required" yaml:"credentials" mapstructure:"credentials"`
}

type MockserverConfig struct {
	Image      string `validate:"required"`
	ServerName string `yaml:"name" validate:"required" mapstructure:"name"`
	Endpoint   string `yaml:"endpoint" mapstructure:"endpoint"`
}

type EchoserverConfig struct {
	Image string `validate:"required"`
}

type EnsemblerData struct {
	PythonVersion string
	EnsemblerID   int
}

type EnsemblerConfig struct {
	BaseName string `yaml:"base_name" validate:"required" mapstructure:"base_name"`
	Entities []EnsemblerData
}

func LoadFromFiles(filepaths ...string) (*Config, error) {
	v := viper.New()

	b, _ := yaml.Marshal(DefaultConfig())
	defaultConfig := bytes.NewReader(b)
	v.SetConfigType("yaml")
	_ = v.MergeConfig(defaultConfig)

	for _, f := range filepaths {
		v.SetConfigFile(f)
		err := v.MergeInConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to read config from file '%s': %s", f, err)
		}
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	config := &Config{}
	err := v.Unmarshal(config, func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(" "),
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config values: %s", err)
	}

	// NOTE: This section is added to parse any fields in k8sConfig that does not
	// have yaml tags.
	// For example `certificate-authority-data` is not unmarshalled
	// by vipers unmarshal method.
	var byteForm []byte
	// convert back to byte string
	byteForm, err = yaml.Marshal(v.AllSettings())
	if err != nil {
		return nil, err
	}
	// use sigyaml.Unmarshal to convert to json object then unmarshal
	if err := sigyaml.Unmarshal(byteForm, config); err != nil {
		return nil, err
	}

	err = validator.New().Struct(config)
	if err != nil {
		return nil, fmt.Errorf("failed config validation: %s", err)
	}

	return config, nil
}

type AssertLoggerT interface {
	assert.TestingT
	httpexpect.Logger
}

func NewHTTPExpect(t AssertLoggerT, baseURL string) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  baseURL,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{httpexpect.NewDebugPrinter(t, true)},
	})
}
