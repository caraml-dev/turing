package models

import (
	"database/sql/driver"
	"encoding/json"

	batchensembler "github.com/gojek/turing/engines/batch-ensembler/pkg/api/proto/v1"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

// EnsemblingJob holds the information required for an ensembling job to be done asynchronously
type EnsemblingJob struct {
	Model
	Name            string           `json:"name"`
	VersionID       ID               `json:"version_id"`
	EnsemblerID     ID               `json:"ensembler_id"`
	ProjectID       ID               `json:"project_id"`
	EnvironmentName string           `json:"environment_name"`
	InfraConfig     *InfraConfig     `json:"infra_config"`
	EnsemblerConfig *EnsemblerConfig `json:"ensembler_config"`
	Status          State            `json:"status" gorm:"default:pending"`
	Error           string           `json:"error"`
}

// EnsemblerConfig stores the infra and ensembler config
type EnsemblerConfig struct {
	EnsemblerConfig batchensembler.BatchEnsemblingJob `json:"ensembler_config"`
}

// UnmarshalJSON unmarshals the json into the proto message, used by the json.Marshaler interface
func (r *EnsemblerConfig) UnmarshalJSON(data []byte) error {
	return protojson.Unmarshal(data, &r.EnsemblerConfig)
}

// MarshalJSON unmarshals the json into the proto message, used by the json.Marshaler interface
func (r *EnsemblerConfig) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(&r.EnsemblerConfig)
}

// Value returns json value, implement driver.Valuer interface
func (r *EnsemblerConfig) Value() (driver.Value, error) {
	return protojson.Marshal(&r.EnsemblerConfig)
}

// Scan scans value into Jsonb, implements sql.Scanner interface
func (r *EnsemblerConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return protojson.Unmarshal(b, &r.EnsemblerConfig)
}

// InfraConfig stores the infrastructure related configurations required.
type InfraConfig struct {
	ServiceAccountName string                       `json:"service_account_name"`
	ImageRef           string                       `json:"image_ref"`
	Resources          *BatchEnsemblingJobResources `json:"resources"`
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

// BatchEnsemblingJobResources contains the resources delared to run the ensembling job.
type BatchEnsemblingJobResources struct {
	Requests *Resource `json:"requests"`
	Limits   *Resource `json:"limits"`
}

// Resource contains the Kubernetes resource request and limits
type Resource struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// State is the state of the finite machine ensembling job.
// Possible states:
// JobPending --▶ JobFailedSubmission
//     |
//     |--▶ JobFailedBuildImage
//     |
//     ▼
// JobRunning --▶ JobFailed
//     |
//     |
//     |--▶ JobTerminating --▶ JobTerminated
//     |
//     |
//     ▼
// JobCompleted
type State string

const (
	// JobPending is when the job has just been introduced.
	JobPending State = "pending"
	// JobRunning is when the job has been picked up and running.
	JobRunning State = "running"
	// JobTerminating is when the job has begun stopping.
	JobTerminating State = "terminating"
	// JobTerminated is when the job has stopped. This is a terminal state.
	JobTerminated State = "terminated"
	// JobCompleted is when the job has successfully completed. This is a terminal state.
	JobCompleted State = "completed"
	// JobFailed is when the job has failed. This is a terminal state.
	JobFailed State = "failed"
	// JobFailedSubmission is when the job has failed to submit. This is a terminal state.
	JobFailedSubmission State = "failed_submission"
	// JobFailedBuildImage is when the job has failed to build an ensembling image.
	JobFailedBuildImage State = "failed_building"
)

// IsTerminal checks if the job has reached a final state.
func (s State) IsTerminal() bool {
	return s == JobTerminated || s == JobFailedSubmission || s == JobFailed || s == JobCompleted || s == JobFailedBuildImage
}

// IsSuccessful checks if the ensembling job has completed.
func (s State) IsSuccessful() bool {
	return s == JobCompleted
}
