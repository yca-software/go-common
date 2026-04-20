package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	error_helpers "github.com/yca-software/go-common/error"
	observer "github.com/yca-software/go-common/observer"
	repository "github.com/yca-software/go-common/repository"
)

// Product represents a test model for integration testing
type Product struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Price       float64   `db:"price"`
	Stock       int       `db:"stock"`
	Category    *string   `db:"category"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// MockMetricsHook implements QueryMetricsHook for testing
type MockMetricsHook struct {
	Records []MetricRecord
}

type MetricRecord struct {
	Operation string
	Table     string
	QueryType string
	Status    string
	Duration  float64
}

func (m *MockMetricsHook) RecordQuery(operation, table, queryType, status string, duration float64) {
	m.Records = append(m.Records, MetricRecord{
		Operation: operation,
		Table:     table,
		QueryType: queryType,
		Status:    status,
		Duration:  duration,
	})
}

func (m *MockMetricsHook) Reset() {
	m.Records = []MetricRecord{}
}

// RepositoryTestSuite is the test suite for repository integration tests
type RepositoryTestSuite struct {
	suite.Suite
	db              *sqlx.DB
	container       testcontainers.Container
	repo            repository.Repository[Product]
	repoWithMetrics repository.Repository[Product]
	metricsHook     *MockMetricsHook
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func TestNewRepository_NilDB_Panics(t *testing.T) {
	columns := []string{"id", "name"}
	assert.Panics(t, func() {
		repository.NewRepository[Product](nil, "products", columns, nil)
	})
}

func TestNewRepository_EmptyTableName_Panics(t *testing.T) {
	db := &sqlx.DB{}
	columns := []string{"id", "name"}
	assert.Panics(t, func() {
		repository.NewRepository[Product](db, "", columns, nil)
	})
}

func TestNewRepository_EmptyColumns_Panics(t *testing.T) {
	db := &sqlx.DB{}
	assert.Panics(t, func() {
		repository.NewRepository[Product](db, "products", nil, nil)
	})
	assert.Panics(t, func() {
		repository.NewRepository[Product](db, "products", []string{}, nil)
	})
}

func TestWrapSQLError_NilReturnsNil(t *testing.T) {
	assert.Nil(t, repository.WrapSQLError(nil))
}

func TestWrapSQLError_ErrNoRows_ReturnsNotFound(t *testing.T) {
	err := repository.WrapSQLError(sql.ErrNoRows)
	require.Error(t, err)
	var apiErr *error_helpers.Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func (suite *RepositoryTestSuite) SetupSuite() {
	ctx := context.Background()

	// Start PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(suite.T(), err)

	suite.container = postgresContainer

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(suite.T(), err)

	// Parse connection string and create sqlx.DB
	config, err := pgx.ParseConfig(connStr)
	require.NoError(suite.T(), err)

	db := stdlib.OpenDB(*config)
	suite.db = sqlx.NewDb(db, "pgx")

	// Run migrations
	suite.runMigrations()

	// Create repositories
	columns := []string{"id", "name", "description", "price", "stock", "category", "is_active", "created_at", "updated_at"}
	suite.repo = repository.NewRepository[Product](suite.db, "products", columns, nil)

	// Create repository with metrics
	suite.metricsHook = &MockMetricsHook{Records: []MetricRecord{}}
	var metricsHook observer.QueryMetricsHook = suite.metricsHook
	suite.repoWithMetrics = repository.NewRepository[Product](suite.db, "products", columns, metricsHook)
}

func (suite *RepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.container != nil {
		ctx := context.Background()
		suite.container.Terminate(ctx)
	}
}

func (suite *RepositoryTestSuite) SetupTest() {
	// Clean up before each test
	_, err := suite.db.Exec("TRUNCATE TABLE products RESTART IDENTITY CASCADE")
	require.NoError(suite.T(), err)
	suite.metricsHook.Reset()
}

func (suite *RepositoryTestSuite) runMigrations() {
	migrationSQL := `
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL,
			stock INTEGER NOT NULL DEFAULT 0,
			category VARCHAR(100),
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name)
		);

		CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
		CREATE INDEX IF NOT EXISTS idx_products_is_active ON products(is_active);
		CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);
	`
	_, err := suite.db.Exec(migrationSQL)
	require.NoError(suite.T(), err)
}

// TestCreate tests the Create method
func (suite *RepositoryTestSuite) TestCreate() {
	data := map[string]any{
		"name":        "Test Product",
		"description": "A test product",
		"price":       99.99,
		"stock":       10,
		"category":    "electronics",
		"is_active":   true,
	}

	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Verify the product was created
	product, err := suite.repo.BaseGet(nil, squirrel.Eq{"name": "Test Product"}, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Product", product.Name)
	assert.Equal(suite.T(), 99.99, product.Price)
	assert.Equal(suite.T(), 10, product.Stock)
}

// TestCreateWithMetrics tests Create with metrics hook
func (suite *RepositoryTestSuite) TestCreateWithMetrics() {
	data := map[string]any{
		"name":      "Test Product",
		"price":     99.99,
		"stock":     10,
		"is_active": true,
	}

	err := suite.repoWithMetrics.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Verify metrics were recorded
	assert.Greater(suite.T(), len(suite.metricsHook.Records), 0)
	found := false
	for _, record := range suite.metricsHook.Records {
		if record.Operation == "exec" && record.Table == "products" {
			found = true
			assert.Equal(suite.T(), "success", record.Status)
			assert.Greater(suite.T(), record.Duration, 0.0)
		}
	}
	assert.True(suite.T(), found, "Expected to find exec metrics for products table")
}

// TestGet tests the Get method
func (suite *RepositoryTestSuite) TestGet() {
	// Create a product first
	data := map[string]any{
		"name":      "Get Test Product",
		"price":     49.99,
		"stock":     5,
		"is_active": true,
	}
	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Get the product
	product, err := suite.repo.BaseGet(nil, squirrel.Eq{"name": "Get Test Product"}, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Get Test Product", product.Name)
	assert.Equal(suite.T(), 49.99, product.Price)
}

// TestBaseGet_NilCondition_ReturnsError tests that BaseGet with nil condition returns ErrConditionRequired
func (suite *RepositoryTestSuite) TestBaseGet_NilCondition_ReturnsError() {
	_, err := suite.repo.BaseGet(nil, nil, nil)
	assert.ErrorIs(suite.T(), err, repository.ErrConditionRequired)
}

// TestGetNotFound tests Get when record doesn't exist
func (suite *RepositoryTestSuite) TestGetNotFound() {
	product, err := suite.repo.BaseGet(nil, squirrel.Eq{"name": "Non-existent Product"}, nil)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), product)
	// Should be a NotFoundError (StatusCode 404)
	var notFoundErr *error_helpers.Error
	require.ErrorAs(suite.T(), err, &notFoundErr)
	assert.Equal(suite.T(), 404, notFoundErr.StatusCode)
}

// TestGetWithColumns tests Get with specific columns
func (suite *RepositoryTestSuite) TestGetWithColumns() {
	// Create a product first
	data := map[string]any{
		"name":      "Column Test Product",
		"price":     29.99,
		"stock":     3,
		"is_active": true,
	}
	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Get only specific columns
	columns := []string{"id", "name", "price"}
	product, err := suite.repo.BaseGet(nil, squirrel.Eq{"name": "Column Test Product"}, &columns)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Column Test Product", product.Name)
	assert.Equal(suite.T(), 29.99, product.Price)
	// Stock should be zero (default) since it wasn't selected
	assert.Equal(suite.T(), 0, product.Stock)
}

// TestSelect tests the Select method
func (suite *RepositoryTestSuite) TestSelect() {
	// Create multiple products
	products := []map[string]any{
		{"name": "Product 1", "price": 10.0, "stock": 1, "is_active": true, "category": "cat1"},
		{"name": "Product 2", "price": 20.0, "stock": 2, "is_active": true, "category": "cat1"},
		{"name": "Product 3", "price": 30.0, "stock": 3, "is_active": false, "category": "cat2"},
	}

	for _, p := range products {
		err := suite.repo.BaseCreate(nil, p)
		require.NoError(suite.T(), err)
	}

	// Select all active products
	results, err := suite.repo.BaseSelect(nil, squirrel.Eq{"is_active": true}, nil, "")
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), *results, 2)

	// Select with sorting
	results, err = suite.repo.BaseSelect(nil, squirrel.Eq{"category": "cat1"}, nil, "price ASC")
	require.NoError(suite.T(), err)
	require.Len(suite.T(), *results, 2)
	assert.Equal(suite.T(), "Product 1", (*results)[0].Name)
	assert.Equal(suite.T(), "Product 2", (*results)[1].Name)
}

func (suite *RepositoryTestSuite) TestSelect_InvalidSortRejected() {
	data := map[string]any{
		"name":      "Sort Validation Product",
		"price":     19.99,
		"stock":     2,
		"is_active": true,
	}
	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	_, err = suite.repo.BaseSelect(nil, nil, nil, "price; DROP TABLE products; --")
	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid sort")
}

func (suite *RepositoryTestSuite) TestPaginatedSelect_MultiSortAllowed() {
	products := []map[string]any{
		{"name": "Sort A", "price": 10.0, "stock": 2, "is_active": true},
		{"name": "Sort B", "price": 10.0, "stock": 1, "is_active": true},
	}
	for _, p := range products {
		err := suite.repo.BaseCreate(nil, p)
		require.NoError(suite.T(), err)
	}

	results, err := suite.repo.BasePaginatedSelect(nil, nil, nil, "price DESC, stock ASC", 10, 0)
	require.NoError(suite.T(), err)
	require.GreaterOrEqual(suite.T(), len(*results), 2)
	assert.Equal(suite.T(), "Sort B", (*results)[0].Name)
	assert.Equal(suite.T(), "Sort A", (*results)[1].Name)
}

// TestPaginatedSelect tests the PaginatedSelect method
func (suite *RepositoryTestSuite) TestPaginatedSelect() {
	// Create 10 products
	for i := 1; i <= 10; i++ {
		data := map[string]any{
			"name":      fmt.Sprintf("Product %d", i),
			"price":     float64(i * 10),
			"stock":     i,
			"is_active": true,
		}
		err := suite.repo.BaseCreate(nil, data)
		require.NoError(suite.T(), err)
	}

	// Get first page (limit 3, offset 0)
	results, err := suite.repo.BasePaginatedSelect(nil, nil, nil, "price ASC", 3, 0)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), *results, 3)
	assert.Equal(suite.T(), "Product 1", (*results)[0].Name)

	// Get second page
	results, err = suite.repo.BasePaginatedSelect(nil, nil, nil, "price ASC", 3, 3)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), *results, 3)
	assert.Equal(suite.T(), "Product 4", (*results)[0].Name)
}

// TestCount tests the Count method
func (suite *RepositoryTestSuite) TestCount() {
	// Create products
	products := []map[string]any{
		{"name": "Count Product 1", "price": 10.0, "stock": 1, "is_active": true},
		{"name": "Count Product 2", "price": 20.0, "stock": 2, "is_active": true},
		{"name": "Count Product 3", "price": 30.0, "stock": 3, "is_active": false},
	}

	for _, p := range products {
		err := suite.repo.BaseCreate(nil, p)
		require.NoError(suite.T(), err)
	}

	// Count all
	count, err := suite.repo.BaseCount(nil, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, count)

	// Count with condition
	count, err = suite.repo.BaseCount(nil, squirrel.Eq{"is_active": true})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count)
}

// TestUpdate tests the Update method
func (suite *RepositoryTestSuite) TestUpdate() {
	// Create a product
	data := map[string]any{
		"name":      "Update Test Product",
		"price":     50.0,
		"stock":     5,
		"is_active": true,
	}
	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Update the product
	updateData := map[string]any{
		"price": 75.0,
		"stock": 10,
	}
	err = suite.repo.BaseUpdate(nil, squirrel.Eq{"name": "Update Test Product"}, updateData)
	require.NoError(suite.T(), err)

	// Verify the update
	product, err := suite.repo.BaseGet(nil, squirrel.Eq{"name": "Update Test Product"}, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 75.0, product.Price)
	assert.Equal(suite.T(), 10, product.Stock)
}

// TestBaseDelete_NilCondition_ReturnsError tests that BaseDelete with nil condition returns ErrConditionRequired
func (suite *RepositoryTestSuite) TestBaseDelete_NilCondition_ReturnsError() {
	err := suite.repo.BaseDelete(nil, nil)
	assert.ErrorIs(suite.T(), err, repository.ErrConditionRequired)
}

// TestBaseUpdate_NilCondition_ReturnsError tests that BaseUpdate with nil condition returns ErrConditionRequired
func (suite *RepositoryTestSuite) TestBaseUpdate_NilCondition_ReturnsError() {
	err := suite.repo.BaseUpdate(nil, nil, map[string]any{"price": 10.0})
	assert.ErrorIs(suite.T(), err, repository.ErrConditionRequired)
}

// TestBaseUpdate_EmptyData_ReturnsError tests that BaseUpdate with empty data returns error
func (suite *RepositoryTestSuite) TestBaseUpdate_EmptyData_ReturnsError() {
	data := map[string]any{"name": "Update Empty Data Product", "price": 10.0, "stock": 1, "is_active": true}
	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)
	err = suite.repo.BaseUpdate(nil, squirrel.Eq{"name": "Update Empty Data Product"}, map[string]any{})
	assert.Error(suite.T(), err)
	err = suite.repo.BaseUpdate(nil, squirrel.Eq{"name": "Update Empty Data Product"}, nil)
	assert.Error(suite.T(), err)
}

// TestBaseCreate_EmptyData_ReturnsError tests that BaseCreate with empty or nil data returns error
func (suite *RepositoryTestSuite) TestBaseCreate_EmptyData_ReturnsError() {
	err := suite.repo.BaseCreate(nil, map[string]any{})
	assert.Error(suite.T(), err)
	err = suite.repo.BaseCreate(nil, nil)
	assert.Error(suite.T(), err)
}

// TestDelete tests the Delete method
func (suite *RepositoryTestSuite) TestDelete() {
	// Create a product
	data := map[string]any{
		"name":      "Delete Test Product",
		"price":     25.0,
		"stock":     3,
		"is_active": true,
	}
	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Delete the product
	err = suite.repo.BaseDelete(nil, squirrel.Eq{"name": "Delete Test Product"})
	require.NoError(suite.T(), err)

	// Verify it's deleted
	product, err := suite.repo.BaseGet(nil, squirrel.Eq{"name": "Delete Test Product"}, nil)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), product)
}

// TestCreateMany tests the CreateMany method
func (suite *RepositoryTestSuite) TestCreateMany() {
	columns := []string{"name", "price", "stock", "is_active"}
	data := []map[string]any{
		{"name": "Bulk Product 1", "price": 10.0, "stock": 1, "is_active": true},
		{"name": "Bulk Product 2", "price": 20.0, "stock": 2, "is_active": true},
		{"name": "Bulk Product 3", "price": 30.0, "stock": 3, "is_active": true},
	}

	err := suite.repo.BaseCreateMany(nil, columns, data, false)
	require.NoError(suite.T(), err)

	// Verify all were created
	count, err := suite.repo.BaseCount(nil, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, count)
}

// TestCreateManyWithConflict tests CreateMany with ignoreConflict
func (suite *RepositoryTestSuite) TestCreateManyWithConflict() {
	// Create a product first
	data := map[string]any{
		"name":      "Conflict Product",
		"price":     50.0,
		"stock":     5,
		"is_active": true,
	}
	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Try to create the same product again with ignoreConflict
	columns := []string{"name", "price", "stock", "is_active"}
	bulkData := []map[string]any{
		{"name": "Conflict Product", "price": 60.0, "stock": 6, "is_active": true},
		{"name": "New Product", "price": 70.0, "stock": 7, "is_active": true},
	}

	err = suite.repo.BaseCreateMany(nil, columns, bulkData, true)
	require.NoError(suite.T(), err)

	// Verify only one Conflict Product exists and New Product was created
	products, err := suite.repo.BaseSelect(nil, nil, nil, "")
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), *products, 2)
}

// TestTransaction tests transaction support
func (suite *RepositoryTestSuite) TestTransaction() {
	// Begin transaction
	tx, err := suite.repo.BeginTx()
	require.NoError(suite.T(), err)
	defer tx.Rollback()

	// Create product in transaction
	data := map[string]any{
		"name":      "Tx Product",
		"price":     100.0,
		"stock":     10,
		"is_active": true,
	}
	err = suite.repo.BaseCreate(tx, data)
	require.NoError(suite.T(), err)

	// Verify it exists within transaction
	product, err := suite.repo.BaseGet(tx, squirrel.Eq{"name": "Tx Product"}, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Tx Product", product.Name)

	// Commit transaction
	err = tx.Commit()
	require.NoError(suite.T(), err)

	// Verify it exists after commit
	product, err = suite.repo.BaseGet(nil, squirrel.Eq{"name": "Tx Product"}, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Tx Product", product.Name)
}

// TestTransactionRollback tests transaction rollback
func (suite *RepositoryTestSuite) TestTransactionRollback() {
	// Begin transaction
	tx, err := suite.repo.BeginTx()
	require.NoError(suite.T(), err)

	// Create product in transaction
	data := map[string]any{
		"name":      "Rollback Product",
		"price":     200.0,
		"stock":     20,
		"is_active": true,
	}
	err = suite.repo.BaseCreate(tx, data)
	require.NoError(suite.T(), err)

	// Rollback transaction
	err = tx.Rollback()
	require.NoError(suite.T(), err)

	// Verify it doesn't exist after rollback
	product, err := suite.repo.BaseGet(nil, squirrel.Eq{"name": "Rollback Product"}, nil)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), product)
}

// TestTransactionUpdate tests update within transaction
func (suite *RepositoryTestSuite) TestTransactionUpdate() {
	// Create product first
	data := map[string]any{
		"name":      "Tx Update Product",
		"price":     50.0,
		"stock":     5,
		"is_active": true,
	}
	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Begin transaction
	tx, err := suite.repo.BeginTx()
	require.NoError(suite.T(), err)
	defer tx.Rollback()

	// Update in transaction
	updateData := map[string]any{"price": 150.0}
	err = suite.repo.BaseUpdate(tx, squirrel.Eq{"name": "Tx Update Product"}, updateData)
	require.NoError(suite.T(), err)

	// Verify update within transaction
	product, err := suite.repo.BaseGet(tx, squirrel.Eq{"name": "Tx Update Product"}, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 150.0, product.Price)

	// Commit
	err = tx.Commit()
	require.NoError(suite.T(), err)

	// Verify update persisted
	product, err = suite.repo.BaseGet(nil, squirrel.Eq{"name": "Tx Update Product"}, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 150.0, product.Price)
}

// TestTransactionCount tests count within transaction
func (suite *RepositoryTestSuite) TestTransactionCount() {
	// Create products
	for i := 1; i <= 5; i++ {
		data := map[string]any{
			"name":      fmt.Sprintf("Tx Count Product %d", i),
			"price":     float64(i * 10),
			"stock":     i,
			"is_active": true,
		}
		err := suite.repo.BaseCreate(nil, data)
		require.NoError(suite.T(), err)
	}

	// Begin transaction
	tx, err := suite.repo.BeginTx()
	require.NoError(suite.T(), err)
	defer tx.Commit()

	// Count in transaction
	count, err := suite.repo.BaseCount(tx, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 5, count)
}

// TestTransactionSelect tests select within transaction
func (suite *RepositoryTestSuite) TestTransactionSelect() {
	// Create products
	for i := 1; i <= 3; i++ {
		data := map[string]any{
			"name":      fmt.Sprintf("Tx Select Product %d", i),
			"price":     float64(i * 10),
			"stock":     i,
			"is_active": true,
		}
		err := suite.repo.BaseCreate(nil, data)
		require.NoError(suite.T(), err)
	}

	// Begin transaction
	tx, err := suite.repo.BeginTx()
	require.NoError(suite.T(), err)
	defer tx.Commit()

	// Select in transaction
	results, err := suite.repo.BaseSelect(tx, nil, nil, "price ASC")
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), *results, 3)
}

// TestTransactionPaginatedSelect tests paginated select within transaction
func (suite *RepositoryTestSuite) TestTransactionPaginatedSelect() {
	// Create products
	for i := 1; i <= 5; i++ {
		data := map[string]any{
			"name":      fmt.Sprintf("Tx Paginated Product %d", i),
			"price":     float64(i * 10),
			"stock":     i,
			"is_active": true,
		}
		err := suite.repo.BaseCreate(nil, data)
		require.NoError(suite.T(), err)
	}

	// Begin transaction
	tx, err := suite.repo.BeginTx()
	require.NoError(suite.T(), err)
	defer tx.Commit()

	// Paginated select in transaction
	results, err := suite.repo.BasePaginatedSelect(tx, nil, nil, "price ASC", 2, 0)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), *results, 2)
}

// TestTransactionCreateMany tests create many within transaction
func (suite *RepositoryTestSuite) TestTransactionCreateMany() {
	// Begin transaction
	tx, err := suite.repo.BeginTx()
	require.NoError(suite.T(), err)
	defer tx.Commit()

	columns := []string{"name", "price", "stock", "is_active"}
	data := []map[string]any{
		{"name": "Tx Bulk 1", "price": 10.0, "stock": 1, "is_active": true},
		{"name": "Tx Bulk 2", "price": 20.0, "stock": 2, "is_active": true},
	}

	err = suite.repo.BaseCreateMany(tx, columns, data, false)
	require.NoError(suite.T(), err)

	// Verify within transaction
	count, err := suite.repo.BaseCount(tx, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count)
}

// TestTransactionDelete tests delete within transaction
func (suite *RepositoryTestSuite) TestTransactionDelete() {
	// Create product
	data := map[string]any{
		"name":      "Tx Delete Product",
		"price":     50.0,
		"stock":     5,
		"is_active": true,
	}
	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Begin transaction
	tx, err := suite.repo.BeginTx()
	require.NoError(suite.T(), err)
	defer tx.Commit()

	// Delete in transaction
	err = suite.repo.BaseDelete(tx, squirrel.Eq{"name": "Tx Delete Product"})
	require.NoError(suite.T(), err)

	// Verify deleted within transaction
	product, err := suite.repo.BaseGet(tx, squirrel.Eq{"name": "Tx Delete Product"}, nil)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), product)
}

// TestGetQueryBuilder tests the GetQueryBuilder method
func (suite *RepositoryTestSuite) TestGetQueryBuilder() {
	builder := suite.repo.GetQueryBuilder()
	assert.NotNil(suite.T(), builder)

	// Test that it uses dollar placeholders (PostgreSQL)
	sql, args, err := builder.Select("id").From("products").Where(squirrel.Eq{"id": 1}).ToSql()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), sql, "$1")
	assert.Len(suite.T(), args, 1)
}

// TestGetDB tests the GetDB method
func (suite *RepositoryTestSuite) TestGetDB() {
	db := suite.repo.GetDB()
	assert.NotNil(suite.T(), db)
	assert.Equal(suite.T(), suite.db, db)
}

// TestUniqueConstraintError tests error handling for unique constraint violations
func (suite *RepositoryTestSuite) TestUniqueConstraintError() {
	// Create a product
	data := map[string]any{
		"name":      "Unique Product",
		"price":     50.0,
		"stock":     5,
		"is_active": true,
	}
	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Try to create another with the same name (unique constraint)
	err = suite.repo.BaseCreate(nil, data)
	assert.Error(suite.T(), err)
	// Should be a conflict error (StatusCode 409)
	var conflictErr *error_helpers.Error
	require.ErrorAs(suite.T(), err, &conflictErr)
	assert.Equal(suite.T(), 409, conflictErr.StatusCode)
}

// TestMetricsHookAllOperations tests that all operations record metrics
func (suite *RepositoryTestSuite) TestMetricsHookAllOperations() {
	// Create
	data := map[string]any{
		"name":      "Metrics Test Product",
		"price":     50.0,
		"stock":     5,
		"is_active": true,
	}
	err := suite.repoWithMetrics.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Get
	_, err = suite.repoWithMetrics.BaseGet(nil, squirrel.Eq{"name": "Metrics Test Product"}, nil)
	require.NoError(suite.T(), err)

	// Select
	_, err = suite.repoWithMetrics.BaseSelect(nil, nil, nil, "")
	require.NoError(suite.T(), err)

	// Count
	_, err = suite.repoWithMetrics.BaseCount(nil, nil)
	require.NoError(suite.T(), err)

	// Update
	updateData := map[string]any{"price": 75.0}
	err = suite.repoWithMetrics.BaseUpdate(nil, squirrel.Eq{"name": "Metrics Test Product"}, updateData)
	require.NoError(suite.T(), err)

	// Delete
	err = suite.repoWithMetrics.BaseDelete(nil, squirrel.Eq{"name": "Metrics Test Product"})
	require.NoError(suite.T(), err)

	// Verify metrics were recorded for all operations
	operations := map[string]bool{
		"exec":   false,
		"get":    false,
		"select": false,
		"count":  false,
	}

	for _, record := range suite.metricsHook.Records {
		if record.Table == "products" && record.Status == "success" {
			operations[record.Operation] = true
		}
	}

	for op, found := range operations {
		assert.True(suite.T(), found, "Expected to find %s operation in metrics", op)
	}
}

// TestComplexQuery tests complex queries with multiple conditions
func (suite *RepositoryTestSuite) TestComplexQuery() {
	// Create products with various attributes
	products := []map[string]any{
		{"name": "Complex 1", "price": 10.0, "stock": 5, "is_active": true, "category": "A"},
		{"name": "Complex 2", "price": 20.0, "stock": 10, "is_active": true, "category": "A"},
		{"name": "Complex 3", "price": 30.0, "stock": 15, "is_active": false, "category": "B"},
		{"name": "Complex 4", "price": 40.0, "stock": 20, "is_active": true, "category": "B"},
	}

	for _, p := range products {
		err := suite.repo.BaseCreate(nil, p)
		require.NoError(suite.T(), err)
	}

	// Complex query: active products in category A with price > 15
	condition := squirrel.And{
		squirrel.Eq{"is_active": true},
		squirrel.Eq{"category": "A"},
		squirrel.Gt{"price": 15.0},
	}

	results, err := suite.repo.BaseSelect(nil, condition, nil, "price ASC")
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), *results, 1)
	assert.Equal(suite.T(), "Complex 2", (*results)[0].Name)
}

// TestEmptyResultSet tests handling of empty result sets
func (suite *RepositoryTestSuite) TestEmptyResultSet() {
	// Select with condition that matches nothing
	results, err := suite.repo.BaseSelect(nil, squirrel.Eq{"name": "Non-existent"}, nil, "")
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), results)
	assert.Len(suite.T(), *results, 0)

	// Paginated select with no results
	results, err = suite.repo.BasePaginatedSelect(nil, squirrel.Eq{"name": "Non-existent"}, nil, "", 10, 0)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), results)
	assert.Len(suite.T(), *results, 0)
}

// TestNullValues tests handling of NULL values
func (suite *RepositoryTestSuite) TestNullValues() {
	// Create product with NULL description and category
	data := map[string]any{
		"name":        "Null Test Product",
		"price":       50.0,
		"stock":       5,
		"is_active":   true,
		"description": nil,
		"category":    nil,
	}
	err := suite.repo.BaseCreate(nil, data)
	require.NoError(suite.T(), err)

	// Retrieve and verify NULLs are handled
	product, err := suite.repo.BaseGet(nil, squirrel.Eq{"name": "Null Test Product"}, nil)
	require.NoError(suite.T(), err)
	assert.Nil(suite.T(), product.Description)
	assert.Nil(suite.T(), product.Category)
}
