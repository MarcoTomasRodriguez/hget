// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Logger is an autogenerated mock type for the Logger type
type Logger struct {
	mock.Mock
}

// Error provides a mock function with given fields: message, a
func (_m *Logger) Error(message string, a ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, message)
	_ca = append(_ca, a...)
	_m.Called(_ca...)
}

// Info provides a mock function with given fields: message, a
func (_m *Logger) Info(message string, a ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, message)
	_ca = append(_ca, a...)
	_m.Called(_ca...)
}

// Warn provides a mock function with given fields: message, a
func (_m *Logger) Warn(message string, a ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, message)
	_ca = append(_ca, a...)
	_m.Called(_ca...)
}

type mockConstructorTestingTNewLogger interface {
	mock.TestingT
	Cleanup(func())
}

// NewLogger creates a new instance of Logger. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewLogger(t mockConstructorTestingTNewLogger) *Logger {
	mock := &Logger{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}