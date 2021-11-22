//go:build integration
// +build integration

package database

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/db"
	"github.com/gojek/turing/api/turing/log"
	"github.com/jinzhu/gorm"
)

func connectionString(db string) string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", host, port, user, db, password)
}

func create(conn *sql.DB, dbName string) (*sql.DB, error) {
	if _, err := conn.Exec("CREATE DATABASE " + dbName); err != nil {
		return nil, err
	} else if testDb, err := sql.Open("postgres", connectionString(dbName)); err != nil {
		if _, err := conn.Exec("DROP DATABASE " + dbName); err != nil {
			log.Fatalf("Failed to cleanup integration test database: \n%s", err)
		}
		return nil, err
	} else {
		return testDb, nil
	}
}

// CreateTestDatabase connects to test postgreSQL instance (either local or the one
// at CI environment) and creates a new database with an up-to-date schema
func CreateTestDatabase() (*gorm.DB, func(), error) {
	testDbName := fmt.Sprintf("mlp_id_%d", time.Now().UnixNano())

	connStr := connectionString(database)
	log.Infof("connecting to test db: %s", connStr)
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, nil, err
	}

	testDb, err := create(conn, testDbName)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		if err := testDb.Close(); err != nil {
			log.Fatalf("Failed to close connection to integration test database: \n%s", err)
		} else if _, err := conn.Exec("DROP DATABASE " + testDbName); err != nil {
			log.Fatalf("Failed to cleanup integration test database: \n%s", err)
		} else if err = conn.Close(); err != nil {
			log.Fatalf("Failed to close database: \n%s", err)
		}
	}

	dbCfg := &config.DatabaseConfig{
		Host:             host,
		Port:             port,
		User:             user,
		Password:         password,
		Database:         testDbName,
		MigrationsFolder: migrationsFolder,
	}
	if err = db.Migrate(dbCfg); err != nil {
		cleanup()
		return nil, nil, err
	} else if gormDb, err := gorm.Open("postgres", testDb); err != nil {
		cleanup()
		return nil, nil, err
	} else {
		return gormDb, cleanup, nil
	}
}

// WithTestDatabase handles the lifecycle of the database creation/migration/destruction
// for a test case/suite.
func WithTestDatabase(t *testing.T, test func(t *testing.T, db *gorm.DB)) {
	if testDb, cleanupFn, err := CreateTestDatabase(); err != nil {
		t.Fatalf("Fail to create an integration test database: \n%s", err)
	} else {
		test(t, testDb)
		cleanupFn()
	}
}
