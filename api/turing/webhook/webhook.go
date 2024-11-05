package webhook

import (
	"context"

	"github.com/caraml-dev/mlp/api/pkg/webhooks"
)

var (
	OnRouterCreated    = webhooks.EventType("on-router-created")
	OnRouterUpdated    = webhooks.EventType("on-router-updated")
	OnRouterDeleted    = webhooks.EventType("on-router-deleted")
	OnRouterDeployed   = webhooks.EventType("on-router-deployed")
	OnRouterUndeployed = webhooks.EventType("on-router-undeployed")

	OnRouterVersionCreated  = webhooks.EventType("on-router-version-created")
	OnRouterVersionDeleted  = webhooks.EventType("on-router-version-deleted")
	OnRouterVersionDeployed = webhooks.EventType("on-router-version-deployed")

	OnEnsemblerCreated = webhooks.EventType("on-ensembler-created")
	OnEnsemblerUpdated = webhooks.EventType("on-ensembler-updated")
	OnEnsemblerDeleted = webhooks.EventType("on-ensembler-deleted")
)

var events = []webhooks.EventType{
	OnRouterCreated,
	OnRouterUpdated,
	OnRouterDeleted,
	OnRouterDeployed,
	OnRouterVersionCreated,
	OnRouterVersionDeleted,
	OnRouterVersionDeployed,
	OnRouterUndeployed,
	OnEnsemblerCreated,
	OnEnsemblerUpdated,
	OnEnsemblerDeleted,
}

type webhook struct {
	webhookManager webhooks.WebhookManager
}

type Client interface {
	TriggerWebhooks(ctx context.Context, eventType webhooks.EventType, body interface{}) error
}

func NewWebhook(cfg *webhooks.Config) (Client, error) {
	webhookManager, err := webhooks.InitializeWebhooks(cfg, events)
	if err != nil {
		return nil, err
	}

	return webhook{
		webhookManager: webhookManager,
	}, nil
}

func (w webhook) TriggerWebhooks(ctx context.Context, eventType webhooks.EventType, body interface{}) error {
	if !w.isEventConfigured(eventType) {
		return nil
	}

	return w.webhookManager.InvokeWebhooks(
		ctx,
		eventType,
		body,
		webhooks.NoOpCallback,
		webhooks.NoOpErrorHandler,
	)
}

func (w webhook) isEventConfigured(eventType webhooks.EventType) bool {
	return w.webhookManager != nil && w.webhookManager.IsEventConfigured(eventType)
}
