package service

import (
	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
)

type EventService interface {
	ListEvents(routerID int) ([]*models.Event, error)
	ClearEvents(routerID int) error
	Save(event *models.Event) error
}

// NewEventService creates a new events service
func NewEventService(db *gorm.DB) EventService {
	return &eventsService{db: db}
}

type eventsService struct {
	db *gorm.DB
}

func (svc *eventsService) query() *gorm.DB {
	return svc.db.
		Select("events.*")
}

func (svc eventsService) ListEvents(routerID int) ([]*models.Event, error) {
	var events []*models.Event
	query := svc.query().
		Where("router_id = ?", routerID).
		Order("created_at asc").
		Find(&events)
	return events, query.Error
}

func (svc eventsService) ClearEvents(routerID int) error {
	query := svc.query().
		Where("router_id = ?", routerID).
		Delete(&models.Event{})
	return query.Error
}

func (svc eventsService) Save(event *models.Event) error {
	return svc.db.Save(event).Error
}
