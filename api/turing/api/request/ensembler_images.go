package request

import "github.com/caraml-dev/turing/api/turing/models"

type BuildEnsemblerImageRequest struct {
	RunnerType models.EnsemblerRunnerType `json:"runner_type" validate:"required"`
}
