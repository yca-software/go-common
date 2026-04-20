package email_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	email "github.com/yca-software/go-common/email"
)

type EmailTestSuite struct {
	suite.Suite
	emailService email.EmailService
}

func TestEmailTestSuite(t *testing.T) {
	suite.Run(t, new(EmailTestSuite))
}

func (s *EmailTestSuite) SetupSuite() {
	s.emailService = email.NewEmailService(&email.Config{
		ResendAPIKey:  "test-key",
		FromEmail:     "test@example.com",
		FromName:      "Test App",
		TemplatesPath: "testdata",
	})
}

func (s *EmailTestSuite) TestPrepareEmailBody() {
	data := map[string]any{
		"Name":    "John",
		"AppName": "My App",
	}

	body, err := s.emailService.PrepareEmailBody("welcome", data)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), body, "Hello John")
	assert.Contains(s.T(), body, "Welcome to My App")
}

func (s *EmailTestSuite) TestPrepareEmailBody_MissingTemplate() {
	_, err := s.emailService.PrepareEmailBody("nonexistent", map[string]any{})
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "failed to read template")
}

func (s *EmailTestSuite) TestPrepareEmailBody_EmptyData() {
	body, err := s.emailService.PrepareEmailBody("simple", map[string]any{})
	require.NoError(s.T(), err)
	assert.Contains(s.T(), body, "Hello")
}

func (s *EmailTestSuite) TestPrepareEmailBody_ComplexTemplate() {
	data := map[string]any{
		"Title": "Welcome",
		"User": map[string]any{
			"Name":  "John",
			"Email": "john@example.com",
		},
		"ShowLink": true,
		"Link":     "https://example.com",
	}

	body, err := s.emailService.PrepareEmailBody("complex", data)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), body, "Welcome")
	assert.Contains(s.T(), body, "User: John")
	assert.Contains(s.T(), body, "Email: john@example.com")
	assert.Contains(s.T(), body, "Click here")
	assert.Contains(s.T(), body, "https://example.com")
}

func (s *EmailTestSuite) TestPrepareEmailBody_InvalidTemplateSyntax() {
	_, err := s.emailService.PrepareEmailBody("invalid", map[string]any{"Name": "John"})
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "failed to parse template")
}

func (s *EmailTestSuite) TestNewEmailService_NilConfig_Panics() {
	assert.Panics(s.T(), func() {
		email.NewEmailService(nil)
	})
}

func (s *EmailTestSuite) TestPrepareEmailBody_RejectsPathTraversal() {
	_, err := s.emailService.PrepareEmailBody("../../../etc/passwd", map[string]any{})
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid template name")

	_, err = s.emailService.PrepareEmailBody("..", map[string]any{})
	assert.Error(s.T(), err)

	_, err = s.emailService.PrepareEmailBody("welcome/../other", map[string]any{})
	assert.Error(s.T(), err)
}
