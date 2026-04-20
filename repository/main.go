package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	observer "github.com/yca-software/go-common/observer"
)

// Executor runs queries (either a wrapped DB or a wrapped Tx).
type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Get(dest any, query string, args ...any) error
	Select(dest any, query string, args ...any) error
}

// Tx is an Executor that can be committed or rolled back.
// Use DB() for non-tx work and BeginTx() for a transaction.
type Tx interface {
	Executor
	Commit() error
	Rollback() error
}

type repository[T any] struct {
	db        *DBWrapper
	tableName string
	columns   []string
	metrics   observer.QueryMetricsHook
}

type Repository[T any] interface {
	DB() Executor
	GetDB() *sqlx.DB
	BeginTx() (Tx, error)
	GetQueryBuilder() squirrel.StatementBuilderType

	BaseCount(exec Executor, condition squirrel.Sqlizer) (int, error)
	BaseGet(exec Executor, condition squirrel.Sqlizer, columns *[]string) (*T, error)
	BaseSelect(exec Executor, condition squirrel.Sqlizer, columns *[]string, sort string) (*[]T, error)
	BasePaginatedSelect(exec Executor, condition squirrel.Sqlizer, columns *[]string, sort string, limit, offset uint64) (*[]T, error)
	BaseCreate(exec Executor, data map[string]any) error
	BaseCreateMany(exec Executor, columns []string, data []map[string]any, ignoreConflict bool) error
	BaseDelete(exec Executor, condition squirrel.Sqlizer) error
	BaseUpdate(exec Executor, condition squirrel.Sqlizer, data map[string]any) error
}

var ErrConditionRequired = errors.New("repository: condition is required for Get, Delete, and Update")

func NewRepository[T any](db *sqlx.DB, tableName string, columns []string, metricsHook observer.QueryMetricsHook) Repository[T] {
	if db == nil {
		panic("repository: db must not be nil")
	}
	if tableName == "" {
		panic("repository: table name must not be empty")
	}
	if len(columns) == 0 {
		panic("repository: columns must not be empty")
	}
	wrappedDB := NewDBWrapper(db, metricsHook)
	return &repository[T]{
		db:        wrappedDB,
		tableName: tableName,
		columns:   columns,
		metrics:   metricsHook,
	}
}

func (r *repository[T]) DB() Executor {
	return r.db
}

func (r *repository[T]) GetDB() *sqlx.DB {
	return r.db.DB
}

func (r *repository[T]) BeginTx() (Tx, error) {
	return r.db.Beginx()
}

func (r *repository[T]) GetQueryBuilder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

// resolveExec returns the wrapped DB when exec is nil (use repo's connection), otherwise returns exec (tx).
func (r *repository[T]) resolveExec(exec Executor) Executor {
	if exec == nil {
		return r.db
	}
	return exec
}

func (r *repository[T]) BaseCount(exec Executor, condition squirrel.Sqlizer) (int, error) {
	query := r.GetQueryBuilder().
		Select("count(*)").
		From(r.tableName)

	if condition != nil {
		query = query.Where(condition)
	}

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return 0, WrapSQLError(err)
	}

	start := time.Now()
	var count int
	if err = r.resolveExec(exec).Get(&count, sqlStr, args...); err != nil {
		if r.metrics != nil {
			recordMetrics(r.metrics, sqlStr, "count", start, err)
		}
		return 0, WrapSQLError(err)
	}
	if r.metrics != nil {
		recordMetrics(r.metrics, sqlStr, "count", start, nil)
	}
	return count, nil
}

func (r *repository[T]) BaseGet(exec Executor, condition squirrel.Sqlizer, columns *[]string) (*T, error) {
	if condition == nil {
		return nil, ErrConditionRequired
	}
	cols := r.columns
	if columns != nil {
		cols = *columns
	}

	sqlStr, args, err := r.GetQueryBuilder().Select(cols...).From(r.tableName).Where(condition).ToSql()
	if err != nil {
		return nil, WrapSQLError(err)
	}

	start := time.Now()
	result := new(T)
	if err = r.resolveExec(exec).Get(result, sqlStr, args...); err != nil {
		if r.metrics != nil {
			recordMetrics(r.metrics, sqlStr, "get", start, err)
		}
		return nil, WrapSQLError(err)
	}
	if r.metrics != nil {
		recordMetrics(r.metrics, sqlStr, "get", start, nil)
	}
	return result, nil
}

func (r *repository[T]) BaseSelect(exec Executor, condition squirrel.Sqlizer, columns *[]string, sort string) (*[]T, error) {
	cols := r.columns
	if columns != nil {
		cols = *columns
	}

	query := r.GetQueryBuilder().Select(cols...).From(r.tableName)
	if condition != nil {
		query = query.Where(condition)
	}
	if sort != "" {
		orderBys, sortErr := r.buildSafeOrderBy(sort)
		if sortErr != nil {
			return nil, WrapSQLError(sortErr)
		}
		query = query.OrderBy(orderBys...)
	}

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, WrapSQLError(err)
	}

	start := time.Now()
	var results []T
	if err = r.resolveExec(exec).Select(&results, sqlStr, args...); err != nil {
		if r.metrics != nil {
			recordMetrics(r.metrics, sqlStr, "select", start, err)
		}
		return nil, WrapSQLError(err)
	}
	if r.metrics != nil {
		recordMetrics(r.metrics, sqlStr, "select", start, nil)
	}
	return &results, nil
}

