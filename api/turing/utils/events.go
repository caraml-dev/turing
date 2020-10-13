package utils

import (
	"github.com/gojek/turing/api/turing/models"
)

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
