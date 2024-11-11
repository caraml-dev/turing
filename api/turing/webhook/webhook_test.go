package webhook

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/caraml-dev/mlp/api/pkg/webhooks"
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/turing/api/turing/models"
)

type routerRequest struct {
	EventType webhooks.EventType `json:"event_type"`
	Router    *models.Router     `json:"router"`
}

func TestNewWebhook(t *testing.T) {
	type args struct {
		cfg *webhooks.Config
	}
	tests := []struct {
		name    string
		args    args
		want    Client
		wantErr bool
	}{
		{
			name: "positive",
			args: args{
				cfg: &webhooks.Config{},
			},
			want: webhook{},
		},
		{
			name: "negative - num retries is negative",
			args: args{
				cfg: &webhooks.Config{
					Enabled: true,
					Config: map[webhooks.EventType][]webhooks.WebhookConfig{
						OnRouterCreated: {{NumRetries: -1}},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWebhook(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWebhook() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webhook_triggerEvent(t *testing.T) {
	mockWebhookManager := webhooks.NewMockWebhookManager(t)

	type fields struct {
		manager webhooks.WebhookManager
	}
	type args struct {
		ctx       context.Context
		eventType webhooks.EventType
		body      interface{}
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		mockFunc func(args)
	}{
		{
			name: "positive - event not configured",
			fields: fields{
				manager: mockWebhookManager,
			},
			args: args{
				ctx:       context.TODO(),
				eventType: OnRouterCreated,
				body: routerRequest{
					EventType: OnRouterCreated,
					Router:    &models.Router{},
				},
			},
			mockFunc: func(args args) {
				mockWebhookManager.On("IsEventConfigured", args.eventType).Once().
					Return(false)
			},
		},
		{
			name: "positive - invoke webhook",
			fields: fields{
				manager: mockWebhookManager,
			},
			args: args{
				ctx:       context.TODO(),
				eventType: OnRouterCreated,
				body: routerRequest{
					EventType: OnRouterCreated,
					Router:    &models.Router{},
				},
			},
			mockFunc: func(args args) {
				mockWebhookManager.On("IsEventConfigured", args.eventType).
					Once().Return(true)
				mockWebhookManager.On(
					"InvokeWebhooks",
					args.ctx,
					args.eventType,
					&Request{
						EventType: args.eventType,
						Data:      args.body,
					},
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
		},
		{
			name: "negative - invoke webhook",
			fields: fields{
				manager: mockWebhookManager,
			},
			args: args{
				ctx:       context.TODO(),
				eventType: OnRouterCreated,
				body: routerRequest{
					EventType: OnRouterCreated,
					Router:    &models.Router{},
				},
			},
			mockFunc: func(args args) {
				mockWebhookManager.On("IsEventConfigured", args.eventType).
					Once().Return(true)
				mockWebhookManager.On(
					"InvokeWebhooks",
					args.ctx,
					args.eventType,
					&Request{
						EventType: args.eventType,
						Data:      args.body,
					},
					mock.Anything,
					mock.Anything,
				).Once().Return(errors.New("mock error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := webhook{
				webhookManager: tt.fields.manager,
			}
			tt.mockFunc(tt.args)
			if err := w.TriggerWebhooks(tt.args.ctx, tt.args.eventType, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("triggerEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
