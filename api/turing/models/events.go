package models

import "fmt"

type EventType string

const (
	EventTypeError EventType = "error"
	EventTypeInfo  EventType = "info"
)

type EventStage string

const (
	EventStageDeployingDependencies      EventStage = "deploying dependencies"
	EventStageDeployingServices          EventStage = "deploying services"
	EventStageDeploymentSuccess          EventStage = "deployment success"
	EventStageDeploymentFailed           EventStage = "deployment failed"
	EventStageRollback                   EventStage = "rollback deployment"
	EventStageUpdatingEndpoint           EventStage = "updating endpoint"
	EventStageUndeployingPreviousVersion EventStage = "undeploying previous version"
	EventStageDeletingDependencies       EventStage = "deleting dependencies"
	EventStageUndeployingServices        EventStage = "undeploying services"
	EventStageDeletingEndpoint           EventStage = "deleting endpoint"
	EventStageUndeploymentFailed         EventStage = "undeployment failed"
	EventStageUndeploymentSuccess        EventStage = "undeployment success"
)

// Event is a log of an event taking place during deployment
// or undeployment of a router.
type Event struct {
	Model

	// Router id this event is for
	RouterID ID      `json:"-"`
	Router   *Router `json:"-"`

	// Version of router that triggered this deployment event.
	// May not necessarily pertain to the resource described by the event,
	// e.g. for removal of old versions, version will point to the new version
	// that triggered the event.
	Version uint `json:"version"`

	// EventType type of event
	EventType EventType `json:"event_type"`

	// Stage of deployment/undeployment
	Stage EventStage `json:"stage"`

	// Message describing the event
	Message string `json:"message"`
}

// NewErrorEvent creates a new error event
func NewErrorEvent(stage EventStage, message string, args ...interface{}) *Event {
	return newEvent(stage, EventTypeError, message, args...)
}

// NewInfoEvent creates a new info event
func NewInfoEvent(stage EventStage, message string, args ...interface{}) *Event {
	return newEvent(stage, EventTypeInfo, message, args...)
}

func newEvent(stage EventStage, eventType EventType, message string, args ...interface{}) *Event {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	return &Event{
		Stage:     stage,
		EventType: eventType,
		Message:   message,
	}
}

func (e *Event) SetRouter(router *Router) {
	e.RouterID = router.ID
}

func (e *Event) SetVersion(version uint) {
	e.Version = version
}

type EventChannel struct {
	c        chan *Event
	s        chan bool
	isClosed bool
}

func NewEventChannel() *EventChannel {
	ch := make(chan *Event)
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
func (ch *EventChannel) Write(event *Event) {
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
func (ch *EventChannel) Read() (*Event, bool) {
	event, more := <-ch.c
	return event, !more
}
