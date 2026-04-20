package logger

import "github.com/stretchr/testify/mock"

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Log(data LogData) {
	m.Called(data)
}
