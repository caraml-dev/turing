package vault

import (
	"github.com/gojek/mlp/api/pkg/vault"
	"github.com/gojek/turing/api/turing/config"
	"github.com/pkg/errors"
)

// NewClientFromConfig creates a vault client from the given config
func NewClientFromConfig(cfg *config.Config) (vault.VaultClient, error) {
	vaultConfig := &vault.Config{
		Address: cfg.VaultConfig.Address,
		Token:   cfg.VaultConfig.Token,
	}
	vaultClient, err := vault.NewVaultClient(vaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize vault")
	}
	return vaultClient, nil
}
