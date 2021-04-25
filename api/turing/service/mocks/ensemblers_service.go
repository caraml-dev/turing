// Code generated by mockery 2.7.4. DO NOT EDIT.

package mocks

import (
	models "github.com/gojek/turing/api/turing/models"
	mock "github.com/stretchr/testify/mock"

	service "github.com/gojek/turing/api/turing/service"
)

// EnsemblersService is an autogenerated mock type for the EnsemblersService type
type EnsemblersService struct {
	mock.Mock
}

// FindByID provides a mock function with given fields: id
func (_m *EnsemblersService) FindByID(id models.ID) (models.EnsemblerLike, error) {
	ret := _m.Called(id)

	var r0 models.EnsemblerLike
	if rf, ok := ret.Get(0).(func(models.ID) models.EnsemblerLike); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(models.EnsemblerLike)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.ID) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: projectID, query
func (_m *EnsemblersService) List(projectID models.ID, query service.ListEnsemblersQuery) (*service.PaginatedResults, error) {
	ret := _m.Called(projectID, query)

	var r0 *service.PaginatedResults
	if rf, ok := ret.Get(0).(func(models.ID, service.ListEnsemblersQuery) *service.PaginatedResults); ok {
		r0 = rf(projectID, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*service.PaginatedResults)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.ID, service.ListEnsemblersQuery) error); ok {
		r1 = rf(projectID, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: ensembler
func (_m *EnsemblersService) Save(ensembler models.EnsemblerLike) (models.EnsemblerLike, error) {
	ret := _m.Called(ensembler)

	var r0 models.EnsemblerLike
	if rf, ok := ret.Get(0).(func(models.EnsemblerLike) models.EnsemblerLike); ok {
		r0 = rf(ensembler)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(models.EnsemblerLike)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(models.EnsemblerLike) error); ok {
		r1 = rf(ensembler)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
