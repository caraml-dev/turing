//go:build e2e

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-playground/validator"
	"github.com/gojek/mlp/api/pkg/vault"
	"github.com/mitchellh/mapstructure"
	"github.com/ory/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type Config struct {
	TestID      string `yaml:"test_id" validate:"required" mapstructure:"test_id"`
	APIBasePath string `yaml:"api_base_path" validate:"required" mapstructure:"api_base_path"`

	Project ProjectConfig `validate:"required,dive"`
	Cluster ClusterConfig `validate:"required,dive"`

	Mockserver MockserverConfig `validate:"required,dive"`
	Echoserver EchoserverConfig `validate:"required,dive"`

	Vault VaultConfig `yaml:"vault" validate:"required,dive"`

	PythonVersions []string `yaml:"python_versions" validate:"required" mapstructure:"python_versions"`

	// Dynamically computed
	Ensemblers EnsemblerConfig `yaml:"ensemblers" validate:"required,dive"`

	// KubeconfigUseLocal specifies whether the test helper should use local Kube config to
	// authenticate to the cluster. The Kube config is assumed to be available at $HOME/.kube/config.
	// If false, the helper will use the cluster credentials from the configured Vault environment.
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

type ClusterConfig struct {
	Name        string `validate:"required"`
	Credentials *vault.ClusterSecret
}

type MockserverConfig struct {
	Image      string `validate:"required"`
	ServerName string `yaml:"name" validate:"required" mapstructure:"name"`
	Endpoint   string `yaml:"endpoint" mapstructure:"endpoint"`
}

type EchoserverConfig struct {
	Image string `validate:"required"`
}

type VaultConfig struct {
	Address string `yaml:"addr" validate:"required" mapstructure:"addr"`
	Token   string `yaml:"token" validate:"required" mapstructure:"token"`
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

	fmt.Println(v.AllKeys())

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	fmt.Println(v.AllKeys())

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
