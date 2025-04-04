// Code generated by mockery v2.52.4. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// TempPaths is an autogenerated mock type for the TempPaths type
type TempPaths struct {
	mock.Mock
}

// Add provides a mock function with given fields: key, value
func (_m *TempPaths) Add(key string, value string) {
	_m.Called(key, value)
}

// GetPath provides a mock function with given fields: key
func (_m *TempPaths) GetPath(key string) (string, error) {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for GetPath")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(key)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPathIfExists provides a mock function with given fields: key
func (_m *TempPaths) GetPathIfExists(key string) string {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for GetPathIfExists")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetPaths provides a mock function with no fields
func (_m *TempPaths) GetPaths() map[string]string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetPaths")
	}

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func() map[string]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	return r0
}

// NewTempPaths creates a new instance of TempPaths. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTempPaths(t interface {
	mock.TestingT
	Cleanup(func())
}) *TempPaths {
	mock := &TempPaths{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
