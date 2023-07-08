// Code generated by mockery v2.28.2. DO NOT EDIT.

package mocks

import (
	json "encoding/json"

	client "github.com/caraml-dev/mlp/api/client"

	merlinclient "github.com/caraml-dev/merlin/client"

	mock "github.com/stretchr/testify/mock"

	models "github.com/caraml-dev/turing/api/turing/models"

	service "github.com/caraml-dev/turing/api/turing/service"
)

// DeploymentService is an autogenerated mock type for the DeploymentService type
type DeploymentService struct {
	mock.Mock
}

// DeleteRouterEndpoint provides a mock function with given fields: project, environment, routerVersion
func (_m *DeploymentService) DeleteRouterEndpoint(project *client.Project, environment *merlinclient.Environment, routerVersion *models.RouterVersion) error {
	ret := _m.Called(project, environment, routerVersion)

	var r0 error
	if rf, ok := ret.Get(0).(func(*client.Project, *merlinclient.Environment, *models.RouterVersion) error); ok {
		r0 = rf(project, environment, routerVersion)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeployRouterVersion provides a mock function with given fields: project, environment, router, routerVersion, routerServiceAccountKey, enricherServiceAccountKey, ensemblerServiceAccountKey, expEngineServiceAccountKey, pyfuncEnsembler, experimentConfig, eventsCh
func (_m *DeploymentService) DeployRouterVersion(project *client.Project, environment *merlinclient.Environment, router *models.Router, routerVersion *models.RouterVersion, routerServiceAccountKey string, enricherServiceAccountKey string, ensemblerServiceAccountKey string, expEngineServiceAccountKey string, pyfuncEnsembler *models.PyFuncEnsembler, experimentConfig json.RawMessage, eventsCh *service.EventChannel) (string, error) {
	ret := _m.Called(project, environment, router, routerVersion, routerServiceAccountKey, enricherServiceAccountKey, ensemblerServiceAccountKey, expEngineServiceAccountKey, pyfuncEnsembler, experimentConfig, eventsCh)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(*client.Project, *merlinclient.Environment, *models.Router, *models.RouterVersion, string, string, string, string, *models.PyFuncEnsembler, json.RawMessage, *service.EventChannel) (string, error)); ok {
		return rf(project, environment, router, routerVersion, routerServiceAccountKey, enricherServiceAccountKey, ensemblerServiceAccountKey, expEngineServiceAccountKey, pyfuncEnsembler, experimentConfig, eventsCh)
	}
	if rf, ok := ret.Get(0).(func(*client.Project, *merlinclient.Environment, *models.Router, *models.RouterVersion, string, string, string, string, *models.PyFuncEnsembler, json.RawMessage, *service.EventChannel) string); ok {
		r0 = rf(project, environment, router, routerVersion, routerServiceAccountKey, enricherServiceAccountKey, ensemblerServiceAccountKey, expEngineServiceAccountKey, pyfuncEnsembler, experimentConfig, eventsCh)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(*client.Project, *merlinclient.Environment, *models.Router, *models.RouterVersion, string, string, string, string, *models.PyFuncEnsembler, json.RawMessage, *service.EventChannel) error); ok {
		r1 = rf(project, environment, router, routerVersion, routerServiceAccountKey, enricherServiceAccountKey, ensemblerServiceAccountKey, expEngineServiceAccountKey, pyfuncEnsembler, experimentConfig, eventsCh)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLocalSecret provides a mock function with given fields: serviceAccountKeyFilePath
func (_m *DeploymentService) GetLocalSecret(serviceAccountKeyFilePath string) (*string, error) {
	ret := _m.Called(serviceAccountKeyFilePath)

	var r0 *string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*string, error)); ok {
		return rf(serviceAccountKeyFilePath)
	}
	if rf, ok := ret.Get(0).(func(string) *string); ok {
		r0 = rf(serviceAccountKeyFilePath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(serviceAccountKeyFilePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UndeployRouterVersion provides a mock function with given fields: project, environment, routerVersion, eventsCh, isCleanUp
func (_m *DeploymentService) UndeployRouterVersion(project *client.Project, environment *merlinclient.Environment, routerVersion *models.RouterVersion, eventsCh *service.EventChannel, isCleanUp bool) error {
	ret := _m.Called(project, environment, routerVersion, eventsCh, isCleanUp)

	var r0 error
	if rf, ok := ret.Get(0).(func(*client.Project, *merlinclient.Environment, *models.RouterVersion, *service.EventChannel, bool) error); ok {
		r0 = rf(project, environment, routerVersion, eventsCh, isCleanUp)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewDeploymentService interface {
	mock.TestingT
	Cleanup(func())
}

// NewDeploymentService creates a new instance of DeploymentService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDeploymentService(t mockConstructorTestingTNewDeploymentService) *DeploymentService {
	mock := &DeploymentService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
