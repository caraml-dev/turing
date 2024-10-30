package webhook

import (
	"context"

	"github.com/caraml-dev/mlp/api/pkg/webhooks"

	"github.com/caraml-dev/turing/api/turing/models"
)

type Client interface {
	TriggerRouterEvent(ctx context.Context, eventType webhooks.EventType, router *models.Router) error
	TriggerRouterDeploymentEvent(
		ctx context.Context,
		eventType webhooks.EventType,
		router *models.RouterVersion,
		projectID uint,
	) error
	TriggerEnsemblerEvent(ctx context.Context, eventType webhooks.EventType, ensembler models.EnsemblerLike) error
}

func NewWebhook(cfg *webhooks.Config) (Client, error) {
	var eventTypeList []webhooks.EventType

	for eventType := range eventListRouter {
		eventTypeList = append(eventTypeList, eventType)
	}

	for eventType := range eventListEnsembler {
		eventTypeList = append(eventTypeList, eventType)
	}

	webhookManager, err := webhooks.InitializeWebhooks(cfg, eventTypeList)
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
	if isValid := eventListRouter[eventType]; !isValid {
		return ErrInvalidEventType
	}

	body := &routerRequest{
		EventType: eventType,
		Router:    router,
	}

	return w.triggerEvent(ctx, eventType, body)
}

func (w *webhook) TriggerRouterDeploymentEvent(
	ctx context.Context,
	eventType webhooks.EventType,
	router *models.RouterVersion,
	projectID uint,
) error {
	if isValid := eventListRouterDeployment[eventType]; !isValid {
		return ErrInvalidEventType
	}

	body := &routerDeploymentRequest{
		EventType:     eventType,
		ProjectID:     projectID,
		RouterVersion: router,
	}

	return w.triggerEvent(ctx, eventType, body)
}

func (w *webhook) TriggerEnsemblerEvent(
	ctx context.Context,
	eventType webhooks.EventType,
	ensembler models.EnsemblerLike,
) error {
	if isValid := eventListEnsembler[eventType]; !isValid {
		return ErrInvalidEventType
	}

	body := &ensemblerRequest{
		EventType: eventType,
		Ensembler: ensembler,
	}

	return w.triggerEvent(ctx, eventType, body)
}
