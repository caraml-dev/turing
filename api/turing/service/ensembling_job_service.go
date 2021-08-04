package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/gojek/turing/api/turing/config"
	openapi "github.com/gojek/turing/api/turing/generated"
	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
)

const (
	// SparkHomeFolder is the home folder of the spark user in the Docker container
	// used in engines/batch-ensembler/Dockerfile
	SparkHomeFolder = "/home/spark"
	// EnsemblerFolder is the folder created by the Turing SDK that contains
	// the ensembler dependencies and pickled Python files.
	EnsemblerFolder = "ensembler"
)

// EnsemblingJobFindByIDOptions contains the options allowed when finding ensembling jobs.
type EnsemblingJobFindByIDOptions struct {
	ProjectID *models.ID
}

// EnsemblingJobListOptions holds query parameters for EnsemblersService.List method.
type EnsemblingJobListOptions struct {
	PaginationOptions
	ProjectID          *models.ID      `schema:"project_id" validate:"required"`
	EnsemblerID        *models.ID      `schema:"ensembler_id"`
	Statuses           []models.Status `schema:"status"`
	Search             *string         `schema:"search"`
	RetryCountLessThan *int            `schema:"-"`
	UpdatedAtBefore    *time.Time      `schema:"-"`
}

// EnsemblingJobService is the data access object for the EnsemblingJob from the db.
type EnsemblingJobService interface {
	Save(ensemblingJob *models.EnsemblingJob) error
	Delete(ensemblingJob *models.EnsemblingJob) error
	FindByID(id models.ID, options EnsemblingJobFindByIDOptions) (*models.EnsemblingJob, error)
	List(options EnsemblingJobListOptions) (*PaginatedResults, error)
	CreateEnsemblingJob(
		job *models.EnsemblingJob,
		projectID models.ID,
		ensembler *models.PyFuncEnsembler,
	) (*models.EnsemblingJob, error)
	MarkEnsemblingJobForTermination(ensemblingJob *models.EnsemblingJob) error
}

// NewEnsemblingJobService creates a new ensembling job service
func NewEnsemblingJobService(
	db *gorm.DB,
	defaultEnvironment string,
	defaultConfig config.DefaultEnsemblingJobConfigurations,
) EnsemblingJobService {
	return &ensemblingJobService{
		db:                 db,
		defaultEnvironment: defaultEnvironment,
		defaultConfig:      defaultConfig,
	}
}

type ensemblingJobService struct {
	db                 *gorm.DB
	defaultEnvironment string
	defaultConfig      config.DefaultEnsemblingJobConfigurations
}

// Save the given router to the db. Updates the existing record if already exists
func (s *ensemblingJobService) Save(ensemblingJob *models.EnsemblingJob) error {
	return s.db.Save(ensemblingJob).Error
}

func (s *ensemblingJobService) Delete(ensemblingJob *models.EnsemblingJob) error {
	return s.db.Delete(ensemblingJob).Error
}

func (s *ensemblingJobService) FindByID(
	id models.ID,
	options EnsemblingJobFindByIDOptions,
) (*models.EnsemblingJob, error) {
	query := s.db.Where("id = ?", id)

	if options.ProjectID != nil {
		query = query.Where("project_id = ?", options.ProjectID)
	}

	var ensemblingJob models.EnsemblingJob
	result := query.First(&ensemblingJob)

	if err := result.Error; err != nil {
		return nil, err
	}

	return &ensemblingJob, nil
}

func (s *ensemblingJobService) List(options EnsemblingJobListOptions) (*PaginatedResults, error) {
	var results []*models.EnsemblingJob
	var count int
	done := make(chan bool, 1)

	query := s.db
	if options.ProjectID != nil {
		query = query.Where("project_id = ?", options.ProjectID)
	}

	if options.EnsemblerID != nil {
		query = query.Where("ensembler_id = ?", options.EnsemblerID)
	}

	if options.Search != nil && len(*options.Search) > 0 {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", *options.Search))
	}

	if options.Statuses != nil {
		query = query.Where("status IN (?)", options.Statuses)
	}

	if options.RetryCountLessThan != nil {
		query = query.Where("retry_count < ?", options.RetryCountLessThan)
	}

	if options.UpdatedAtBefore != nil {
		query = query.Where("updated_at < ?", options.UpdatedAtBefore)
	}

	go func() {
		query.Model(&results).Count(&count)
		done <- true
	}()

	result := query.
		Scopes(PaginationScope(options.PaginationOptions)).
		Find(&results)
	<-done

	if err := result.Error; err != nil {
		return nil, err
	}

	paginatedResults := createPaginatedResults(options.PaginationOptions, count, results)
	return paginatedResults, nil
}

