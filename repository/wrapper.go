package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	observer "github.com/yca-software/go-common/observer"
)

// DBWrapper wraps a sqlx.DB to track database query metrics
// It embeds *sqlx.DB so it can be used as a drop-in replacement
type DBWrapper struct {
	*sqlx.DB
	hook observer.QueryMetricsHook
}

func NewDBWrapper(db *sqlx.DB, hook observer.QueryMetricsHook) *DBWrapper {
	if db == nil {
		panic("repository: db must not be nil")
	}
	return &DBWrapper{
		DB:   db,
		hook: hook,
	}
}

func (w *DBWrapper) Beginx() (*TxWrapper, error) {
	tx, err := w.DB.Beginx()
	if err != nil {
		return nil, err
	}
	return NewTxWrapper(tx, w.hook), nil
}

func (w *DBWrapper) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*TxWrapper, error) {
	tx, err := w.DB.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return NewTxWrapper(tx, w.hook), nil
}

func (w *DBWrapper) Exec(query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := w.DB.Exec(query, args...)
	recordMetrics(w.hook, query, "exec", start, err)
	return result, err
}

func (w *DBWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := w.DB.ExecContext(ctx, query, args...)
	recordMetrics(w.hook, query, "exec", start, err)
	return result, err
}

func (w *DBWrapper) Get(dest any, query string, args ...any) error {
	start := time.Now()
	err := w.DB.Get(dest, query, args...)
	recordMetrics(w.hook, query, "get", start, err)
	return err
}

func (w *DBWrapper) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	start := time.Now()
	err := w.DB.GetContext(ctx, dest, query, args...)
	recordMetrics(w.hook, query, "get", start, err)
	return err
}

func (w *DBWrapper) Select(dest any, query string, args ...any) error {
	start := time.Now()
	err := w.DB.Select(dest, query, args...)
	recordMetrics(w.hook, query, "select", start, err)
	return err
}

func (w *DBWrapper) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	start := time.Now()
	err := w.DB.SelectContext(ctx, dest, query, args...)
	recordMetrics(w.hook, query, "select", start, err)
	return err
}

func (w *DBWrapper) Query(query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := w.DB.Query(query, args...)
	recordMetrics(w.hook, query, "query", start, err)
	return rows, err
}

func (w *DBWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := w.DB.QueryContext(ctx, query, args...)
	recordMetrics(w.hook, query, "query", start, err)
	return rows, err
}

func (w *DBWrapper) Queryx(query string, args ...any) (*sqlx.Rows, error) {
	start := time.Now()
	rows, err := w.DB.Queryx(query, args...)
	recordMetrics(w.hook, query, "query", start, err)
	return rows, err
}

func (w *DBWrapper) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	start := time.Now()
	rows, err := w.DB.QueryxContext(ctx, query, args...)
	recordMetrics(w.hook, query, "query", start, err)
	return rows, err
}

type TxWrapper struct {
	*sqlx.Tx
	hook observer.QueryMetricsHook
}

func NewTxWrapper(tx *sqlx.Tx, hook observer.QueryMetricsHook) *TxWrapper {
	if tx == nil {
		panic("repository: tx must not be nil")
	}
	return &TxWrapper{
		Tx:   tx,
		hook: hook,
	}
}

func (w *TxWrapper) Exec(query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := w.Tx.Exec(query, args...)
	recordMetrics(w.hook, query, "exec", start, err)
	return result, err
}

func (w *TxWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := w.Tx.ExecContext(ctx, query, args...)
	recordMetrics(w.hook, query, "exec", start, err)
	return result, err
}

func (w *TxWrapper) Get(dest any, query string, args ...any) error {
	start := time.Now()
	err := w.Tx.Get(dest, query, args...)
	recordMetrics(w.hook, query, "get", start, err)
	return err
}

func (w *TxWrapper) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	start := time.Now()
	err := w.Tx.GetContext(ctx, dest, query, args...)
	recordMetrics(w.hook, query, "get", start, err)
	return err
}

func (w *TxWrapper) Select(dest any, query string, args ...any) error {
	start := time.Now()
	err := w.Tx.Select(dest, query, args...)
	recordMetrics(w.hook, query, "select", start, err)
	return err
}

func (w *TxWrapper) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	start := time.Now()
	err := w.Tx.SelectContext(ctx, dest, query, args...)
	recordMetrics(w.hook, query, "select", start, err)
	return err
}

func (w *TxWrapper) Query(query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := w.Tx.Query(query, args...)
	recordMetrics(w.hook, query, "query", start, err)
	return rows, err
}

func (w *TxWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := w.Tx.QueryContext(ctx, query, args...)
	recordMetrics(w.hook, query, "query", start, err)
	return rows, err
}

func (w *TxWrapper) Queryx(query string, args ...any) (*sqlx.Rows, error) {
	start := time.Now()
	rows, err := w.Tx.Queryx(query, args...)
	recordMetrics(w.hook, query, "query", start, err)
	return rows, err
}

func (w *TxWrapper) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	start := time.Now()
	rows, err := w.Tx.QueryxContext(ctx, query, args...)
	recordMetrics(w.hook, query, "query", start, err)
	return rows, err
}
