package database

import (
	"context"
	"errors"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	DefaultMaxOpenConns    = 20
	DefaultMaxIdleConns    = 20
	DefaultConnMaxLifetime = 30 * time.Minute
	DefaultConnMaxIdleTime = 5 * time.Minute
	DefaultPingTimeout     = 5 * time.Second
)

type SQLClientConfig struct {
	DriverName      string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	PingTimeout     time.Duration
}

var (
	ErrEmptyDriverName = errors.New("database: driver name must not be empty")
	ErrEmptyDSN        = errors.New("database: DSN must not be empty")
)

func NewSQLClient(cfg SQLClientConfig) (*sqlx.DB, error) {
	if strings.TrimSpace(cfg.DriverName) == "" {
		return nil, ErrEmptyDriverName
	}
	if strings.TrimSpace(cfg.DSN) == "" {
		return nil, ErrEmptyDSN
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = DefaultMaxOpenConns
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = DefaultMaxIdleConns
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = DefaultConnMaxLifetime
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = DefaultConnMaxIdleTime
	}
	if cfg.PingTimeout == 0 {
		cfg.PingTimeout = DefaultPingTimeout
	}

	client, err := sqlx.Connect(cfg.DriverName, cfg.DSN)
	if err != nil {
		return nil, err
	}

	client.SetMaxOpenConns(cfg.MaxOpenConns)
	client.SetMaxIdleConns(cfg.MaxIdleConns)
	client.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	client.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout)
	defer cancel()

	if err := client.PingContext(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
