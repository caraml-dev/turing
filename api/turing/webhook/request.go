package webhook

import (
	"github.com/caraml-dev/mlp/api/pkg/webhooks"
	"github.com/caraml-dev/turing/api/turing/models"
)

type Request struct {
	EventType webhooks.EventType     `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
}

// Adds the eventType to the body of the webhook request so that a single webhook endpoint is able to respond
// differently to different event types, especially if the same webhook endpoint is configured for multiple events,
// this is because the event type does not normally get sent to the webhook endpoint.
func generateBody(eventType webhooks.EventType, item interface{}) *Request {
	data := make(map[string]interface{})

	switch item.(type) {
	case *models.EnsemblerLike, models.EnsemblerLike:
		data["ensembler"] = item
	case *models.RouterVersion:
		data["router_version"] = item
	case *models.Router:
		data["router"] = item
	}

	return &Request{
		EventType: eventType,
		Data:      data,
	}
}
