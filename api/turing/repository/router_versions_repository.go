package repository

import (
	"gorm.io/gorm"

	"github.com/caraml-dev/turing/api/turing/models"
)

// routerVersionsRepository implements service.RouterVersionsRepository
type routerVersionsRepository struct {
	db *gorm.DB
}

func NewRouterVersionsRepository(db *gorm.DB) *routerVersionsRepository {
	return &routerVersionsRepository{
		db: db,
	}
}

func (rvr *routerVersionsRepository) query() *gorm.DB {
	return rvr.db.
		Preload("Router").
		Preload("Enricher").
		Preload("Ensembler")
}

func (rvr *routerVersionsRepository) List(routerID models.ID) ([]*models.RouterVersion, error) {
	var routerVersions []*models.RouterVersion
	query := rvr.query().Where("router_id = ?", routerID).Find(&routerVersions)
	return routerVersions, query.Error
}

func (rvr *routerVersionsRepository) ListByStatus(
	routerID models.ID,
	status models.RouterVersionStatus,
) ([]*models.RouterVersion, error) {
	var routerVersions []*models.RouterVersion
	query := rvr.query().
		Where("router_id = ?", routerID).
		Where("status = ?", status).
		Find(&routerVersions)
	return routerVersions, query.Error
}

func (rvr *routerVersionsRepository) Save(routerVersion *models.RouterVersion) (*models.RouterVersion, error) {
	var err error
	tx := rvr.db.Begin()

	if routerVersion.Model.IsNew() {
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
	return rvr.FindByID(routerVersion.ID)
}

func (rvr *routerVersionsRepository) FindByID(routerVersionID models.ID) (*models.RouterVersion, error) {
	var routerVersion models.RouterVersion
	query := rvr.query().
		Where("router_versions.id = ?", routerVersionID).
		First(&routerVersion)
	return &routerVersion, query.Error
}

func (rvr *routerVersionsRepository) FindByRouterIDAndVersion(
	routerID models.ID,
	version uint,
) (*models.RouterVersion, error) {
	var routerVersion models.RouterVersion
	query := rvr.query().
		Where("router_id = ?", routerID).
		Where("version = ?", version).
		First(&routerVersion)
	return &routerVersion, query.Error
}

func (rvr *routerVersionsRepository) FindLatestVersion(
	routerID models.ID,
) (*models.RouterVersion, error) {
	var routerVersion models.RouterVersion
	query := rvr.query().
		Where("router_id = ?", routerID).
		Order("version desc").
		First(&routerVersion)
	return &routerVersion, query.Error
}

func (rvr *routerVersionsRepository) Delete(routerVersion *models.RouterVersion) error {
	tx := rvr.db.Begin()
	tx.Delete(routerVersion)
	if routerVersion.Enricher != nil {
		tx.Delete(routerVersion.Enricher)
	}
	if routerVersion.Ensembler != nil {
		tx.Delete(routerVersion.Ensembler)
	}
	return tx.Commit().Error
}
