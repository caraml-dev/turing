package vault

import (
	"github.com/gojek/mlp/api/pkg/vault"
	"github.com/pkg/errors"

	"github.com/caraml-dev/turing/api/turing/config"
)

// NewClientFromConfig creates a vault client from the given config
func NewClientFromConfig(cfg *config.Config) (vault.VaultClient, error) {
	if cfg.ClusterConfig.InClusterConfig {
		// Here we don't need vault, we can use in cluster credentials
		return nil, nil
	}
	vaultConfig := &vault.Config{
		Address: cfg.ClusterConfig.VaultConfig.Address,
		Token:   cfg.ClusterConfig.VaultConfig.Token,
	}
	vaultClient, err := vault.NewVaultClient(vaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize vault")
	}
	return vaultClient, nil
}
