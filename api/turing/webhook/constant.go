package webhook

import (
	"errors"

	"github.com/caraml-dev/mlp/api/pkg/webhooks"
)

var (
	OnRouterCreated    = webhooks.EventType("OnRouterCreated")
	OnRouterUpdated    = webhooks.EventType("OnRouterUpdated")
	OnRouterDeleted    = webhooks.EventType("OnRouterDeleted")
	OnRouterDeployed   = webhooks.EventType("OnRouterDeployed")
	OnRouterUndeployed = webhooks.EventType("OnRouterUndeployed")

	OnEnsemblerCreated = webhooks.EventType("OnEnsemblerCreated")
	OnEnsemblerUpdated = webhooks.EventType("OnEnsemblerUpdated")
	OnEnsemblerDeleted = webhooks.EventType("OnEnsemblerDeleted")
)

var (
	// event list for router event
	eventListRouter = map[webhooks.EventType]bool{
		OnRouterCreated:    true,
		OnRouterUpdated:    true,
		OnRouterDeleted:    true,
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
