package models

type PyFuncEnsembler struct {
	Model
	// ProjectID id of the project this ensembler belongs to,
	// as retrieved from the MLP API.
	ProjectID ID `json:"project_id"`

	Name string `json:"name" validate:"required,min=3,max=50"`

	MlflowURL string `json:"mlflow_url" gorm:"-"`

	ExperimentID ID `json:"mlflow_experiment_id" gorm:"column:mlflow_experiment_id"`

	RunID string `json:"mlflow_run_id" gorm:"column:mlflow_run_id"`

	ArtifactURI string `json:"artifact_uri" gorm:"artifact_uri"`
}
