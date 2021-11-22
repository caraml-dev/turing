// +build integration

package database

import (
	"os"
	"strconv"
)

var (
	host             = getEnvOrDefault("DATABASE_HOST", "localhost")
	port, _          = strconv.Atoi(getEnvOrDefault("DATABASE_PORT", "5432"))
	user             = getEnvOrDefault("DATABASE_USER", "turing")
	password         = getEnvOrDefault("DATABASE_PASSWORD", "turing")
	database         = getEnvOrDefault("DATABASE_NAME", "turing")
	migrationsFolder = "../../db-migrations"
)

func getEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
