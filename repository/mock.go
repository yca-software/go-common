package repository

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
)

// NoopResult implements sql.Result for use with MockTx.
type NoopResult struct{}

func (NoopResult) LastInsertId() (int64, error) { return 0, nil }
func (NoopResult) RowsAffected() (int64, error) { return 0, nil }

// MockTx implements Tx for testing code that uses BeginTx() with mocked repositories.
// Exec, Get, and Select are no-ops so they do not need to be mocked; mock Commit and Rollback in tests.
type MockTx struct {
	mock.Mock
}

func NewMockTx() *MockTx {
	return &MockTx{}
}

func (m *MockTx) Exec(query string, args ...any) (sql.Result, error) {
	return NoopResult{}, nil
}

func (m *MockTx) Get(dest any, query string, args ...any) error {
	return nil
}

func (m *MockTx) Select(dest any, query string, args ...any) error {
	return nil
}

func (m *MockTx) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

type MockRepository[T any] struct {
	mock.Mock
}

func NewMockRepository[T any]() *MockRepository[T] {
	return &MockRepository[T]{}
}

func (m *MockRepository[T]) DB() Executor {
	args := m.Called()
	return args.Get(0).(Executor)
}

func (m *MockRepository[T]) GetDB() *sqlx.DB {
	args := m.Called()
	return args.Get(0).(*sqlx.DB)
}

func (m *MockRepository[T]) BeginTx() (Tx, error) {
	args := m.Called()
	return args.Get(0).(Tx), args.Error(1)
}

func (m *MockRepository[T]) GetQueryBuilder() squirrel.StatementBuilderType {
	args := m.Called()
	return args.Get(0).(squirrel.StatementBuilderType)
}

func (m *MockRepository[T]) BaseCount(exec Executor, condition squirrel.Sqlizer) (int, error) {
	args := m.Called(exec, condition)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockRepository[T]) BaseGet(exec Executor, condition squirrel.Sqlizer, columns *[]string) (*T, error) {
	args := m.Called(exec, condition, columns)
	return args.Get(0).(*T), args.Error(1)
}

func (m *MockRepository[T]) BaseSelect(exec Executor, condition squirrel.Sqlizer, columns *[]string, sort string) (*[]T, error) {
	args := m.Called(exec, condition, columns, sort)
	return args.Get(0).(*[]T), args.Error(1)
}

func (m *MockRepository[T]) BasePaginatedSelect(exec Executor, condition squirrel.Sqlizer, columns *[]string, sort string, limit, offset uint64) (*[]T, error) {
	args := m.Called(exec, condition, columns, sort, limit, offset)
	return args.Get(0).(*[]T), args.Error(1)
}

func (m *MockRepository[T]) BaseCreate(exec Executor, data map[string]any) error {
	args := m.Called(exec, data)
	return args.Error(0)
}

func (m *MockRepository[T]) BaseCreateMany(exec Executor, columns []string, data []map[string]any, ignoreConflict bool) error {
	args := m.Called(exec, columns, data, ignoreConflict)
	return args.Error(0)
}

func (m *MockRepository[T]) BaseDelete(exec Executor, condition squirrel.Sqlizer) error {
	args := m.Called(exec, condition)
	return args.Error(0)
}

func (m *MockRepository[T]) BaseUpdate(exec Executor, condition squirrel.Sqlizer, data map[string]any) error {
	args := m.Called(exec, condition, data)
	return args.Error(0)
}
