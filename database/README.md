# Database

SQL client for PostgreSQL via **pgx** (stdlib driver) and **sqlx**. Provides connection pooling, configurable timeouts, and optional ping on creation. Used by applications (e.g. go-api) to obtain `*sqlx.DB` for repositories and migrations.

## Structure

| File        | Role                                                                 |
| ----------- | -------------------------------------------------------------------- |
| **main.go** | `SQLClientConfig`, `NewSQLClient(cfg)`, defaults for pool and ping.  |
| **main_test.go** | Tests using testcontainers (Postgres). Requires Docker.         |

## Config

- **DriverName** — e.g. `pgx` (stdlib); must not be empty.
- **DSN** — Connection string; must not be empty.
- **MaxOpenConns**, **MaxIdleConns**, **ConnMaxLifetime**, **ConnMaxIdleTime** — Pool settings; zero values use package defaults.
- **PingTimeout** — Optional; when set, `NewSQLClient` pings the DB with context timeout.

## Usage

```go
import "github.com/yca-software/go-common/database"

db, err := database.NewSQLClient(database.SQLClientConfig{
    DriverName: "pgx",
    DSN:        os.Getenv("DATABASE_POSTGRES_DSN"),
})
// use db with repository package or migrations
```

Errors: `ErrEmptyDriverName`, `ErrEmptyDSN` when config is invalid.
