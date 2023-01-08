// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	context "context"
	io "io"

	download "github.com/MarcoTomasRodriguez/hget/internal/download"

	mock "github.com/stretchr/testify/mock"
)

// Network is an autogenerated mock type for the Network type
type Network struct {
	mock.Mock
}

// DownloadResource provides a mock function with given fields: url, start, end, writer, ctx
func (_m *Network) DownloadResource(url string, start int64, end int64, writer io.Writer, ctx context.Context) error {
	ret := _m.Called(url, start, end, writer, ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64, int64, io.Writer, context.Context) error); ok {
		r0 = rf(url, start, end, writer, ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FetchResource provides a mock function with given fields: url
func (_m *Network) FetchResource(url string) (download.Resource, error) {
	ret := _m.Called(url)

	var r0 download.Resource
	if rf, ok := ret.Get(0).(func(string) download.Resource); ok {
		r0 = rf(url)
	} else {
		r0 = ret.Get(0).(download.Resource)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(url)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewNetwork interface {
	mock.TestingT
	Cleanup(func())
}

// NewNetwork creates a new instance of Network. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewNetwork(t mockConstructorTestingTNewNetwork) *Network {
	mock := &Network{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}