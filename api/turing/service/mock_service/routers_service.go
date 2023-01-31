// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/user/Documents/Code/github/test_refactoring/turing/api/turing/service/routers_service.go

// Package mock_service is a generated GoMock package.
package mock_service

import (
	models "github.com/caraml-dev/turing/api/turing/models"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockRoutersService is a mock of RoutersService interface
type MockRoutersService struct {
	ctrl     *gomock.Controller
	recorder *MockRoutersServiceMockRecorder
}

// MockRoutersServiceMockRecorder is the mock recorder for MockRoutersService
type MockRoutersServiceMockRecorder struct {
	mock *MockRoutersService
}

// NewMockRoutersService creates a new mock instance
func NewMockRoutersService(ctrl *gomock.Controller) *MockRoutersService {
	mock := &MockRoutersService{ctrl: ctrl}
	mock.recorder = &MockRoutersServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRoutersService) EXPECT() *MockRoutersServiceMockRecorder {
	return m.recorder
}

// ListRouters mocks base method
func (m *MockRoutersService) ListRouters(projectID models.ID, environmentName string) ([]*models.Router, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListRouters", projectID, environmentName)
	ret0, _ := ret[0].([]*models.Router)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListRouters indicates an expected call of ListRouters
func (mr *MockRoutersServiceMockRecorder) ListRouters(projectID, environmentName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRouters", reflect.TypeOf((*MockRoutersService)(nil).ListRouters), projectID, environmentName)
}

// Save mocks base method
func (m *MockRoutersService) Save(router *models.Router) (*models.Router, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", router)
	ret0, _ := ret[0].(*models.Router)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Save indicates an expected call of Save
func (mr *MockRoutersServiceMockRecorder) Save(router interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockRoutersService)(nil).Save), router)
}

// FindByID mocks base method
func (m *MockRoutersService) FindByID(routerID models.ID) (*models.Router, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", routerID)
	ret0, _ := ret[0].(*models.Router)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID
func (mr *MockRoutersServiceMockRecorder) FindByID(routerID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockRoutersService)(nil).FindByID), routerID)
}

// FindByProjectAndName mocks base method
func (m *MockRoutersService) FindByProjectAndName(projectID models.ID, routerName string) (*models.Router, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByProjectAndName", projectID, routerName)
	ret0, _ := ret[0].(*models.Router)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByProjectAndName indicates an expected call of FindByProjectAndName
func (mr *MockRoutersServiceMockRecorder) FindByProjectAndName(projectID, routerName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByProjectAndName", reflect.TypeOf((*MockRoutersService)(nil).FindByProjectAndName), projectID, routerName)
}

// FindByProjectAndEnvironmentAndName mocks base method
func (m *MockRoutersService) FindByProjectAndEnvironmentAndName(projectID models.ID, environmentName, routerName string) (*models.Router, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByProjectAndEnvironmentAndName", projectID, environmentName, routerName)
	ret0, _ := ret[0].(*models.Router)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByProjectAndEnvironmentAndName indicates an expected call of FindByProjectAndEnvironmentAndName
func (mr *MockRoutersServiceMockRecorder) FindByProjectAndEnvironmentAndName(projectID, environmentName, routerName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByProjectAndEnvironmentAndName", reflect.TypeOf((*MockRoutersService)(nil).FindByProjectAndEnvironmentAndName), projectID, environmentName, routerName)
}

// Delete mocks base method
func (m *MockRoutersService) Delete(router *models.Router) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", router)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockRoutersServiceMockRecorder) Delete(router interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockRoutersService)(nil).Delete), router)
}
