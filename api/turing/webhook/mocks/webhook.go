// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	webhooks "github.com/caraml-dev/mlp/api/pkg/webhooks"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// TriggerWebhooks provides a mock function with given fields: ctx, eventType, body
func (_m *Client) TriggerWebhooks(ctx context.Context, eventType webhooks.EventType, body interface{}) error {
	ret := _m.Called(ctx, eventType, body)

	if len(ret) == 0 {
		panic("no return value specified for TriggerWebhooks")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, webhooks.EventType, interface{}) error); ok {
		r0 = rf(ctx, eventType, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
