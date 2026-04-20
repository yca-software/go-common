package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type LogData struct {
	Level     string
	RequestID string
	UserID    string
	APIKeyID  string
	Location  string
	Error     error
	Message   string
	Data      any
}

type Logger interface {
	Log(data LogData)
}

type loggerImpl struct {
	logger zerolog.Logger
}

func New() Logger {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	zlog := zerolog.New(output).With().Timestamp().Logger()

	return &loggerImpl{
		logger: zlog,
	}
}

// NewWithWriter creates a Logger that writes JSON to w. Used for testing so log output can be captured and asserted.
func NewWithWriter(w io.Writer) Logger {
	zlog := zerolog.New(w).With().Timestamp().Logger()
	return &loggerImpl{
		logger: zlog,
	}
}

func (l *loggerImpl) Log(data LogData) {
	var event *zerolog.Event
	switch data.Level {
	case "info":
		event = l.logger.Info()
	case "error":
		event = l.logger.Error()
	case "debug":
		event = l.logger.Debug()
	case "warn":
		event = l.logger.Warn()
	case "fatal":
		event = l.logger.Fatal()
	case "panic":
		event = l.logger.Panic()
	default:
		event = l.logger.Info()
	}

	event = event.
		Str("req_id", data.RequestID).
		Str("location", data.Location)

	if data.UserID != "" {
		event = event.Str("req_user_id", data.UserID)
	}
	if data.APIKeyID != "" {
		event = event.Str("req_api_key_id", data.APIKeyID)
	}

	if data.Error != nil {
		event = event.Err(SanitizeError(data.Error))
	}

	if data.Data != nil {
		event = event.Any("data", SanitizeData(data.Data))
	}

	event.Msg(redactString(data.Message))
}
