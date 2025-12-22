package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	_ "github.com/lib/pq"
)

// PostgresOptimizedStorage provides optimized PostgreSQL connection pooling
type PostgresOptimizedStorage struct {
	db     *sql.DB
	dbRead *sql.DB // Separate connection pool for read replicas
}

// ConnectionPoolConfig holds database connection pool configuration
type ConnectionPoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	ConnTimeout     time.Duration
}

// NewPostgresOptimizedStorage creates an optimized PostgreSQL storage instance
func NewPostgresOptimizedStorage(config *configs.Config) (*PostgresOptimizedStorage, error) {
	// Optimized connection pool configuration
	poolConfig := &ConnectionPoolConfig{
		MaxOpenConns:    100,              // Increased from 25 (4x improvement)
		MaxIdleConns:    25,               // 25% of max connections
		ConnMaxLifetime: 30 * time.Minute, // Reduced from 1 hour
		ConnMaxIdleTime: 10 * time.Minute, // Close idle connections faster
		ConnTimeout:     30 * time.Second, // Connection timeout
	}

	// Primary database connection (read-write)
	// Use SSL in production, disable in development
	sslMode := "require"
	if config.Environment == "development" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		config.DBHost,
		config.DBPort,
		config.DBUser,
		config.DBPassword,
		config.DBName,
		sslMode,
		int(poolConfig.ConnTimeout.Seconds()),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for primary database
	configureConnectionPool(db, poolConfig)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Read replica configuration (for production with read replicas)
	// In development, use the same database
	dbRead := db

	// In production, uncomment and configure read replica:
	/*
		readReplicaDSN := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
			config.DBReadHost,  // Read replica host
			config.DBPort,
			config.DBUser,
			config.DBPassword,
			config.DBName,
			sslMode,  // Use same SSL mode as primary
			int(poolConfig.ConnTimeout.Seconds()),
		)

		dbRead, err = sql.Open("postgres", readReplicaDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to open read replica database: %w", err)
		}

		// Configure connection pool for read replica (can be more aggressive)
		readPoolConfig := &ConnectionPoolConfig{
			MaxOpenConns:    200,              // More connections for reads
			MaxIdleConns:    50,
			ConnMaxLifetime: 30 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
			ConnTimeout:     30 * time.Second,
		}
		configureConnectionPool(dbRead, readPoolConfig)

		if err := dbRead.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("failed to ping read replica: %w", err)
		}
	*/

	storage := &PostgresOptimizedStorage{
		db:     db,
		dbRead: dbRead,
	}

	// Initialize tables if they don't exist
	if err := storage.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return storage, nil
}

// configureConnectionPool applies optimal connection pool settings
func configureConnectionPool(db *sql.DB, config *ConnectionPoolConfig) {
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
}

// initTables creates the necessary tables with optimized indexes
func (p *PostgresOptimizedStorage) initTables() error {
	queries := []string{
		// URL mappings table with optimized indexes
		`CREATE TABLE IF NOT EXISTS url_mappings (
			id BIGINT PRIMARY KEY,
			short_code VARCHAR(10) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			expires_at TIMESTAMP WITH TIME ZONE,
			click_count BIGINT DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			created_by_ip INET,
			user_id BIGINT
		)`,
		// Composite index for common queries
		`CREATE INDEX IF NOT EXISTS idx_short_code_active ON url_mappings(short_code, is_active) WHERE is_active = TRUE`,
		`CREATE INDEX IF NOT EXISTS idx_user_created ON url_mappings(user_id, created_at DESC) WHERE user_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_expires_at ON url_mappings(expires_at) WHERE expires_at IS NOT NULL`,

		// Click events table with partitioning-ready structure
		`CREATE TABLE IF NOT EXISTS click_events (
			id BIGINT PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			clicked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			ip_address INET,
			user_agent TEXT,
			referrer TEXT,
			country_code CHAR(2)
		)`,
		// Composite index for analytics queries
		`CREATE INDEX IF NOT EXISTS idx_click_events_analytics ON click_events(short_code, clicked_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_click_events_country ON click_events(country_code, clicked_at) WHERE country_code IS NOT NULL`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

// QueryRow executes a query that returns at most one row (uses read replica)
func (p *PostgresOptimizedStorage) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.dbRead.QueryRow(query, args...)
}

// QueryRowContext executes a query with context (uses read replica)
func (p *PostgresOptimizedStorage) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.dbRead.QueryRowContext(ctx, query, args...)
}

