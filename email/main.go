package email

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/resend/resend-go/v3"
)

type EmailService interface {
	SendEmail(toEmail, subject, body string) error
	PrepareEmailBody(templateName string, data any) (string, error)
}

type emailService struct {
	client        *resend.Client
	fromEmail     string
	fromName      string
	templatesPath string
}

type Config struct {
	ResendAPIKey  string
	FromEmail     string
	FromName      string
	TemplatesPath string
}

func NewEmailService(cfg *Config) EmailService {
	if cfg == nil {
		panic("email: config must not be nil")
	}
	return &emailService{
		client:        resend.NewClient(cfg.ResendAPIKey),
		fromEmail:     cfg.FromEmail,
		fromName:      cfg.FromName,
		templatesPath: cfg.TemplatesPath,
	}
}

func (s *emailService) SendEmail(toEmail, subject, body string) error {
	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		To:      []string{toEmail},
		Html:    body,
		Subject: subject,
	})
	return err
}

func (s *emailService) PrepareEmailBody(templateName string, data any) (string, error) {
	if strings.Contains(templateName, "..") || strings.ContainsAny(templateName, `/\`) {
		return "", fmt.Errorf("invalid template name: %q", templateName)
	}
	path := filepath.Join(s.templatesPath, templateName+".html")
	rel, err := filepath.Rel(s.templatesPath, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid template name: %q", templateName)
	}
	htmlData, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	htmlTemplateName := fmt.Sprintf("%s.email.html", templateName)
	htmlTemplate, err := template.New(htmlTemplateName).Parse(string(htmlData))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %q: %w", templateName, err)
	}

	var templateBuffer bytes.Buffer
	if err := htmlTemplate.ExecuteTemplate(&templateBuffer, htmlTemplateName, data); err != nil {
		return "", fmt.Errorf("failed to execute template %q: %w", templateName, err)
	}

	return templateBuffer.String(), nil
}
