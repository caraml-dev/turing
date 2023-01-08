// Code generated by MockGen. DO NOT EDIT.
// Source: router_monitoring_service.go

// Package mock_service is a generated GoMock package.
package mock_service

import (
	models "github.com/caraml-dev/turing/api/turing/models"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockRouterMonitoringService is a mock of RouterMonitoringService interface
type MockRouterMonitoringService struct {
	ctrl     *gomock.Controller
	recorder *MockRouterMonitoringServiceMockRecorder
}

// MockRouterMonitoringServiceMockRecorder is the mock recorder for MockRouterMonitoringService
type MockRouterMonitoringServiceMockRecorder struct {
	mock *MockRouterMonitoringService
}

// NewMockRouterMonitoringService creates a new mock instance
func NewMockRouterMonitoringService(ctrl *gomock.Controller) *MockRouterMonitoringService {
	mock := &MockRouterMonitoringService{ctrl: ctrl}
	mock.recorder = &MockRouterMonitoringServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRouterMonitoringService) EXPECT() *MockRouterMonitoringServiceMockRecorder {
	return m.recorder
}

// GenerateMonitoringURL mocks base method
func (m *MockRouterMonitoringService) GenerateMonitoringURL(projectID models.ID, environmentName, routerName string, routerVersion *uint) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateMonitoringURL", projectID, environmentName, routerName, routerVersion)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateMonitoringURL indicates an expected call of GenerateMonitoringURL
func (mr *MockRouterMonitoringServiceMockRecorder) GenerateMonitoringURL(projectID, environmentName, routerName, routerVersion interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateMonitoringURL", reflect.TypeOf((*MockRouterMonitoringService)(nil).GenerateMonitoringURL), projectID, environmentName, routerName, routerVersion)
}
