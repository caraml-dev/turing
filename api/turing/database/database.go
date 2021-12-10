package database

import (
	"fmt"

	"github.com/gojek/turing/api/turing/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jinzhu/gorm"

	// required for gomigrate
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// enable postgres dialect
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// InitDB initialises a database connection as well as runs the migration scripts.
// It is important to close the database after using it by calling defer db.Close()
func InitDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// Migrate
	err := migrateDB(cfg)
	if err != nil {
		return nil, err
	}

	// Init db
	db, err := gorm.Open(
		"postgres",
		connectionString(cfg),
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to start Gorm DB: %s", err)
	}

	db.LogMode(false)
	return db, nil
}

func connectionString(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Database,
		cfg.Password,
	)
}

// Migrate migrates the database, returns the Migrate object.
func migrateDB(cfg *config.DatabaseConfig) error {
	// run db migrations
	m, err := migrate.New(
		fmt.Sprintf("file://%s", cfg.MigrationsFolder),
		fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
		),
	)
	if err != nil {
		return fmt.Errorf("Failed to open migrations folder: %s", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Failed to run migrations: %s", err)
	}
	if sourceErr, dbErr := m.Close(); sourceErr != nil {
		return fmt.Errorf("Failed to close source after migration")
	} else if dbErr != nil {
		return fmt.Errorf("Failed to close database after migration")
	}

	return nil
}
