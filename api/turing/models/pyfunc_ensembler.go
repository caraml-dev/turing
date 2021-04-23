package models

import (
	"github.com/jinzhu/gorm"
)

type EnsemblerLike interface {
	Kind() EnsemblerType
}

func EnsemblerTable(ensembler EnsemblerLike) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		switch ensembler.Kind() {
		case EnsemblerTypePyFunc:
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
	ProjectID ID `json:"project_id"`

	Type EnsemblerType `json:"type"`

	Name string `json:"name" validate:"required,min=3,max=50"`
}

func (*GenericEnsembler) TableName() string {
	return "ensemblers"
}

func (e *GenericEnsembler) Kind() EnsemblerType {
	return e.Type
}

func (e *GenericEnsembler) Instance() EnsemblerLike {
	switch e.Kind() {
	case EnsemblerTypePyFunc:
		return &PyFuncEnsembler{}
	default:
		return nil
	}
}

type PyFuncEnsembler struct {
	*GenericEnsembler

	MlflowURL string `json:"mlflow_url" gorm:"-"`

	ExperimentID ID `json:"mlflow_experiment_id" gorm:"column:mlflow_experiment_id"`

	RunID string `json:"mlflow_run_id" gorm:"column:mlflow_run_id"`

	ArtifactURI string `json:"artifact_uri" gorm:"artifact_uri"`
}

func (*PyFuncEnsembler) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("type", EnsemblerTypePyFunc)
}

func (*PyFuncEnsembler) Kind() EnsemblerType {
	return EnsemblerTypePyFunc
}
