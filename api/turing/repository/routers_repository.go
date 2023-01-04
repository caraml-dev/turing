package repository

import (
	"gorm.io/gorm"

	"github.com/caraml-dev/turing/api/turing/models"
)

// routersRepository implements service.RoutersRepository
type routersRepository struct {
	db *gorm.DB
}

func NewRoutersRepository(db *gorm.DB) *routersRepository {
	return &routersRepository{
		db: db,
	}
}

func (rr *routersRepository) CountRoutersByCurrentVersionID(routerVersionID models.ID) int64 {
	var upstreamRefs int64
	rr.db.Table("routers").
		Where("routers.curr_router_version_id = ?", routerVersionID).
		Count(&upstreamRefs)
	return upstreamRefs
}
