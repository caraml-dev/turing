package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
)

const (
	sparkHomeFolder string = "/home/spark"
)

// EnsemblingJobFindByIDOptions contains the options allowed when finding ensembling jobs.
type EnsemblingJobFindByIDOptions struct {
	ProjectID *models.ID
}

// EnsemblingJobListOptions holds query parameters for EnsemblersService.List method.
type EnsemblingJobListOptions struct {
	PaginationOptions
	ProjectID *models.ID
	Status    *models.Status
}

// EnsemblingJobService is the data access object for the EnsemblingJob from the db.
type EnsemblingJobService interface {
	Save(ensembler *models.EnsemblingJob) error
	FindByID(id models.ID, options EnsemblingJobFindByIDOptions) (*models.EnsemblingJob, error)
	List(options EnsemblingJobListOptions) (*PaginatedResults, error)
	CreateEnsemblingJob(
		request *models.EnsemblingJob,
		projectID models.ID,
		ensembler *models.PyFuncEnsembler,
	) (*models.EnsemblingJob, error)
}

// NewEnsemblingJobService creates a new ensembling job service
func NewEnsemblingJobService(db *gorm.DB, defaultEnvironment string) EnsemblingJobService {
	return &ensemblingJobService{
		db:                 db,
		defaultEnvironment: defaultEnvironment,
	}
}

type ensemblingJobService struct {
	db                 *gorm.DB
	defaultEnvironment string
}

// Save the given router to the db. Updates the existing record if already exists
func (s *ensemblingJobService) Save(ensemblingJob *models.EnsemblingJob) error {
	return s.db.Save(ensemblingJob).Error
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
		return &ensemblingJob, err
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

	if options.Status != nil {
		query = query.Where("status = ?", options.Status)
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
	return fmt.Sprintf("%s: %s", ensemblerName, time.Now().Format(time.RFC3339))
}

func getEnsemblerDirectory(ensembler *models.PyFuncEnsembler) string {
	// Ensembler URI will be a local directory
	// Dockerfile will build copy the artifact into the local directory.
	// See engines/batch-ensembler/app.Dockerfile
	splitURI := strings.Split(ensembler.ArtifactURI, "/")
	return fmt.Sprintf(
		"%s/%s",
		sparkHomeFolder,
		splitURI[len(splitURI)-1],
	)
}

// CreateEnsemblingJob creates an ensembling job.
func (s *ensemblingJobService) CreateEnsemblingJob(
	request *models.EnsemblingJob,
	projectID models.ID,
	ensembler *models.PyFuncEnsembler,
) (*models.EnsemblingJob, error) {
	request.ProjectID = projectID
	request.EnvironmentName = s.defaultEnvironment

	// Populate name if the user does not define a name for the job
	if request.Name == "" {
		request.Name = generateDefaultJobName(ensembler.Name)
	}

	request.JobConfig.JobConfig.Spec.Ensembler.Uri = getEnsemblerDirectory(ensembler)
	request.InfraConfig.ArtifactURI = ensembler.ArtifactURI
	request.InfraConfig.EnsemblerName = ensembler.Name

	// Save ensembling job
	err := s.Save(request)
	if err != nil {
		return nil, err
	}

	return request, nil
}
