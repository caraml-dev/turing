package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	openapi "github.com/caraml-dev/turing/api/turing/generated"
)

// EnsemblingJob holds the information required for an ensembling job to be done asynchronously
type EnsemblingJob struct {
	Model
	Name            string       `json:"name"`
	EnsemblerID     ID           `json:"ensembler_id" validate:"required"`
	ProjectID       ID           `json:"project_id"`
	EnvironmentName string       `json:"environment_name"`
	InfraConfig     *InfraConfig `json:"infra_config" validate:"required"`
	JobConfig       *JobConfig   `json:"job_config" validate:"required"`
	RetryCount      int          `json:"-" gorm:"default:0"`
	Status          Status       `json:"status" gorm:"default:pending"`
	Error           string       `json:"error"`
	MonitoringURL   string       `json:"monitoring_url" gorm:"-"`
	RunID           int          `json:"-"`
}

// BeforeCreate sets the ensembling job name and run_id before creating
func (job *EnsemblingJob) BeforeCreate(tx *gorm.DB) error {
	if job.Model.IsNew() {
		// Only update fields if the record is new.
		// See: https://github.com/go-gorm/gorm/issues/4553
		var latest EnsemblingJob
		err := tx.Select("ensembling_jobs.run_id").
			Where("ensembler_id = ?", job.EnsemblerID).
			Order("run_id desc").
			FirstOrInit(&latest, &EnsemblingJob{RunID: 0}).
			Error

		if err != nil {
			return err
		}
		job.RunID = latest.RunID + 1
		job.Name = fmt.Sprintf("%s-%d", *job.InfraConfig.EnsemblerName, job.RunID)

	}

	return nil
}

// JobConfig stores the infra and ensembler config
type JobConfig openapi.EnsemblerConfig

// Value returns json value, implement driver.Valuer interface
func (c JobConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan scans value into Jsonb, implements sql.Scanner interface
func (c *JobConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &c)
}

// InfraConfig stores the infrastructure related configurations required.
type InfraConfig struct {
	openapi.EnsemblerInfraConfig
}

// Value returns json value, implement driver.Valuer interface
func (r *InfraConfig) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan scans value into Jsonb, implements sql.Scanner interface
func (r *InfraConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &r)
}

// Status is the state of the finite machine ensembling job.
// Possible statuses:
// JobPending --▶ JobFailedSubmission
//
//	|
//	|
//	|
//	|
//
// JobBuildingImage --▶ JobFailedBuildImage
//
//	|
//	|
//	|
//	|
//	▼
//
// JobRunning --▶ JobFailed
//
//	|
//	|
//	|--▶ JobTerminating --▶ JobTerminated
//	|
//	|
//	▼
//
// JobCompleted
type Status string

const (
	// JobPending is when the job has just been introduced.
	JobPending Status = "pending"
	// JobBuildingImage is when the job is builing a OCI image.
	JobBuildingImage Status = "building"
	// JobRunning is when the job has been picked up and running.
	JobRunning Status = "running"
	// JobTerminating is when the job has been requested to terminate.
	JobTerminating Status = "terminating"
	// JobTerminated is when the job has stopped. This is a terminal state.
	JobTerminated Status = "terminated"
	// JobCompleted is when the job has successfully completed. This is a terminal state.
	JobCompleted Status = "completed"
	// JobFailed is when the job has failed. This is a terminal state.
	JobFailed Status = "failed"
	// JobFailedSubmission is when the job has failed to submit. This is a terminal state.
	JobFailedSubmission Status = "failed_submission"
	// JobFailedBuildImage is when the job has failed to build an ensembling image.
	JobFailedBuildImage Status = "failed_building"
)

// IsTerminal checks if the job has reached a final state.
func (s Status) IsTerminal() bool {
	return s == JobTerminated || s == JobFailedSubmission ||
		s == JobFailed || s == JobCompleted || s == JobFailedBuildImage
}

// IsSuccessful checks if the ensembling job has completed.
func (s Status) IsSuccessful() bool {
	return s == JobCompleted
}
