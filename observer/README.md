# Observer

Prometheus metrics and **QueryMetricsHook** for database observability. Provides HTTP request/duration counters and histograms, DB query duration histograms, and an interface for the repository layer to record query metrics without depending on a specific metrics implementation. Used by go-api for `/metrics` and repository instrumentation.

## Structure

| File        | Role                                                                 |
| ----------- | -------------------------------------------------------------------- |
| **main.go** | `BaseMetrics` (HTTP total/duration, DB query duration), `NewBaseMetrics(namespace, appName)`, `QueryMetricsHook` interface, `EchoMiddleware`, `RegisterCollectors`. |
| **main_test.go** | Tests for middleware and hook behaviour.                             |

## Components

- **BaseMetrics** — Holds Prometheus counter and histogram vecs for HTTP and DB. Register with `NewBaseMetrics`; use `EchoMiddleware(skipper)` for HTTP, and pass a hook implementation to the repository.
- **QueryMetricsHook** — `RecordQuery(operation, table, queryType, status, duration)`. Repository package calls this on each query; implement with `BaseMetrics.DBQueryDuration` so DB latency is exposed.

## Usage

```go
import "github.com/yca-software/go-common/observer"

metrics, _ := observer.NewBaseMetrics("app", "go-api")
e.Use(metrics.EchoMiddleware(nil))
// In repo layer: wrap DB with a hook that calls metrics.RecordQuery(...)
```

Use the same `QueryMetricsHook` when constructing repositories so all queries are recorded.
