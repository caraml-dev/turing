//go:build integration

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/caraml-dev/turing/api/turing/database"
	"github.com/caraml-dev/turing/api/turing/models"
)

func TestEventServiceIntegration(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		svc := NewEventService(db)

		// create router
		router := &models.Router{
			Model: models.Model{
				ID: 1,
			},
			ProjectID:       1,
			EnvironmentName: "env",
			Name:            "hamburger",
			Status:          models.RouterStatusPending,
		}

		events := []*models.Event{
			{
				Router:    router,
				EventType: "info",
				Message:   "hello hello",
			},
			{
				Router:    router,
				EventType: "error",
				Message:   "sum ting wong",
			},
		}
		for _, event := range events {
			err := svc.Save(event)
			assert.NoError(t, err)
		}

		gotEvents, err := svc.ListEvents(int(router.ID))
		assert.NoError(t, err)
		for i := range gotEvents {
			assert.Equal(t, gotEvents[i].RouterID, events[i].RouterID)
			assert.Equal(t, gotEvents[i].EventType, events[i].EventType)
			assert.Equal(t, gotEvents[i].Message, events[i].Message)
		}

		err = svc.ClearEvents(int(router.ID))
		assert.NoError(t, err)
		gotEvents, err = svc.ListEvents(int(router.ID))
		assert.NoError(t, err)
		assert.Empty(t, gotEvents)
	})
}

func TestEventChannel(t *testing.T) {
	ch := NewEventChannel()
	event := models.Event{
		RouterID: 1,
		Message:  "hello",
	}
	go ch.Write(&event)
	got, _ := ch.Read()
	assert.Equal(t, *got, event)
	ch.Close()
	_, done := ch.Read()
	assert.Equal(t, done, true)
}