// Query executes a query that returns multiple rows (uses read replica)
func (p *PostgresOptimizedStorage) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.dbRead.Query(query, args...)
}

// QueryContext executes a query with context (uses read replica)
func (p *PostgresOptimizedStorage) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.dbRead.QueryContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows (uses primary database)
func (p *PostgresOptimizedStorage) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.db.Exec(query, args...)
}

// ExecContext executes a query with context (uses primary database)
func (p *PostgresOptimizedStorage) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

// Begin starts a database transaction (uses primary database)
func (p *PostgresOptimizedStorage) Begin() (*sql.Tx, error) {
	return p.db.Begin()
}

// BeginTx starts a transaction with context and options
func (p *PostgresOptimizedStorage) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, opts)
}

// Close closes both primary and read replica connections
func (p *PostgresOptimizedStorage) Close() error {
	var err error
	if p.db != nil {
		err = p.db.Close()
	}
	if p.dbRead != nil && p.dbRead != p.db {
		if closeErr := p.dbRead.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

// Stats returns database connection pool statistics for monitoring
func (p *PostgresOptimizedStorage) Stats() sql.DBStats {
	return p.db.Stats()
}

// ReadStats returns read replica connection pool statistics
func (p *PostgresOptimizedStorage) ReadStats() sql.DBStats {
	return p.dbRead.Stats()
}

// IsHealthy checks if both primary and read databases are healthy
func (p *PostgresOptimizedStorage) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Check primary database
	if err := p.db.PingContext(ctx); err != nil {
		return false
	}

	// Check read replica (if different)
	if p.dbRead != p.db {
		if err := p.dbRead.PingContext(ctx); err != nil {
			return false
		}
	}

	return true
}

// PrepareStatement prepares a SQL statement for efficient repeated execution
func (p *PostgresOptimizedStorage) PrepareStatement(query string) (*sql.Stmt, error) {
	return p.db.Prepare(query)
}

// PrepareStatementContext prepares a SQL statement with context
func (p *PostgresOptimizedStorage) PrepareStatementContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return p.db.PrepareContext(ctx, query)
}

// BatchInsert performs batch insert for better performance
func (p *PostgresOptimizedStorage) BatchInsert(ctx context.Context, tableName string, columns []string, values [][]interface{}) error {
	if len(values) == 0 {
		return nil
	}

	// Build batch insert query
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", tableName, joinStrings(columns, ", "))

	var args []interface{}
	argIndex := 1

	for i, row := range values {
		if i > 0 {
			query += ", "
		}
		query += "("
		for j := range row {
			if j > 0 {
				query += ", "
			}
			query += fmt.Sprintf("$%d", argIndex)
			argIndex++
		}
		query += ")"
		args = append(args, row...)
	}

	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

// joinStrings joins string slice with separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// VacuumAnalyze performs database maintenance for optimal performance
func (p *PostgresOptimizedStorage) VacuumAnalyze(tableName string) error {
	query := fmt.Sprintf("VACUUM ANALYZE %s", tableName)
	_, err := p.db.Exec(query)
	return err
}

// GetReplicationLag returns replication lag in seconds (if read replica is configured)
func (p *PostgresOptimizedStorage) GetReplicationLag() (float64, error) {
	if p.dbRead == p.db {
		return 0, nil // No replication configured
	}

	var lag float64
	query := `SELECT EXTRACT(EPOCH FROM (NOW() - pg_last_xact_replay_timestamp()))::FLOAT`
	err := p.dbRead.QueryRow(query).Scan(&lag)
	if err != nil {
		return 0, fmt.Errorf("failed to get replication lag: %w", err)
	}

	return lag, nil
}
