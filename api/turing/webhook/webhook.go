package webhook

import (
	"context"

	"github.com/caraml-dev/mlp/api/pkg/webhooks"

	"github.com/caraml-dev/turing/api/turing/models"
)

type Client interface {
	TriggerRouterEvent(ctx context.Context, eventType webhooks.EventType, router *models.Router) error
	TriggerEnsemblerEvent(ctx context.Context, eventType webhooks.EventType, ensembler models.EnsemblerLike) error
}

func NewWebhook(cfg *webhooks.Config) (Client, error) {
	webhookManager, err := webhooks.InitializeWebhooks(cfg, eventList)
	if err != nil {
		return nil, err
	}

	return &webhook{
		manager: webhookManager,
	}, nil
}

type webhook struct {
	manager webhooks.WebhookManager
}

func (w *webhook) triggerEvent(ctx context.Context, eventType webhooks.EventType, body interface{}) error {
	if !w.isEventConfigured(eventType) {
		return nil
	}

	if err := w.manager.InvokeWebhooks(
		ctx,
		eventType,
		body,
		webhooks.NoOpCallback,
		webhooks.NoOpErrorHandler,
	); err != nil {
		return err
	}

	return nil
}

func (w *webhook) isEventConfigured(eventType webhooks.EventType) bool {
	return w.manager != nil && w.manager.IsEventConfigured(eventType)
}

func (w *webhook) TriggerRouterEvent(ctx context.Context, eventType webhooks.EventType, router *models.Router) error {
	body := &routerRequest{
		EventType: eventType,
		Router:    router,
	}

	return w.triggerEvent(ctx, eventType, body)
}

func (w *webhook) TriggerEnsemblerEvent(
	ctx context.Context,
	eventType webhooks.EventType,
	ensembler models.EnsemblerLike,
) error {
	body := &ensemblerRequest{
		EventType: eventType,
		Ensembler: ensembler,
	}

	return w.triggerEvent(ctx, eventType, body)
}
