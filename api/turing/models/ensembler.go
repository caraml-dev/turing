package models

import (
	"database/sql/driver"
	"encoding/json"
)

// Ensembler contains the configuration for a post-processor for
// a Turing Router.
type Ensembler struct {
	Model
	Type           EnsemblerType            `json:"type" validate:"required"`
	StandardConfig *EnsemblerStandardConfig `json:"standard_config"` // Ensembler config when Type is "standard"
	DockerConfig   *EnsemblerDockerConfig   `json:"docker_config"`   // Ensembler config when Type is "docker"
}

// TableName returns the name of a table, where GORM should store/retrieve
// entities of this type. By default GORM uses the table name, that is
// a plural form of the type's name (i.e `Ensembler` -> `ensemblers`),
// and by implementing `TableName` method it is possible to override it.
func (*Ensembler) TableName() string {
	return "ensembler_configs"
}

type EnsemblerType string

const (
	EnsemblerStandardType EnsemblerType = "standard"
	EnsemblerDockerType   EnsemblerType = "docker"
	EnsemblerTypePyFunc   EnsemblerType = "pyfunc"
)

type EnsemblerStandardConfig struct {
	ExperimentMappings []ExperimentMapping `json:"experiment_mappings" validate:"required,dive"`
}

type EnsemblerDockerConfig struct {
	Image string `json:"image" validate:"required"`
	// Resource requests for ensembler container deployed
	ResourceRequest *ResourceRequest `json:"resource_request" validate:"required"`
	// URL path for the endpoint, e.g "/"
	Endpoint string `json:"endpoint" validate:"required"`
	// Request timeout in duration format e.g. 60s
	Timeout string `json:"timeout" validate:"required"`
	// Port number the container listens to for requests
	Port int `json:"port" validate:"required"`
	// Environment variables to set in the container
	Env EnvVars `json:"env" validate:"required"`
	// secret name in MLP containing service account key
	ServiceAccount string `json:"service_account"`
}

type ExperimentMapping struct {
	Experiment string `json:"experiment" validate:"required"` // Experiment name from the experiment engine
	Treatment  string `json:"treatment" validate:"required"`  // Treatment name for the experiment
	Route      string `json:"route" validate:"required"`      // Route ID to select for the experiment treatment
}

// Value implements sql.driver.Valuer interface so database tools like go-orm knows how to serialize the struct object
// for storage in the database
func (c EnsemblerStandardConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner interface so database tools like go-orm knows how to de-serialize the struct object
// from the database
func (c *EnsemblerStandardConfig) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &c)
}

// Value implements sql.driver.Valuer interface so database tools like go-orm knows how to serialize the struct object
// for storage in the database
func (c EnsemblerDockerConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements sql.Scanner interface so database tools like go-orm knows how to de-serialize the struct object
// from the database
func (c *EnsemblerDockerConfig) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &c)
}