func generateDefaultJobName(ensemblerName string) string {
	t := time.Now().Unix()
	return fmt.Sprintf("%s-%d", ensemblerName, t)
}

func getEnsemblerDirectory(ensembler *models.PyFuncEnsembler) string {
	// Ensembler URI will be a local directory
	// Dockerfile will build copy the artifact into the local directory.
	// See engines/batch-ensembler/app.Dockerfile
	splitURI := strings.Split(ensembler.ArtifactURI, "/")
	return fmt.Sprintf(
		"%s/%s/ensembler",
		SparkHomeFolder,
		splitURI[len(splitURI)-1],
	)
}

// CreateEnsemblingJob creates an ensembling job.
func (s *ensemblingJobService) CreateEnsemblingJob(
	job *models.EnsemblingJob,
	projectID models.ID,
	ensembler *models.PyFuncEnsembler,
) (*models.EnsemblingJob, error) {
	job.ProjectID = projectID
	job.EnvironmentName = s.defaultEnvironment

	// Populate name if the user does not define a name for the job
	if job.Name == "" {
		job.Name = generateDefaultJobName(ensembler.Name)
	}

	job.JobConfig.Spec.Ensembler.Uri = getEnsemblerDirectory(ensembler)
	job.InfraConfig.ArtifactURI = ensembler.ArtifactURI
	job.InfraConfig.EnsemblerName = ensembler.Name

	s.mergeDefaultConfigurations(job)

	// Save ensembling job
	err := s.Save(job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *ensemblingJobService) MarkEnsemblingJobForTermination(job *models.EnsemblingJob) error {
	job.Status = models.JobTerminating
	return s.Save(job)
}

func (s *ensemblingJobService) mergeDefaultConfigurations(job *models.EnsemblingJob) {
	if job.JobConfig.Metadata == nil {
		job.JobConfig.Metadata = &openapi.EnsemblingJobMeta{
			Annotations: make(map[string]string),
		}
	}

	// Only apply default if key does not exist, we should respect the users annotation override.
	for key, value := range s.defaultConfig.SparkConfigAnnotations {
		if _, ok := job.JobConfig.Metadata.Annotations[key]; !ok {
			job.JobConfig.Metadata.Annotations[key] = value
		}
	}

	if job.InfraConfig.Resources == nil {
		configCopy := s.defaultConfig.BatchEnsemblingJobResources
		job.InfraConfig.Resources = &configCopy
		return
	}

	if job.InfraConfig.Resources.DriverCpuRequest == nil {
		job.InfraConfig.Resources.DriverCpuRequest = s.defaultConfig.BatchEnsemblingJobResources.DriverCpuRequest
	}

	if job.InfraConfig.Resources.DriverMemoryRequest == nil {
		job.InfraConfig.Resources.DriverMemoryRequest = s.defaultConfig.BatchEnsemblingJobResources.DriverMemoryRequest
	}

	if job.InfraConfig.Resources.ExecutorReplica == nil {
		job.InfraConfig.Resources.ExecutorReplica = s.defaultConfig.BatchEnsemblingJobResources.ExecutorReplica
	}

	if job.InfraConfig.Resources.ExecutorCpuRequest == nil {
		job.InfraConfig.Resources.ExecutorCpuRequest = s.defaultConfig.BatchEnsemblingJobResources.ExecutorCpuRequest
	}

	if job.InfraConfig.Resources.ExecutorMemoryRequest == nil {
		job.InfraConfig.Resources.ExecutorMemoryRequest = s.defaultConfig.BatchEnsemblingJobResources.ExecutorMemoryRequest
	}
}
