package observer_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	observer "github.com/yca-software/go-common/observer"
)

type MetricsTestSuite struct {
	suite.Suite
}

func TestMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsTestSuite))
}

func (s *MetricsTestSuite) SetupTest() {
	// Clean up before each test
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
}

func (s *MetricsTestSuite) TestNewBaseMetrics() {
	metrics, err := observer.NewBaseMetrics("test", "testapp")
	require.NoError(s.T(), err)
	s.NotNil(metrics)
	s.NotNil(metrics.HTTPRequestsTotal)
	s.NotNil(metrics.HTTPRequestDuration)
	s.NotNil(metrics.DBQueryDuration)
}

func (s *MetricsTestSuite) TestEchoMiddleware() {
	metrics, err := observer.NewBaseMetrics("test", "testapp")
	require.NoError(s.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.Use(metrics.EchoMiddleware(nil))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *MetricsTestSuite) TestEchoMiddleware_WithSkipper() {
	metrics, err := observer.NewBaseMetrics("test", "testapp")
	require.NoError(s.T(), err)

	skipper := func(c echo.Context) bool {
		return c.Path() == "/metrics"
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	e.Use(metrics.EchoMiddleware(skipper))
	e.GET("/metrics", func(c echo.Context) error {
		return c.String(http.StatusOK, "metrics")
	})

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
}

func (s *MetricsTestSuite) TestEchoMiddleware_ErrorResponse() {
	metrics, err := observer.NewBaseMetrics("test", "testapp")
	require.NoError(s.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()

	e.Use(metrics.EchoMiddleware(nil))
	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "error")
	})

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusInternalServerError, rec.Code)
}

func (s *MetricsTestSuite) TestGetQueryMetricsHook() {
	metrics, err := observer.NewBaseMetrics("test", "testapp")
	require.NoError(s.T(), err)

	hook := metrics.GetQueryMetricsHook()
	s.NotNil(hook)

	// Test that the hook can record queries
	hook.RecordQuery("select", "users", "select_users", "success", 0.001)
	// Should not panic
}

func (s *MetricsTestSuite) TestRegisterCollectors() {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_counter",
			Help: "Test counter",
		},
		[]string{"label"},
	)

	err := observer.RegisterCollectors(counter)
	s.NoError(err)

	// Registering again should handle AlreadyRegisteredError
	err = observer.RegisterCollectors(counter)
	s.NoError(err)
}

func (s *MetricsTestSuite) TestRegisterCollectors_NilCollector() {
	err := observer.RegisterCollectors(nil)
	s.NoError(err, "Nil collector should be skipped")
}

func (s *MetricsTestSuite) TestRegisterCollectors_MultipleCollectors() {
	counter1 := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "counter1"},
		[]string{"label"},
	)
	counter2 := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "counter2"},
		[]string{"label"},
	)

	err := observer.RegisterCollectors(counter1, counter2)
	s.NoError(err)
}

func (s *MetricsTestSuite) TestQueryMetricsHook_RecordQuery() {
	metrics, err := observer.NewBaseMetrics("test", "testapp")
	require.NoError(s.T(), err)

	hook := metrics.GetQueryMetricsHook()

	// Record various query types
	hook.RecordQuery("select", "users", "select_users", "success", 0.001)
	hook.RecordQuery("insert", "users", "insert_users", "success", 0.002)
	hook.RecordQuery("update", "users", "update_users", "error", 0.003)
	hook.RecordQuery("delete", "users", "delete_users", "success", 0.001)

	// Should not panic
	s.NotNil(hook)
}
