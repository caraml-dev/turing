package service

import (
	"errors"

	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
)

// RouterVersionsService is the data access object for RouterVersions from the db.
type RouterVersionsService interface {
	// List all RouterVersions associated with the given routerID
	ListRouterVersions(routerID uint) ([]*models.RouterVersion, error)
	// Lists the RouterVersions for the given Router matching the given status.
	ListRouterVersionsWithStatus(routerID uint, status models.RouterVersionStatus) ([]*models.RouterVersion, error)
	// Save the given RouterVersion to the db. Updates the existing record if already exists.
	Save(routerVersion *models.RouterVersion) (*models.RouterVersion, error)
	// Finds the RouterVersion matching the given id.
	FindByID(routerVersionID uint) (*models.RouterVersion, error)
	// Finds the RouterVersion for the given Router matching the given version.
	FindByRouterIDAndVersion(routerID uint, version uint) (*models.RouterVersion, error)
	// Finds the latest RouterVersion for the given Router matching the given version.
	FindLatestVersionbyRouterID(routerID uint) (*models.RouterVersion, error)
	// Deletes the given RouterVersion from the db. This method deletes all child objects (enricher, ensembler).
	Delete(routerVersion *models.RouterVersion) error
}

func NewRouterVersionsService(db *gorm.DB) RouterVersionsService {
	return &routerVersionsService{db: db}
}

type routerVersionsService struct {
	db *gorm.DB
}

func (service *routerVersionsService) query() *gorm.DB {
	return service.db.
		Preload("Router").
		Preload("Enricher").
		Preload("Ensembler").
		Joins("LEFT JOIN enrichers on enrichers.id = router_versions.enricher_id").
		Joins("LEFT JOIN ensemblers on ensemblers.id = router_versions.enricher_id").
		Select("router_versions.*")
}

func (service *routerVersionsService) ListRouterVersions(routerID uint) ([]*models.RouterVersion, error) {
	var routerVersions []*models.RouterVersion
	query := service.query().Where("router_id = ?", routerID).Find(&routerVersions)
	return routerVersions, query.Error
}

func (service *routerVersionsService) ListRouterVersionsWithStatus(
	routerID uint,
	status models.RouterVersionStatus,
) ([]*models.RouterVersion, error) {
	var routerVersions []*models.RouterVersion
	query := service.query().
		Where("router_id = ?", routerID).
		Where("status = ?", status).
		Find(&routerVersions)
	return routerVersions, query.Error
}

func (service *routerVersionsService) Save(routerVersion *models.RouterVersion) (*models.RouterVersion, error) {
	var err error
	tx := service.db.Begin()
	if service.db.NewRecord(routerVersion) {
		if routerVersion.Ensembler != nil {
			if err := tx.Create(routerVersion.Ensembler).Error; err != nil {
				return nil, err
			}
			routerVersion.SetEnsemblerID(routerVersion.Ensembler.ID)
		}
		if routerVersion.Enricher != nil {
			if err := tx.Create(routerVersion.Enricher).Error; err != nil {
				return nil, err
			}
			routerVersion.SetEnricherID(routerVersion.Enricher.ID)
		}
		err = tx.Create(routerVersion).Error
	} else {
		// We don't allow ensembler and enricher updates. Changes to those elements
		// will always result in a new version being spawned.
		err = tx.Save(routerVersion).Error
	}

	if err != nil {
		return nil, err
	}
	tx.Commit()
	return service.FindByID(routerVersion.ID)
}

func (service *routerVersionsService) FindByID(
	routerVersionID uint,
) (*models.RouterVersion, error) {
	var routerVersion models.RouterVersion
	query := service.query().
		Where("router_versions.id = ?", routerVersionID).
		First(&routerVersion)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &routerVersion, nil
}

func (service *routerVersionsService) FindByRouterIDAndVersion(
	routerID uint,
	version uint,
) (*models.RouterVersion, error) {
	var routerVersion models.RouterVersion
	query := service.query().
		Where("router_id = ?", routerID).
		Where("version = ?", version).
		First(&routerVersion)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &routerVersion, nil
}

func (service *routerVersionsService) FindLatestVersionbyRouterID(
	routerID uint,
) (*models.RouterVersion, error) {
	var routerVersion models.RouterVersion
	query := service.query().
		Where("router_id = ?", routerID).
		Order("version desc").
		First(&routerVersion)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &routerVersion, nil
}

func (service *routerVersionsService) Delete(routerVersion *models.RouterVersion) error {
	if routerVersion.ID == 0 {
		return errors.New("router version must have valid primary key to be deleted")
	}

	var upstreamRefs int
	service.db.Table("routers").
		Where("routers.curr_router_version_id = ?", routerVersion.ID).
		Count(&upstreamRefs)
	if upstreamRefs != 0 {
		return errors.New("unable to delete router version - there exists a router that is currently using this version")
	}

	tx := service.db.Begin()
	tx.Delete(routerVersion)
	if routerVersion.Enricher != nil {
		tx.Delete(routerVersion.Enricher)
	}
	if routerVersion.Ensembler != nil {
		tx.Delete(routerVersion.Ensembler)
	}
	return tx.Commit().Error
}
