package mocks

import "github.com/stretchr/testify/mock"

type RPCClient struct {
	mock.Mock
}

func (_m *RPCClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	ret := _m.Called(serviceMethod, args, reply)

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(error)
	}

	return r0
}
