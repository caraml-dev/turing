package webhook

import "github.com/caraml-dev/mlp/api/pkg/webhooks"

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

var eventList = []webhooks.EventType{
	OnRouterCreated,
	OnRouterUpdated,
	OnRouterDeleted,
	OnRouterDeployed,
	OnRouterUndeployed,
	OnEnsemblerCreated,
	OnEnsemblerUpdated,
	OnEnsemblerDeleted,
}
