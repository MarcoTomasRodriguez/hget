// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	io "io"

	progressbar "github.com/MarcoTomasRodriguez/hget/pkg/progressbar"
	mock "github.com/stretchr/testify/mock"
)

// ProgressBar is an autogenerated mock type for the ProgressBar type
type ProgressBar struct {
	mock.Mock
}

// Add provides a mock function with given fields: total, units, prefix
func (_m *ProgressBar) Add(total int64, units progressbar.Units, prefix string) io.Writer {
	ret := _m.Called(total, units, prefix)

	var r0 io.Writer
	if rf, ok := ret.Get(0).(func(int64, progressbar.Units, string) io.Writer); ok {
		r0 = rf(total, units, prefix)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Writer)
		}
	}

	return r0
}

// Start provides a mock function with given fields:
func (_m *ProgressBar) Start() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Stop provides a mock function with given fields:
func (_m *ProgressBar) Stop() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewProgressBar interface {
	mock.TestingT
	Cleanup(func())
}

// NewProgressBar creates a new instance of ProgressBar. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewProgressBar(t mockConstructorTestingTNewProgressBar) *ProgressBar {
	mock := &ProgressBar{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}