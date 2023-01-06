// Note: The following tests correspond to a previous version, where mocking each part was not possible.
// The code was refactored for better maintainability and testability, so these tests are temporary and can be
// improved using mock implementations.
package progressbar

import (
	"github.com/stretchr/testify/mock"
	"io"
)

type MockedProgressBar struct {
	mock.Mock
}

func (m *MockedProgressBar) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockedProgressBar) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockedProgressBar) Add(total int64, units Units, prefix string) io.Writer {
	args := m.Called(total, units, prefix)
	return args.Get(0).(io.Writer)
}
