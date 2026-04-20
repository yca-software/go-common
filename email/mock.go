package email

import "github.com/stretchr/testify/mock"

type MockEmailService struct {
	mock.Mock
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

func (m *MockEmailService) SendEmail(toEmail, subject, body string) error {
	args := m.Called(toEmail, subject, body)
	return args.Error(0)
}

func (m *MockEmailService) PrepareEmailBody(templateName string, data any) (string, error) {
	args := m.Called(templateName, data)
	return args.String(0), args.Error(1)
}
