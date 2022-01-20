package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewErrorEvent(t *testing.T) {
	errorEvent := NewErrorEvent(EventStageDeployingServices, "Error deploying %s", "svc")
	expected := Event{
		Stage:     EventStage("deploying services"),
		EventType: EventType("error"),
		Message:   "Error deploying svc",
	}
	assert.Equal(t, expected, *errorEvent)
}

func TestNewInfoEvent(t *testing.T) {
	infoEvent := NewInfoEvent(EventStageDeployingDependencies, "Deploying %s", "fluentd")
	expected := Event{
		Stage:     EventStage("deploying dependencies"),
		EventType: EventType("info"),
		Message:   "Deploying fluentd",
	}
	assert.Equal(t, expected, *infoEvent)
}

func TestEventSetters(t *testing.T) {
	event := Event{}
	testRouter := Router{
		Model: Model{
			ID: 1,
		},
	}
	event.SetRouter(&testRouter)
	event.SetVersion(uint(10))
	// Validate
	assert.Equal(t, ID(1), event.RouterID)
	assert.Equal(t, uint(10), event.Version)
}