func (r *repository[T]) BasePaginatedSelect(exec Executor, condition squirrel.Sqlizer, columns *[]string, sort string, limit, offset uint64) (*[]T, error) {
	cols := r.columns
	if columns != nil {
		cols = *columns
	}

	query := r.GetQueryBuilder().Select(cols...).From(r.tableName)
	if condition != nil {
		query = query.Where(condition)
	}
	if sort != "" {
		orderBys, sortErr := r.buildSafeOrderBy(sort)
		if sortErr != nil {
			return nil, WrapSQLError(sortErr)
		}
		query = query.OrderBy(orderBys...)
	}
	query = query.Limit(limit).Offset(offset)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, WrapSQLError(err)
	}
	start := time.Now()
	var results []T
	if err = r.resolveExec(exec).Select(&results, sqlStr, args...); err != nil {
		if r.metrics != nil {
			recordMetrics(r.metrics, sqlStr, "select", start, err)
		}
		return nil, WrapSQLError(err)
	}
	if r.metrics != nil {
		recordMetrics(r.metrics, sqlStr, "select", start, nil)
	}
	return &results, nil
}

func (r *repository[T]) BaseCreate(exec Executor, data map[string]any) error {
	if len(data) == 0 {
		return WrapSQLError(errors.New("repository: create data must not be empty"))
	}
	sqlStr, args, err := r.GetQueryBuilder().Insert(r.tableName).SetMap(data).ToSql()
	if err != nil {
		return WrapSQLError(err)
	}

	start := time.Now()
	_, err = r.resolveExec(exec).Exec(sqlStr, args...)
	if r.metrics != nil {
		recordMetrics(r.metrics, sqlStr, "exec", start, err)
	}
	return WrapSQLError(err)
}

func (r *repository[T]) BaseCreateMany(exec Executor, columns []string, data []map[string]any, ignoreConflict bool) error {
	if len(data) == 0 || len(columns) == 0 {
		return nil
	}

	query := r.GetQueryBuilder().Insert(r.tableName).Columns(columns...)
	for _, row := range data {
		rowValues := make([]any, len(columns))
		for i, col := range columns {
			rowValues[i] = row[col]
		}
		query = query.Values(rowValues...)
	}
	if ignoreConflict {
		query = query.Suffix("ON CONFLICT DO NOTHING")
	}

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return WrapSQLError(err)
	}
	start := time.Now()
	_, err = r.resolveExec(exec).Exec(sqlStr, args...)
	if r.metrics != nil {
		recordMetrics(r.metrics, sqlStr, "exec", start, err)
	}
	return WrapSQLError(err)
}

func (r *repository[T]) BaseDelete(exec Executor, condition squirrel.Sqlizer) error {
	if condition == nil {
		return ErrConditionRequired
	}
	query := r.GetQueryBuilder().Delete(r.tableName).Where(condition)
	sqlStr, args, err := query.ToSql()
	if err != nil {
		return WrapSQLError(err)
	}
	start := time.Now()
	result, err := r.resolveExec(exec).Exec(sqlStr, args...)
	if r.metrics != nil {
		recordMetrics(r.metrics, sqlStr, "exec", start, err)
	}
	if err != nil {
		return WrapSQLError(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return WrapSQLError(err)
	}
	if rowsAffected == 0 {
		return ErrNotFoundNoRowsAffected()
	}
	return nil
}

func (r *repository[T]) buildSafeOrderBy(sort string) ([]string, error) {
	allowedColumns := make(map[string]string, len(r.columns))
	for _, column := range r.columns {
		normalized := normalizeSortIdentifier(column)
		if normalized == "" {
			continue
		}
		allowedColumns[normalized] = normalized
	}

	parts := strings.Split(sort, ",")
	orderBys := make([]string, 0, len(parts))
	for _, part := range parts {
		clause := strings.TrimSpace(part)
		if clause == "" {
			continue
		}

		tokens := strings.Fields(clause)
		if len(tokens) == 0 || len(tokens) > 2 {
			return nil, fmt.Errorf("repository: invalid sort clause: %q", clause)
		}

		column := normalizeSortIdentifier(tokens[0])
		safeColumn, ok := allowedColumns[column]
		if !ok {
			return nil, fmt.Errorf("repository: invalid sort column: %q", tokens[0])
		}

		direction := "ASC"
		if len(tokens) == 2 {
			switch strings.ToUpper(tokens[1]) {
			case "ASC", "DESC":
				direction = strings.ToUpper(tokens[1])
			default:
				return nil, fmt.Errorf("repository: invalid sort direction: %q", tokens[1])
			}
		}

		orderBys = append(orderBys, fmt.Sprintf("%s %s", safeColumn, direction))
	}

	if len(orderBys) == 0 {
		return nil, errors.New("repository: sort must not be empty")
	}

	return orderBys, nil
}

func normalizeSortIdentifier(identifier string) string {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return ""
	}

	parts := strings.Split(trimmed, ".")
	last := strings.TrimSpace(parts[len(parts)-1])
	last = strings.Trim(last, `"`)
	return strings.ToLower(last)
}

func (r *repository[T]) BaseUpdate(exec Executor, condition squirrel.Sqlizer, data map[string]any) error {
	if condition == nil {
		return ErrConditionRequired
	}
	if len(data) == 0 {
		return WrapSQLError(errors.New("repository: update data must not be empty"))
	}
	query := r.GetQueryBuilder().Update(r.tableName).SetMap(data).Where(condition)
	sqlStr, args, err := query.ToSql()
	if err != nil {
		return WrapSQLError(err)
	}
	start := time.Now()
	result, err := r.resolveExec(exec).Exec(sqlStr, args...)
	if r.metrics != nil {
		recordMetrics(r.metrics, sqlStr, "exec", start, err)
	}
	if err != nil {
		return WrapSQLError(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return WrapSQLError(err)
	}
	if rowsAffected == 0 {
		return ErrNotFoundNoRowsAffected()
	}
	return nil
}
