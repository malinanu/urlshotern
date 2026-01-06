package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	_ "github.com/lib/pq"
)

// DatabaseManager manages multiple database connections with different permission levels
type DatabaseManager struct {
	AppDB      *sql.DB // Regular application operations (limited permissions)
	AdminDB    *sql.DB // Admin operations (full permissions)
	ReadOnlyDB *sql.DB // Read-only operations (analytics, reporting)
	config     *configs.Config
}

// NewDatabaseManager creates a new database manager with multiple connections
func NewDatabaseManager(config *configs.Config) (*DatabaseManager, error) {
	manager := &DatabaseManager{
		config: config,
	}

	// Initialize app database connection
	appDB, err := initDatabaseConnection(config.AppDB, config.Environment, "app")
	if err != nil {
		return nil, fmt.Errorf("failed to init app DB: %w", err)
	}
	manager.AppDB = appDB

	// Initialize admin database connection
	adminDB, err := initDatabaseConnection(config.AdminDB, config.Environment, "admin")
	if err != nil {
		return nil, fmt.Errorf("failed to init admin DB: %w", err)
	}
	manager.AdminDB = adminDB

	// Initialize read-only database connection
	readOnlyDB, err := initDatabaseConnection(config.ReadOnlyDB, config.Environment, "readonly")
	if err != nil {
		return nil, fmt.Errorf("failed to init readonly DB: %w", err)
	}
	manager.ReadOnlyDB = readOnlyDB

	return manager, nil
}

// initDatabaseConnection initializes a database connection with the given configuration
func initDatabaseConnection(dbConfig configs.DatabaseConfig, environment, name string) (*sql.DB, error) {
	// Use SSL in production, disable in development
	sslMode := "require"
	if environment == "development" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s database: %w", name, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(dbConfig.MaxConns)
	db.SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(dbConfig.ConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(dbConfig.ConnMaxIdleTime) * time.Second)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping %s database: %w", name, err)
	}

	return db, nil
}

// Close closes all database connections
func (dm *DatabaseManager) Close() error {
	var errors []error

	if err := dm.AppDB.Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close app DB: %w", err))
	}

	if err := dm.AdminDB.Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close admin DB: %w", err))
	}

	if err := dm.ReadOnlyDB.Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close readonly DB: %w", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %v", errors)
	}

	return nil
}

// GetPoolStats returns connection pool statistics for all databases
func (dm *DatabaseManager) GetPoolStats() map[string]interface{} {
	return map[string]interface{}{
		"app_db":      getPoolStats(dm.AppDB),
		"admin_db":    getPoolStats(dm.AdminDB),
		"readonly_db": getPoolStats(dm.ReadOnlyDB),
	}
}

// getPoolStats returns statistics for a single database connection pool
func getPoolStats(db *sql.DB) map[string]interface{} {
	stats := db.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// GetAppDB returns the application database connection
func (dm *DatabaseManager) GetAppDB() *sql.DB {
	return dm.AppDB
}

// GetAdminDB returns the admin database connection
func (dm *DatabaseManager) GetAdminDB() *sql.DB {
	return dm.AdminDB
}

// GetReadOnlyDB returns the read-only database connection
func (dm *DatabaseManager) GetReadOnlyDB() *sql.DB {
	return dm.ReadOnlyDB
}

// HealthCheck performs health checks on all database connections
func (dm *DatabaseManager) HealthCheck(ctx context.Context) map[string]interface{} {
	result := make(map[string]interface{})

	// Check app database
	if err := dm.AppDB.PingContext(ctx); err != nil {
		result["app_db"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		result["app_db"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check admin database
	if err := dm.AdminDB.PingContext(ctx); err != nil {
		result["admin_db"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		result["admin_db"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check readonly database
	if err := dm.ReadOnlyDB.PingContext(ctx); err != nil {
		result["readonly_db"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		result["readonly_db"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	return result
}

// Transaction helpers

// BeginAppTx starts a transaction on the app database
func (dm *DatabaseManager) BeginAppTx(ctx context.Context) (*sql.Tx, error) {
	return dm.AppDB.BeginTx(ctx, nil)
}

// BeginAdminTx starts a transaction on the admin database
func (dm *DatabaseManager) BeginAdminTx(ctx context.Context) (*sql.Tx, error) {
	return dm.AdminDB.BeginTx(ctx, nil)
}

// Database access helpers for app database

// AppQueryRow executes a query on the app database that returns at most one row
func (dm *DatabaseManager) AppQueryRow(query string, args ...interface{}) *sql.Row {
	return dm.AppDB.QueryRow(query, args...)
}

// AppQueryRowContext executes a query with context on the app database
func (dm *DatabaseManager) AppQueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return dm.AppDB.QueryRowContext(ctx, query, args...)
}

// AppExec executes a query on the app database that doesn't return rows
func (dm *DatabaseManager) AppExec(query string, args ...interface{}) (sql.Result, error) {
	return dm.AppDB.Exec(query, args...)
}

// AppExecContext executes a query with context on the app database
func (dm *DatabaseManager) AppExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return dm.AppDB.ExecContext(ctx, query, args...)
}

// AppQuery executes a query on the app database that returns multiple rows
func (dm *DatabaseManager) AppQuery(query string, args ...interface{}) (*sql.Rows, error) {
	return dm.AppDB.Query(query, args...)
}

// AppQueryContext executes a query with context on the app database
func (dm *DatabaseManager) AppQueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return dm.AppDB.QueryContext(ctx, query, args...)
}

// Database access helpers for admin database

// AdminQueryRow executes a query on the admin database that returns at most one row
func (dm *DatabaseManager) AdminQueryRow(query string, args ...interface{}) *sql.Row {
	return dm.AdminDB.QueryRow(query, args...)
}

// AdminQueryRowContext executes a query with context on the admin database
func (dm *DatabaseManager) AdminQueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return dm.AdminDB.QueryRowContext(ctx, query, args...)
}

// AdminExec executes a query on the admin database that doesn't return rows
func (dm *DatabaseManager) AdminExec(query string, args ...interface{}) (sql.Result, error) {
	return dm.AdminDB.Exec(query, args...)
}

// AdminExecContext executes a query with context on the admin database
func (dm *DatabaseManager) AdminExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return dm.AdminDB.ExecContext(ctx, query, args...)
}

// AdminQuery executes a query on the admin database that returns multiple rows
func (dm *DatabaseManager) AdminQuery(query string, args ...interface{}) (*sql.Rows, error) {
	return dm.AdminDB.Query(query, args...)
}

// AdminQueryContext executes a query with context on the admin database
func (dm *DatabaseManager) AdminQueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return dm.AdminDB.QueryContext(ctx, query, args...)
}

// Database access helpers for readonly database

// ReadOnlyQueryRow executes a query on the readonly database that returns at most one row
func (dm *DatabaseManager) ReadOnlyQueryRow(query string, args ...interface{}) *sql.Row {
	return dm.ReadOnlyDB.QueryRow(query, args...)
}

// ReadOnlyQueryRowContext executes a query with context on the readonly database
func (dm *DatabaseManager) ReadOnlyQueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return dm.ReadOnlyDB.QueryRowContext(ctx, query, args...)
}

// ReadOnlyQuery executes a query on the readonly database that returns multiple rows
func (dm *DatabaseManager) ReadOnlyQuery(query string, args ...interface{}) (*sql.Rows, error) {
	return dm.ReadOnlyDB.Query(query, args...)
}

// ReadOnlyQueryContext executes a query with context on the readonly database
func (dm *DatabaseManager) ReadOnlyQueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return dm.ReadOnlyDB.QueryContext(ctx, query, args...)
}
