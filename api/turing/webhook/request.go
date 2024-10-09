package webhook

import (
	"github.com/caraml-dev/mlp/api/pkg/webhooks"

	"github.com/caraml-dev/turing/api/turing/models"
)

type routerRequest struct {
	EventType webhooks.EventType `json:"event_type"`
	Router    *models.Router     `json:"router"`
}

type ensemblerRequest struct {
	EventType webhooks.EventType   `json:"event_type"`
	Ensembler models.EnsemblerLike `json:"ensembler"`
}
