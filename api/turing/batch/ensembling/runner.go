package batchensembling

import (
	"github.com/gojek/turing/api/turing/batch"
	"github.com/gojek/turing/api/turing/imagebuilder"
	"github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
)

type ensemblingJobRunner struct {
	ensemblingJobService service.EnsemblingJobService
	mlpService           service.MLPService
	imageBuilder         imagebuilder.ImageBuilder
	injectGojekLabels    bool
	environment          string
	batchSize            int
}

var (
	labelTeamName         = "gojek.com/team"
	labelStreamName       = "gojek.com/stream"
	labelAppName          = "gojek.com/app"
	labelEnvironment      = "gojek.com/environment"
	labelOrchestratorName = "gojek.com/orchestrator"
	valueOrchestratorName = "turing"
)

// NewBatchEnsemblingJobRunner creates a new batch ensembling job
// This service controls the orchestration of batch ensembling jobs.
func NewBatchEnsemblingJobRunner(
	ensemblingJobService service.EnsemblingJobService,
	mlpService service.MLPService,
	imageBuilder imagebuilder.ImageBuilder,
	injectGojekLabels bool,
	environment string,
	batchSize int,
) batch.JobRunner {
	return &ensemblingJobRunner{
		ensemblingJobService: ensemblingJobService,
		mlpService:           mlpService,
		imageBuilder:         imageBuilder,
		injectGojekLabels:    injectGojekLabels,
		environment:          environment,
		batchSize:            batchSize,
	}
}

func (r *ensemblingJobRunner) Run() {
	r.processNewEnsemblingJobs()
	r.updateEnsemblingJobStatus()
}

func (r *ensemblingJobRunner) processNewEnsemblingJobs() {
	pendingStatus := models.JobPending
	options := service.EnsemblingJobListOptions{
		PaginationOptions: service.PaginationOptions{
			Page:     testutils.NullableInt(1),
			PageSize: &r.batchSize,
		},
		Status: &pendingStatus,
	}
	ensemblingJobsPaginated, err := r.ensemblingJobService.List(options)
	if err != nil {
		// TODO: send metric to Prometheus
		log.Errorf("unable to query ensembling jobs", err)
	}
	ensemblingJobs := ensemblingJobsPaginated.Results.([]*models.EnsemblingJob)
	for _, ensemblingJob := range ensemblingJobs {
		// Build Image
		imageRef, imageBuildErr := r.buildImage(ensemblingJob)
		if imageBuildErr != nil {
			r.saveStatus(ensemblingJob, models.JobFailedBuildImage, imageBuildErr.Error())
			continue
		}

		// Submit to Kubernetes
		kubernetesError := r.submitToKubernetes(ensemblingJob, imageRef)
		if kubernetesError != nil {
			r.saveStatus(ensemblingJob, models.JobFailedSubmission, imageBuildErr.Error())
			continue
		}

		r.saveStatus(ensemblingJob, models.JobRunning, "")
	}
}

func (r *ensemblingJobRunner) saveStatus(
	ensemblingJob *models.EnsemblingJob,
	status models.Status,
	errorMessage string,
) {
	ensemblingJob.Status = status
	ensemblingJob.Error = errorMessage
	err := r.ensemblingJobService.Save(ensemblingJob)
	if err != nil {
		log.Errorf("unable to save ensembling job", err)
	}
}

func (r *ensemblingJobRunner) buildImage(ensemblingJob *models.EnsemblingJob) (string, error) {
	mlpProject, err := r.mlpService.GetProject(ensemblingJob.ProjectID)
	if err != nil {
		return "", err
	}

	buildLabels := make(map[string]string)
	if r.injectGojekLabels {
		buildLabels[labelTeamName] = mlpProject.Team
		buildLabels[labelStreamName] = mlpProject.Stream
		buildLabels[labelAppName] = ensemblingJob.InfraConfig.EnsemblerName
		buildLabels[labelEnvironment] = r.environment
		buildLabels[labelOrchestratorName] = valueOrchestratorName
	}

	request := imagebuilder.BuildImageRequest{
		ProjectName: mlpProject.Name,
		ModelName:   ensemblingJob.InfraConfig.EnsemblerName,
		VersionID:   int(ensemblingJob.EnsemblerID),
		ArtifactURI: ensemblingJob.InfraConfig.ArtifactURI,
		BuildLabels: buildLabels,
	}
	return r.imageBuilder.BuildImage(request)
}

func (r *ensemblingJobRunner) submitToKubernetes(ensemblingJob *models.EnsemblingJob, imageRef string) error {
	// TODO: Implement
	return nil
}

func (r *ensemblingJobRunner) updateEnsemblingJobStatus() {

}
