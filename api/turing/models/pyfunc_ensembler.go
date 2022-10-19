package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type EnsemblerLike interface {
	GetID() ID
	GetProjectID() ID
	GetType() EnsemblerType
	GetName() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time

	SetProjectID(id ID)
	Patch(other EnsemblerLike) error
}

func EnsemblerTable(ensembler EnsemblerLike) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		switch ensembler.GetType() {
		case EnsemblerPyFuncType:
			return tx.Table("pyfunc_ensemblers")
		default:
			return tx.Table("ensemblers")
		}
	}
}

type GenericEnsembler struct {
	Model
	// ProjectID id of the project this ensembler belongs to,
	// as retrieved from the MLP API.
	ProjectID ID `json:"project_id" gorm:"column:project_id"`

	Type EnsemblerType `json:"type" gorm:"column:type" validate:"required,oneof=pyfunc"`

	Name string `json:"name" gorm:"column:name" validate:"required,hostname_rfc1123,lte=20,gte=3"`
}

func (e *GenericEnsembler) GetProjectID() ID {
	return e.ProjectID
}

func (e *GenericEnsembler) SetProjectID(id ID) {
	e.ProjectID = id
}

func (e *GenericEnsembler) GetType() EnsemblerType {
	return e.Type
}

func (e *GenericEnsembler) GetName() string {
	return e.Name
}

func (*GenericEnsembler) TableName() string {
	return "ensemblers"
}

func (e *GenericEnsembler) Patch(other EnsemblerLike) error {
	e.Name = other.GetName()
	return nil
}

func (e *GenericEnsembler) Instance() EnsemblerLike {
	switch e.GetType() {
	case EnsemblerPyFuncType:
		return &PyFuncEnsembler{}
	default:
		return e
	}
}

type PyFuncEnsembler struct {
	*GenericEnsembler

	MlflowURL string `json:"mlflow_url" gorm:"-"`

	ExperimentID ID `json:"mlflow_experiment_id" gorm:"column:mlflow_experiment_id"`

	RunID string `json:"mlflow_run_id" gorm:"column:mlflow_run_id"`

	ArtifactURI string `json:"artifact_uri" gorm:"column:artifact_uri"`

	PythonVersion string `json:"python_version" gorm:"column:python_version"`
}

func (e *PyFuncEnsembler) BeforeCreate(*gorm.DB) error {
	e.Type = EnsemblerPyFuncType
	return nil
}

func (*PyFuncEnsembler) GetType() EnsemblerType {
	return EnsemblerPyFuncType
}

func (e *PyFuncEnsembler) Patch(other EnsemblerLike) error {
	otherPyfunc, ok := other.(*PyFuncEnsembler)
	if !ok {
		return fmt.Errorf("update must be of the same type as as the receiver")
	}
	if err := e.GenericEnsembler.Patch(otherPyfunc); err != nil {
		return err
	}

	e.MlflowURL = otherPyfunc.MlflowURL
	e.ExperimentID = otherPyfunc.ExperimentID
	e.RunID = otherPyfunc.RunID
	e.ArtifactURI = otherPyfunc.ArtifactURI
	e.PythonVersion = otherPyfunc.PythonVersion

	return nil
}
