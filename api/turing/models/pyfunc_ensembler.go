package models

import (
	"github.com/jinzhu/gorm"
)

type EnsemblerLike interface {
	ProjectID() ID
	Type() EnsemblerType
	Name() string
}

func EnsemblerTable(ensembler EnsemblerLike) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		switch ensembler.Type() {
		case EnsemblerTypePyFunc:
			return tx.Table("pyfunc_ensemblers")
		default:
			return tx.Table("ensemblers")
		}
	}
}

type GenericEnsembler struct {
	Model
	// TProjectID id of the project this ensembler belongs to,
	// as retrieved from the MLP API.
	TProjectID ID `json:"project_id" gorm:"column:project_id"`

	TType EnsemblerType `json:"type" gorm:"column:type"`

	TName string `json:"name" gorm:"column:name" validate:"required,min=3,max=50"`
}

func (e *GenericEnsembler) ProjectID() ID {
	return e.TProjectID
}

func (e *GenericEnsembler) Type() EnsemblerType {
	return e.TType
}

func (e *GenericEnsembler) Name() string {
	return e.TName
}

func (*GenericEnsembler) TableName() string {
	return "ensemblers"
}

func (e *GenericEnsembler) Instance() EnsemblerLike {
	switch e.Type() {
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

	ArtifactURI string `json:"artifact_uri" gorm:"column:artifact_uri"`
}

func (*PyFuncEnsembler) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("type", EnsemblerTypePyFunc)
}

func (*PyFuncEnsembler) Type() EnsemblerType {
	return EnsemblerTypePyFunc
}
