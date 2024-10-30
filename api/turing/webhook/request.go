package webhook

import (
	"github.com/caraml-dev/mlp/api/pkg/webhooks"

	"github.com/caraml-dev/turing/api/turing/models"
)

type routerRequest struct {
	EventType webhooks.EventType `json:"event_type"`
	Router    *models.Router     `json:"router"`
}

type routerDeploymentRequest struct {
	EventType     webhooks.EventType    `json:"event_type"`
	ProjectID     uint                  `json:"project_id"`
	RouterVersion *models.RouterVersion `json:"router_version"`
}

type ensemblerRequest struct {
	EventType webhooks.EventType   `json:"event_type"`
	Ensembler models.EnsemblerLike `json:"ensembler"`
}
