// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	context "context"

	download "github.com/MarcoTomasRodriguez/hget/internal/download"
	mock "github.com/stretchr/testify/mock"
)

// Downloader is an autogenerated mock type for the Downloader type
type Downloader struct {
	mock.Mock
}

// DeleteDownloadById provides a mock function with given fields: id
func (_m *Downloader) DeleteDownloadById(id string) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Download provides a mock function with given fields: _a0, ctx
func (_m *Downloader) Download(_a0 download.Download, ctx context.Context) error {
	ret := _m.Called(_a0, ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(download.Download, context.Context) error); ok {
		r0 = rf(_a0, ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FindAllDownloads provides a mock function with given fields:
func (_m *Downloader) FindAllDownloads() ([]download.Download, error) {
	ret := _m.Called()

	var r0 []download.Download
	if rf, ok := ret.Get(0).(func() []download.Download); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]download.Download)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindDownloadById provides a mock function with given fields: id
func (_m *Downloader) FindDownloadById(id string) (download.Download, error) {
	ret := _m.Called(id)

	var r0 download.Download
	if rf, ok := ret.Get(0).(func(string) download.Download); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(download.Download)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindDownloadByUrl provides a mock function with given fields: url
func (_m *Downloader) FindDownloadByUrl(url string) (download.Download, error) {
	ret := _m.Called(url)

	var r0 download.Download
	if rf, ok := ret.Get(0).(func(string) download.Download); ok {
		r0 = rf(url)
	} else {
		r0 = ret.Get(0).(download.Download)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(url)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InitDownload provides a mock function with given fields: url, workers
func (_m *Downloader) InitDownload(url string, workers uint8) (download.Download, error) {
	ret := _m.Called(url, workers)

	var r0 download.Download
	if rf, ok := ret.Get(0).(func(string, uint8) download.Download); ok {
		r0 = rf(url, workers)
	} else {
		r0 = ret.Get(0).(download.Download)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, uint8) error); ok {
		r1 = rf(url, workers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewDownloader interface {
	mock.TestingT
	Cleanup(func())
}

// NewDownloader creates a new instance of Downloader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDownloader(t mockConstructorTestingTNewDownloader) *Downloader {
	mock := &Downloader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}