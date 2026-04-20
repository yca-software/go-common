package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	database "github.com/yca-software/go-common/database"
)

type DatabaseTestSuite struct {
	suite.Suite
	postgresContainer testcontainers.Container
	testDSN           string
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}

func (s *DatabaseTestSuite) SetupSuite() {
	ctx := context.Background()

	// Spin up PostgreSQL container using testcontainers
	postgresContainer, err := postgres.Run(ctx, "postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(s.T(), err, "Failed to start PostgreSQL container")

	s.postgresContainer = postgresContainer

	// Get connection string from container
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(s.T(), err, "Failed to get connection string from container")

	s.testDSN = connStr
}

func (s *DatabaseTestSuite) TearDownSuite() {
	if s.postgresContainer != nil {
		ctx := context.Background()
		err := s.postgresContainer.Terminate(ctx)
		assert.NoError(s.T(), err, "Failed to terminate PostgreSQL container")
	}
}

func (s *DatabaseTestSuite) TestNewSQLClient_SuccessfulConnection() {
	cfg := database.SQLClientConfig{
		DriverName: "pgx",
		DSN:        s.testDSN,
	}

	db, err := database.NewSQLClient(cfg)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), db)
	defer db.Close()

	// Verify connection is actually working
	err = db.Ping()
	assert.NoError(s.T(), err)
}

func (s *DatabaseTestSuite) TestNewSQLClient_PingTimeout() {
	cfg := database.SQLClientConfig{
		DriverName:  "pgx",
		DSN:         s.testDSN,
		PingTimeout: 1 * time.Nanosecond, // Very short timeout
	}

	// This should fail due to timeout being too short
	_, err := database.NewSQLClient(cfg)
	assert.Error(s.T(), err, "Expected error with very short ping timeout")
}

func (s *DatabaseTestSuite) TestNewSQLClient_InvalidDSN() {
	cfg := database.SQLClientConfig{
		DriverName: "pgx",
		DSN:        "postgres://invalid:invalid@nonexistent:5432/nonexistent?sslmode=disable",
	}

	db, err := database.NewSQLClient(cfg)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), db)
}

func (s *DatabaseTestSuite) TestNewSQLClient_AppliesDefaults() {
	cfg := database.SQLClientConfig{
		DriverName: "pgx",
		DSN:        s.testDSN,
		// Leave pool/timeout at zero to test defaults
	}

	db, err := database.NewSQLClient(cfg)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), db)
	defer db.Close()

	assert.Equal(s.T(), database.DefaultMaxOpenConns, db.Stats().MaxOpenConnections)
	// Idle count can be 0 depending on timing; we only check that connection succeeded with default config
	err = db.Ping()
	assert.NoError(s.T(), err)
}

func (s *DatabaseTestSuite) TestNewSQLClient_EmptyDriverName_ReturnsError() {
	cfg := database.SQLClientConfig{
		DriverName: "",
		DSN:        s.testDSN,
	}

	db, err := database.NewSQLClient(cfg)
	assert.ErrorIs(s.T(), err, database.ErrEmptyDriverName)
	assert.Nil(s.T(), db)
}

func (s *DatabaseTestSuite) TestNewSQLClient_EmptyDSN_ReturnsError() {
	cfg := database.SQLClientConfig{
		DriverName: "pgx",
		DSN:        "",
	}

	db, err := database.NewSQLClient(cfg)
	assert.ErrorIs(s.T(), err, database.ErrEmptyDSN)
	assert.Nil(s.T(), db)
}
