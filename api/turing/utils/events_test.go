package utils

import (
	"testing"

	"github.com/gojek/turing/api/turing/models"
	"gotest.tools/assert"
)

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
