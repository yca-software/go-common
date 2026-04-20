package observer

import (
	"fmt"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

type BaseMetrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	DBQueryDuration     *prometheus.HistogramVec
}

// QueryMetricsHook is an interface for recording database query metrics
// Implementations can use this to track query performance without the repository
// package directly depending on a specific metrics implementation
type QueryMetricsHook interface {
	RecordQuery(operation, table, queryType, status string, duration float64)
}

func NewBaseMetrics(namespace, appName string) (*BaseMetrics, error) {
	m := &BaseMetrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "http",
				Name:        "requests_total",
				Help:        "Total number of HTTP requests processed by the API.",
				ConstLabels: prometheus.Labels{"app": appName},
			},
			[]string{"method", "route", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Subsystem:   "http",
				Name:        "request_duration_seconds",
				Help:        "Duration of HTTP requests processed by the API.",
				ConstLabels: prometheus.Labels{"app": appName},
				Buckets:     prometheus.DefBuckets,
			},
			[]string{"method", "route", "status"},
		),
		DBQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Subsystem:   "db",
				Name:        "query_duration_seconds",
				Help:        "Duration of database queries in seconds.",
				ConstLabels: prometheus.Labels{"app": appName},
				Buckets:     []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"operation", "table", "query_type", "status"},
		),
	}

	collectors := []prometheus.Collector{
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.DBQueryDuration,
	}

	if err := RegisterCollectors(collectors...); err != nil {
		return nil, err
	}

	return m, nil
}

// EchoMiddleware returns an Echo middleware function that records HTTP metrics
func (m *BaseMetrics) EchoMiddleware(skipper func(echo.Context) bool) echo.MiddlewareFunc {
	if skipper == nil {
		skipper = func(c echo.Context) bool {
			return c.Path() == "/metrics"
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			start := time.Now()
			err := next(c)

			status := c.Response().Status
			route := c.Path()
			if route == "" {
				route = "unmatched"
			}

			labels := []string{c.Request().Method, route, strconv.Itoa(status)}

			if m.HTTPRequestsTotal != nil {
				m.HTTPRequestsTotal.WithLabelValues(labels...).Inc()
			}

			if m.HTTPRequestDuration != nil {
				m.HTTPRequestDuration.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
			}

			return err
		}
	}
}

// GetQueryMetricsHook returns a QueryMetricsHook implementation for repository metrics
func (m *BaseMetrics) GetQueryMetricsHook() QueryMetricsHook {
	return &queryMetricsHook{dbQueryDuration: m.DBQueryDuration}
}

type queryMetricsHook struct {
	dbQueryDuration *prometheus.HistogramVec
}

func (h *queryMetricsHook) RecordQuery(operation, table, queryType, status string, duration float64) {
	if h.dbQueryDuration != nil {
		h.dbQueryDuration.WithLabelValues(operation, table, queryType, status).Observe(duration)
	}
}

// RegisterCollectors registers Prometheus collectors, handling AlreadyRegisteredError gracefully
func RegisterCollectors(cs ...prometheus.Collector) error {
	for _, c := range cs {
		if c == nil {
			continue
		}
		// Try to register - if already registered, unregister the old one first
		if err := prometheus.Register(c); err != nil {
			if already, ok := err.(prometheus.AlreadyRegisteredError); ok {
				// Unregister the existing collector and register our new one
				if !prometheus.Unregister(already.ExistingCollector) {
					return fmt.Errorf("failed to unregister existing collector")
				}
				if err := prometheus.Register(c); err != nil {
					return fmt.Errorf("re-registering collector after unregister: %w", err)
				}
				continue
			}
			return fmt.Errorf("registering collector: %w", err)
		}
	}
	return nil
}
