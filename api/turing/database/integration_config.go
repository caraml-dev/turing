//go:build integration
// +build integration

package database

import (
	"os"
	"strconv"

	"github.com/gojek/turing/api/turing/config"
)

var (
	port, _      = strconv.Atoi(getEnvOrDefault("DATABASE_PORT", "5432"))
	mainDBConfig = &config.DatabaseConfig{
		Host:             getEnvOrDefault("DATABASE_HOST", "localhost"),
		Port:             port,
		User:             getEnvOrDefault("DATABASE_USER", "turing"),
		Password:         getEnvOrDefault("DATABASE_PASSWORD", "turing"),
		Database:         getEnvOrDefault("DATABASE_NAME", "turing"),
		MigrationsFolder: "../../db-migrations",
	}
)

func getTemporaryDBConfig(database string) *config.DatabaseConfig {
	cfg := *mainDBConfig
	cfg.Database = database
	return &cfg
}

func getEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
