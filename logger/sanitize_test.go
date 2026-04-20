package logger_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	logger "github.com/yca-software/go-common/logger"
)

func TestSanitizeError(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		require.Nil(t, logger.SanitizeError(nil))
	})

	t.Run("redacts password in URL", func(t *testing.T) {
		err := errors.New("connect failed: postgres://user:secretpass@localhost:5432/db")
		san := logger.SanitizeError(err)
		require.Contains(t, san.Error(), "[REDACTED]")
		require.NotContains(t, san.Error(), "secretpass")
	})

	t.Run("redacts redis URL", func(t *testing.T) {
		err := errors.New("redis: connection refused redis://:mypassword@localhost:6379")
		san := logger.SanitizeError(err)
		require.Contains(t, san.Error(), "[REDACTED]")
		require.NotContains(t, san.Error(), "mypassword")
	})

	t.Run("preserves safe error message", func(t *testing.T) {
		err := errors.New("duplicate key value violates unique constraint")
		san := logger.SanitizeError(err)
		require.Equal(t, err.Error(), san.Error())
	})
}

func TestSanitizeData(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		require.Nil(t, logger.SanitizeData(nil))
	})

	t.Run("redacts string with DSN", func(t *testing.T) {
		in := "postgres://user:secret@localhost/db"
		out := logger.SanitizeData(in)
		require.Equal(t, "[REDACTED]", out)
	})

	t.Run("redacts sensitive keys in map", func(t *testing.T) {
		in := map[string]any{"user_id": "u1", "password": "secret123", "email": "a@b.com"}
		out := logger.SanitizeData(in).(map[string]any)
		require.Equal(t, "u1", out["user_id"])
		require.Equal(t, "a@b.com", out["email"])
		require.Equal(t, "[REDACTED]", out["password"])
	})

	t.Run("recursively sanitizes nested map", func(t *testing.T) {
		in := map[string]any{"nested": map[string]any{"token": "xyz"}}
		out := logger.SanitizeData(in).(map[string]any)
		nested := out["nested"].(map[string]any)
		require.Equal(t, "[REDACTED]", nested["token"])
	})

	t.Run("sanitizes slice elements", func(t *testing.T) {
		in := []any{"safe", "postgres://user:pass@host/db"}
		out := logger.SanitizeData(in).([]any)
		require.Equal(t, "safe", out[0])
		require.Equal(t, "[REDACTED]", out[1])
	})

	t.Run("returns non-string primitives as-is", func(t *testing.T) {
		require.Equal(t, 42, logger.SanitizeData(42))
		require.True(t, logger.SanitizeData(true).(bool))
	})
}
