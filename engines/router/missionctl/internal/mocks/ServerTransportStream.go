// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	metadata "google.golang.org/grpc/metadata"
)

// ServerTransportStream is an autogenerated mock type for the ServerTransportStream type
type ServerTransportStream struct {
	mock.Mock
}

// Method provides a mock function with given fields:
func (_m *ServerTransportStream) Method() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SendHeader provides a mock function with given fields: md
func (_m *ServerTransportStream) SendHeader(md metadata.MD) error {
	ret := _m.Called(md)

	var r0 error
	if rf, ok := ret.Get(0).(func(metadata.MD) error); ok {
		r0 = rf(md)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetHeader provides a mock function with given fields: md
func (_m *ServerTransportStream) SetHeader(md metadata.MD) error {
	ret := _m.Called(md)

	var r0 error
	if rf, ok := ret.Get(0).(func(metadata.MD) error); ok {
		r0 = rf(md)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTrailer provides a mock function with given fields: md
func (_m *ServerTransportStream) SetTrailer(md metadata.MD) error {
	ret := _m.Called(md)

	var r0 error
	if rf, ok := ret.Get(0).(func(metadata.MD) error); ok {
		r0 = rf(md)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
