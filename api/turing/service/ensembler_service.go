package service

import (
	"math"

	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
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
	var count int
	done := make(chan bool, 1)

	query := service.db
	if options.ProjectID != nil {
		query = query.Where("project_id = ?", options.ProjectID)
	}

	if options.EnsemblerType != nil {
		query = query.Where("type = ?", options.EnsemblerType)
	}

	go func() {
		query.Model(&results).Count(&count)
		done <- true
	}()

	result := query.
		Scopes(PaginationScope(options.PaginationOptions)).
		Find(&results)
	<-done

	page := 1
	if options.Page != nil {
		page = int(math.Max(1, float64(*options.Page)))
	}
	totalPages := 1
	if options.PageSize != nil {
		totalPages = int(math.Ceil(float64(count) / float64(*options.PageSize)))
	}

	if err := result.Error; err != nil {
		return nil, err
	}
	return &PaginatedResults{
		Results: results,
		Paging: Paging{
			Total: count,
			Page:  page,
			Pages: totalPages,
		},
	}, nil
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
