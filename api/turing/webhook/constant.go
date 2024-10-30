package webhook

import (
	"errors"

	"github.com/caraml-dev/mlp/api/pkg/webhooks"
)

var (
	OnRouterCreated    = webhooks.EventType("on-router-created")
	OnRouterUpdated    = webhooks.EventType("on-router-updated")
	OnRouterDeleted    = webhooks.EventType("on-router-deleted")
	OnRouterDeployed   = webhooks.EventType("on-router-deployed")
	OnRouterUndeployed = webhooks.EventType("on-router-undeployed")

	OnEnsemblerCreated = webhooks.EventType("on-ensembler-created")
	OnEnsemblerUpdated = webhooks.EventType("on-ensembler-updated")
	OnEnsemblerDeleted = webhooks.EventType("on-ensembler-deleted")
)

var (
	// event list for router event
	eventListRouter = map[webhooks.EventType]bool{
		OnRouterCreated: true,
		OnRouterUpdated: true,
		OnRouterDeleted: true,
	}

	// event list for router deployment event
	eventListRouterDeployment = map[webhooks.EventType]bool{
		OnRouterDeployed:   true,
		OnRouterUndeployed: true,
	}

	// event list for ensembler event
	eventListEnsembler = map[webhooks.EventType]bool{
		OnEnsemblerCreated: true,
		OnEnsemblerUpdated: true,
		OnEnsemblerDeleted: true,
	}
)

var (
	ErrInvalidEventType = errors.New("invalid event type")
)
