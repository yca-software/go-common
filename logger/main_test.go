package logger_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	logger "github.com/yca-software/go-common/logger"
)

type LoggerTestSuite struct {
	suite.Suite
	logger logger.Logger
	buf    *bytes.Buffer
}

func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

func (s *LoggerTestSuite) SetupTest() {
	s.buf = &bytes.Buffer{}
	s.logger = logger.NewWithWriter(s.buf)
}

// parsedLog holds the JSON fields we assert on.
type parsedLog struct {
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	RequestID string         `json:"req_id"`
	Location  string         `json:"location"`
	UserID    string         `json:"req_user_id"`
	APIKeyID  string         `json:"req_api_key_id"`
	Error     string         `json:"error"`
	Data      map[string]any `json:"data"`
}

func (s *LoggerTestSuite) parseLog() parsedLog {
	s.Require().NotEmpty(s.buf.Bytes(), "expected at least one log line")
	dec := json.NewDecoder(s.buf)
	var out parsedLog
	err := dec.Decode(&out)
	s.Require().NoError(err)
	return out
}

func (s *LoggerTestSuite) TestNew() {
	log := logger.New()
	s.NotNil(log)
}

func (s *LoggerTestSuite) TestLog_Info() {
	data := logger.LogData{
		Level:   "info",
		Message: "Test info message",
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("Test info message", p.Message)
	s.Equal("info", p.Level)
}

func (s *LoggerTestSuite) TestLog_Error() {
	data := logger.LogData{
		Level:   "error",
		Message: "Test error message",
		Error:   errors.New("test error"),
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("Test error message", p.Message)
	s.Equal("error", p.Level)
	s.Equal("test error", p.Error)
}

func (s *LoggerTestSuite) TestLog_WithRequestID() {
	data := logger.LogData{
		Level:     "info",
		Message:   "Test message",
		RequestID: "test-request-123",
		Location:  "test.go:42",
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("Test message", p.Message)
	s.Equal("test-request-123", p.RequestID)
	s.Equal("test.go:42", p.Location)
}

func (s *LoggerTestSuite) TestLog_WithUserID() {
	data := logger.LogData{
		Level:   "info",
		Message: "Test message",
		UserID:  "user-123",
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("Test message", p.Message)
	s.Equal("user-123", p.UserID)
}

func (s *LoggerTestSuite) TestLog_WithAPIKeyID() {
	data := logger.LogData{
		Level:    "info",
		Message:  "Test message",
		APIKeyID: "apikey-abc-123",
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("Test message", p.Message)
	s.Equal("apikey-abc-123", p.APIKeyID)
}

func (s *LoggerTestSuite) TestLog_AllLevels() {
	levels := []struct {
		level string
		want  string
	}{
		{"info", "info"},
		{"error", "error"},
		{"debug", "debug"},
		{"warn", "warn"},
		{"unknown", "info"},
	}

	for _, tc := range levels {
		s.buf.Reset()
		data := logger.LogData{
			Level:   tc.level,
			Message: "Test message",
		}
		s.logger.Log(data)
		p := s.parseLog()
		s.Equal("Test message", p.Message, "level %s", tc.level)
		s.Equal(tc.want, p.Level, "level %s", tc.level)
	}
}

func (s *LoggerTestSuite) TestLog_Panic() {
	log := logger.NewWithWriter(s.buf)
	data := logger.LogData{
		Level:   "panic",
		Message: "Test panic message",
	}
	s.Panics(func() {
		log.Log(data)
	}, "Panic level should panic")
}

func (s *LoggerTestSuite) TestLog_WithData() {
	data := logger.LogData{
		Level:   "info",
		Message: "Test message",
		Data: map[string]any{
			"key1": "value1",
			"key2": float64(42),
		},
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("Test message", p.Message)
	s.Require().NotNil(p.Data)
	s.Equal("value1", p.Data["key1"])
	s.Equal(float64(42), p.Data["key2"])
}

func (s *LoggerTestSuite) TestLog_WithLocation() {
	data := logger.LogData{
		Level:    "info",
		Message:  "Test message",
		Location: "service/user.go:123",
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("Test message", p.Message)
	s.Equal("service/user.go:123", p.Location)
}

func (s *LoggerTestSuite) TestLog_WithRequestIDAndUserID() {
	data := logger.LogData{
		Level:     "info",
		Message:   "Test message",
		RequestID: "req-123",
		UserID:    "user-456",
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("Test message", p.Message)
	s.Equal("req-123", p.RequestID)
	s.Equal("user-456", p.UserID)
}

func (s *LoggerTestSuite) TestLog_WithEmptyRequestID() {
	data := logger.LogData{
		Level:     "info",
		Message:   "Test message",
		RequestID: "",
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("Test message", p.Message)
	s.Empty(p.RequestID)
}

func (s *LoggerTestSuite) TestLog_WithEmptyUserID() {
	data := logger.LogData{
		Level:     "info",
		Message:   "Test message",
		RequestID: "req-123",
		UserID:    "",
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("Test message", p.Message)
	s.Empty(p.UserID)
}

func (s *LoggerTestSuite) TestLog_WithEmptyAPIKeyID() {
	data := logger.LogData{
		Level:     "info",
		Message:   "Test message",
		RequestID: "req-123",
		APIKeyID:  "",
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("Test message", p.Message)
	s.Empty(p.APIKeyID)
}

func (s *LoggerTestSuite) TestLog_EmptyMessage() {
	data := logger.LogData{
		Level:   "info",
		Message: "",
	}

	s.logger.Log(data)
	p := s.parseLog()
	s.Empty(p.Message)
}

func (s *LoggerTestSuite) TestLog_ErrorWithSecretRedacted() {
	data := logger.LogData{
		Level:   "error",
		Message: "db failed",
		Error:   errors.New("connect: postgres://user:secretpass@localhost:5432/db"),
	}
	s.logger.Log(data)
	p := s.parseLog()
	s.Equal("error", p.Level)
	s.Contains(p.Error, "[REDACTED]")
	s.NotContains(p.Error, "secretpass")
}

func (s *LoggerTestSuite) TestLog_DataWithSensitiveKeyRedacted() {
	data := logger.LogData{
		Level:   "info",
		Message: "test",
		Data: map[string]any{
			"user_id":  "u1",
			"password": "should-not-appear",
		},
	}
	s.logger.Log(data)
	p := s.parseLog()
	s.Require().NotNil(p.Data)
	s.Equal("u1", p.Data["user_id"])
	s.Equal("[REDACTED]", p.Data["password"])
}
