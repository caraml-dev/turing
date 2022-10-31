package service

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/caraml-dev/turing/api/turing/models"
)

// EnsemblersService is the data access object for the Ensemblers from the db.
type EnsemblersService interface {
	// FindByID Find an ensembler matching the given id and options
	FindByID(id models.ID, options EnsemblersFindByIDOptions) (models.EnsemblerLike, error)
	// List ensemblers
	List(options EnsemblersListOptions) (*PaginatedResults, error)
	// Save the given router to the db. Updates the existing record if already exists
	Save(ensembler models.EnsemblerLike) (models.EnsemblerLike, error)
}

type EnsemblersFindByIDOptions struct {
	ProjectID *models.ID
}

// EnsemblersListOptions holds query parameters for EnsemblersService.List method
type EnsemblersListOptions struct {
	PaginationOptions
	ProjectID     *models.ID            `schema:"project_id" validate:"required"`
	Search        *string               `schema:"search"`
	EnsemblerType *models.EnsemblerType `schema:"type" validate:"omitempty,oneof=pyfunc docker"`
}

// NewEnsemblersService creates a new ensemblers service
func NewEnsemblersService(db *gorm.DB) EnsemblersService {
	return &ensemblersService{db: db}
}

type ensemblersService struct {
	db *gorm.DB
}

func (service *ensemblersService) FindByID(
	id models.ID,
	options EnsemblersFindByIDOptions,
) (models.EnsemblerLike, error) {
	var ensembler models.GenericEnsembler
	query := service.db.Where("id = ?", id)

	if options.ProjectID != nil {
		query = query.Where("project_id = ?", options.ProjectID)
	}

	result := query.First(&ensembler)
	if err := result.Error; err != nil {
		return nil, err
	}
	instance := ensembler.Instance()
	result = query.Scopes(models.EnsemblerTable(&ensembler)).First(instance)
	if err := result.Error; err != nil {
		return nil, err
	}
	return instance, nil
}

func (service *ensemblersService) List(options EnsemblersListOptions) (*PaginatedResults, error) {
	var results []*models.GenericEnsembler
	var count int64

	query := service.db
	if options.ProjectID != nil {
		query = query.Where("project_id = ?", options.ProjectID)
	}

	if options.Search != nil && len(*options.Search) > 0 {
		query = query.Where("name ilike ?", fmt.Sprintf("%%%s%%", *options.Search))
	}

	if options.EnsemblerType != nil {
		query = query.Where("type = ?", options.EnsemblerType)
	}

	query.Model(&results).Count(&count)
	result := query.
		Scopes(PaginationScope(options.PaginationOptions)).
		Find(&results)

	if err := result.Error; err != nil {
		return nil, err
	}

	paginatedResults := createPaginatedResults(options.PaginationOptions, int(count), results)
	return paginatedResults, nil
}

func (service *ensemblersService) Save(ensembler models.EnsemblerLike) (models.EnsemblerLike, error) {
	result := service.db.
		Scopes(models.EnsemblerTable(ensembler)).
		Save(ensembler)

	if err := result.Error; err != nil {
		return nil, err
	}
	return service.FindByID(ensembler.GetID(), EnsemblersFindByIDOptions{})
}
