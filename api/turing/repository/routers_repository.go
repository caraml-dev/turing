package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/caraml-dev/turing/api/turing/models"
)

// RoutersRepository implements service.RoutersRepository
type RoutersRepository struct {
	db *gorm.DB
}

func NewRoutersRepository(db *gorm.DB) *RoutersRepository {
	return &RoutersRepository{
		db: db,
	}
}

func (rr *RoutersRepository) query() *gorm.DB {
	return rr.db.
		Preload("CurrRouterVersion").
		Preload("CurrRouterVersion.Enricher").
		Preload("CurrRouterVersion.Ensembler")
}

func (rr *RoutersRepository) CountRoutersByCurrentVersionID(routerVersionID models.ID) int64 {
	var upstreamRefs int64
	rr.db.Table("routers").
		Where("routers.curr_router_version_id = ?", routerVersionID).
		Count(&upstreamRefs)
	return upstreamRefs
}

func (rr *RoutersRepository) Save(router *models.Router) (*models.Router, error) {
	if err := rr.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(router).Error; err != nil {
		return nil, err
	}
	return rr.findByID(router.ID)
}

func (rr *RoutersRepository) findByID(routerID models.ID) (*models.Router, error) {
	var router models.Router
	query := rr.query().
		Where("routers.id = ?", routerID).
		First(&router)
	return &router, query.Error
}
