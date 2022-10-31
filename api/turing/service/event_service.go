package service

import (
	"gorm.io/gorm"

	"github.com/caraml-dev/turing/api/turing/models"
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

type EventChannel struct {
	c        chan *models.Event
	s        chan bool
	isClosed bool
}

func NewEventChannel() *EventChannel {
	ch := make(chan *models.Event)
	stopch := make(chan bool)
	return &EventChannel{
		c: ch,
		s: stopch,
	}
}

// Close closes the channel in a thread-safe way, and updates isClosed to true.
func (ch *EventChannel) Close() {
	go func() {
		ch.s <- true
	}()
	close(ch.c)
	ch.isClosed = true
}

// Write safely writes to the channel. If the channel is closed, the
// event will be dropped.
func (ch *EventChannel) Write(event *models.Event) {
	if !ch.isClosed {
		select {
		case stop, ok := <-ch.s:
			if stop || !ok {
				ch.isClosed = true
				return
			}
		default:
			ch.c <- event
		}
	}
}

// Read reads an event from the stream. If the channel is closed,
// returns a boolean indicating that the channel is done.
func (ch *EventChannel) Read() (*models.Event, bool) {
	event, more := <-ch.c
	return event, !more
}
