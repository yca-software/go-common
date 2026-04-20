# Repository

Generic repository layer over **sqlx** and **squirrel**: `Repository[T]` with base CRUD and pagination, optional **QueryMetricsHook** for observability, and error wrapping. Applications (e.g. go-api) embed this in concrete repo structs per entity and add domain-specific methods. Reduces boilerplate and keeps query building consistent.

## Structure

| File          | Role                                                                 |
| ------------- | -------------------------------------------------------------------- |
| **main.go**   | `Executor`, `Tx`, `Repository[T]` interface, `NewRepository[T](db, tableName, columns, metricsHook)`, `BaseGet`, `BaseSelect`, `BasePaginatedSelect`, `BaseCount`, `BaseCreate`, `BaseCreateMany`, `BaseUpdate`, `BaseDelete`. |
| **wrapper.go** | `DBWrapper` — wraps sqlx.DB and Tx, records duration and calls `QueryMetricsHook`. |
| **metrics.go** | Hook invocation from wrapper (operation, table, query type, status).  |
| **error.go**   | Repository-specific errors (e.g. `ErrConditionRequired`).             |
| **mock.go**   | `MockRepository[T]` for tests.                                       |
| **main_test.go** | Tests (with testcontainers when needed).                         |

## Concepts

- **Executor** — Interface for running queries (DB or Tx). Use `repo.DB()` for non-tx and `repo.BeginTx()` for transactions.
- **Tx** — Executor with `Commit` / `Rollback`; pass to base methods when inside a transaction.
- **Base\*** — Methods take `Executor` (nil = use repo’s DB), **squirrel.Sqlizer** condition, and optional column list/sort/limit/offset. Return single entity, slice, count, or error.

## Usage

```go
import yca_repository "github.com/yca-software/go-common/repository"

type repo struct {
    yca_repository.Repository[entities.User]
}
func New(db *sqlx.DB, hook observer.QueryMetricsHook) Repository {
    return &repo{
        Repository: yca_repository.NewRepository[entities.User](db, "users", columns, hook),
    }
}
// In method:
user, err := r.BaseGet(r.DB(), squirrel.Eq{"id": id}, nil)
```

Use **squirrel** with `PlaceholderFormat(squirrel.Dollar)` for Postgres. For new repos, embed `Repository[T]`, add your interface and constructor, and implement domain methods using the base helpers.
