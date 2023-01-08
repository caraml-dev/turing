// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	models "github.com/caraml-dev/turing/api/turing/models"
)

// RouterVersionsRepository is an autogenerated mock type for the RouterVersionsRepository type
type RouterVersionsRepository struct {
	mock.Mock
}

// Delete provides a mock function with given fields: routerVersion
func (_m *RouterVersionsRepository) Delete(routerVersion *models.RouterVersion) error {
	ret := _m.Called(routerVersion)

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.RouterVersion) error); ok {
		r0 = rf(routerVersion)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FindByID provides a mock function with given fields: routerVersionID
func (_m *RouterVersionsRepository) FindByID(routerVersionID models.ID) (*models.RouterVersion, error) {
	ret := _m.Called(routerVersionID)

	var r0 *models.RouterVersion
	if rf, ok := ret.Get(0).(func(models.ID) *models.RouterVersion); ok {
		r0 = rf(routerVersionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.RouterVersion)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.ID) error); ok {
		r1 = rf(routerVersionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindByRouterIDAndVersion provides a mock function with given fields: routerID, version
func (_m *RouterVersionsRepository) FindByRouterIDAndVersion(routerID models.ID, version uint) (*models.RouterVersion, error) {
	ret := _m.Called(routerID, version)

	var r0 *models.RouterVersion
	if rf, ok := ret.Get(0).(func(models.ID, uint) *models.RouterVersion); ok {
		r0 = rf(routerID, version)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.RouterVersion)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.ID, uint) error); ok {
		r1 = rf(routerID, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindLatestVersion provides a mock function with given fields: routerID
func (_m *RouterVersionsRepository) FindLatestVersion(routerID models.ID) (*models.RouterVersion, error) {
	ret := _m.Called(routerID)

	var r0 *models.RouterVersion
	if rf, ok := ret.Get(0).(func(models.ID) *models.RouterVersion); ok {
		r0 = rf(routerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.RouterVersion)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.ID) error); ok {
		r1 = rf(routerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: routerID
func (_m *RouterVersionsRepository) List(routerID models.ID) ([]*models.RouterVersion, error) {
	ret := _m.Called(routerID)

	var r0 []*models.RouterVersion
	if rf, ok := ret.Get(0).(func(models.ID) []*models.RouterVersion); ok {
		r0 = rf(routerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.RouterVersion)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.ID) error); ok {
		r1 = rf(routerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByStatus provides a mock function with given fields: routerID, status
func (_m *RouterVersionsRepository) ListByStatus(routerID models.ID, status models.RouterVersionStatus) ([]*models.RouterVersion, error) {
	ret := _m.Called(routerID, status)

	var r0 []*models.RouterVersion
	if rf, ok := ret.Get(0).(func(models.ID, models.RouterVersionStatus) []*models.RouterVersion); ok {
		r0 = rf(routerID, status)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.RouterVersion)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.ID, models.RouterVersionStatus) error); ok {
		r1 = rf(routerID, status)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: routerVersion
func (_m *RouterVersionsRepository) Save(routerVersion *models.RouterVersion) (*models.RouterVersion, error) {
	ret := _m.Called(routerVersion)

	var r0 *models.RouterVersion
	if rf, ok := ret.Get(0).(func(*models.RouterVersion) *models.RouterVersion); ok {
		r0 = rf(routerVersion)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.RouterVersion)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*models.RouterVersion) error); ok {
		r1 = rf(routerVersion)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewRouterVersionsRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewRouterVersionsRepository creates a new instance of RouterVersionsRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRouterVersionsRepository(t mockConstructorTestingTNewRouterVersionsRepository) *RouterVersionsRepository {
	mock := &RouterVersionsRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
