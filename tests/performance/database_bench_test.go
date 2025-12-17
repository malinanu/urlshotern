package performance

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// BenchmarkResult holds the results of a benchmark
type BenchmarkResult struct {
	Name        string
	Iterations  int
	TotalTime   time.Duration
	AvgTime     time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	P95Time     time.Duration
	P99Time     time.Duration
	Success     int
	Failures    int
}

// Database connection
var db *sql.DB

func setupDB() error {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default connection for local testing (SSL disabled)
		// For production/CI testing with SSL, set DATABASE_URL environment variable:
		// export DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=require"
		dbURL = "postgres://postgres:password@localhost:5432/urlshortener?sslmode=disable"
	}

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db.Ping()
}

func teardownDB() {
	if db != nil {
		db.Close()
	}
}

func TestMain(m *testing.M) {
	if err := setupDB(); err != nil {
		fmt.Printf("Failed to setup database: %v\n", err)
		os.Exit(1)
	}
	defer teardownDB()

	os.Exit(m.Run())
}

// BenchmarkURLInsert benchmarks URL insertion
func BenchmarkURLInsert(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shortCode := randomString(8)
		originalURL := fmt.Sprintf("https://example.com/test-%d", i)

		_, err := db.ExecContext(ctx, `
			INSERT INTO urls (short_code, original_url, created_at)
			VALUES ($1, $2, NOW())
		`, shortCode, originalURL)

		if err != nil {
			b.Errorf("Insert failed: %v", err)
		}
	}
}

// BenchmarkURLLookup benchmarks URL lookup by short_code (primary operation)
func BenchmarkURLLookup(b *testing.B) {
	ctx := context.Background()

	// Setup: Insert test data
	testCodes := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		shortCode := randomString(8)
		testCodes[i] = shortCode
		db.ExecContext(ctx, `
			INSERT INTO urls (short_code, original_url, created_at)
			VALUES ($1, $2, NOW())
		`, shortCode, fmt.Sprintf("https://example.com/test-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shortCode := testCodes[rand.Intn(len(testCodes))]

		var originalURL string
		err := db.QueryRowContext(ctx, `
			SELECT original_url FROM urls WHERE short_code = $1
		`, shortCode).Scan(&originalURL)

		if err != nil {
			b.Errorf("Lookup failed: %v", err)
		}
	}
}

// BenchmarkURLLookupWithExpiry benchmarks URL lookup with expiry check
func BenchmarkURLLookupWithExpiry(b *testing.B) {
	ctx := context.Background()

	// Setup: Insert test data with expiry
	testCodes := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		shortCode := randomString(8)
		testCodes[i] = shortCode
		db.ExecContext(ctx, `
			INSERT INTO urls (short_code, original_url, created_at, expires_at)
			VALUES ($1, $2, NOW(), NOW() + INTERVAL '1 day')
		`, shortCode, fmt.Sprintf("https://example.com/test-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shortCode := testCodes[rand.Intn(len(testCodes))]

		var originalURL string
		err := db.QueryRowContext(ctx, `
			SELECT original_url FROM urls
			WHERE short_code = $1
			AND (expires_at IS NULL OR expires_at > NOW())
		`, shortCode).Scan(&originalURL)

		if err != nil && err != sql.ErrNoRows {
			b.Errorf("Lookup failed: %v", err)
		}
	}
}

// BenchmarkClickInsert benchmarks click tracking insertion
func BenchmarkClickInsert(b *testing.B) {
	ctx := context.Background()

	// Setup: Insert a test URL
	shortCode := randomString(8)
	db.ExecContext(ctx, `
		INSERT INTO urls (short_code, original_url, created_at)
		VALUES ($1, $2, NOW())
	`, shortCode, "https://example.com/test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.ExecContext(ctx, `
			INSERT INTO clicks (short_code, clicked_at, ip_address, user_agent, referrer)
			VALUES ($1, NOW(), $2, $3, $4)
		`, shortCode, fmt.Sprintf("192.168.1.%d", rand.Intn(255)), "Mozilla/5.0", "https://google.com")

		if err != nil {
			b.Errorf("Click insert failed: %v", err)
		}
	}
}

// BenchmarkAnalyticsQuery benchmarks analytics aggregation query
func BenchmarkAnalyticsQuery(b *testing.B) {
	ctx := context.Background()

	// Setup: Insert test data with clicks
	shortCode := randomString(8)
	db.ExecContext(ctx, `
		INSERT INTO urls (short_code, original_url, created_at)
		VALUES ($1, $2, NOW())
	`, shortCode, "https://example.com/test")

	for i := 0; i < 1000; i++ {
		db.ExecContext(ctx, `
			INSERT INTO clicks (short_code, clicked_at, ip_address, user_agent, referrer)
			VALUES ($1, NOW() - ($2 || ' hours')::INTERVAL, $3, $4, $5)
		`, shortCode, rand.Intn(720), fmt.Sprintf("192.168.1.%d", rand.Intn(255)), "Mozilla/5.0", "https://google.com")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := db.QueryContext(ctx, `
			SELECT
				DATE(clicked_at) as date,
				COUNT(*) as clicks,
				COUNT(DISTINCT ip_address) as unique_visitors
			FROM clicks
			WHERE short_code = $1
			AND clicked_at >= NOW() - INTERVAL '30 days'
			GROUP BY DATE(clicked_at)
			ORDER BY date DESC
		`, shortCode)

		if err != nil {
			b.Errorf("Analytics query failed: %v", err)
		}
		rows.Close()
	}
}

// BenchmarkUserURLsList benchmarks listing user's URLs with pagination
func BenchmarkUserURLsList(b *testing.B) {
	ctx := context.Background()

	// Setup: Insert test user URLs
	userID := int64(1)
	for i := 0; i < 100; i++ {
		shortCode := randomString(8)
		db.ExecContext(ctx, `
			INSERT INTO urls (short_code, original_url, user_id, created_at)
			VALUES ($1, $2, $3, NOW())
		`, shortCode, fmt.Sprintf("https://example.com/test-%d", i), userID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := rand.Intn(10) * 20
		limit := 20

		rows, err := db.QueryContext(ctx, `
			SELECT short_code, original_url, created_at,
				   COALESCE((SELECT COUNT(*) FROM clicks WHERE clicks.short_code = urls.short_code), 0) as click_count
			FROM urls
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`, userID, limit, offset)

		if err != nil {
			b.Errorf("User URLs list query failed: %v", err)
		}
		rows.Close()
	}
}

// BenchmarkConcurrentLookups tests concurrent read performance
func BenchmarkConcurrentLookups(b *testing.B) {
	ctx := context.Background()

	// Setup: Insert test data
	testCodes := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		shortCode := randomString(8)
		testCodes[i] = shortCode
		db.ExecContext(ctx, `
			INSERT INTO urls (short_code, original_url, created_at)
			VALUES ($1, $2, NOW())
		`, shortCode, fmt.Sprintf("https://example.com/test-%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			shortCode := testCodes[rand.Intn(len(testCodes))]

			var originalURL string
			err := db.QueryRowContext(ctx, `
				SELECT original_url FROM urls WHERE short_code = $1
			`, shortCode).Scan(&originalURL)

			if err != nil {
				b.Errorf("Concurrent lookup failed: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentInserts tests concurrent write performance
func BenchmarkConcurrentInserts(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			shortCode := fmt.Sprintf("%s-%d", randomString(8), i)
			originalURL := fmt.Sprintf("https://example.com/test-%d", i)

			_, err := db.ExecContext(ctx, `
				INSERT INTO urls (short_code, original_url, created_at)
				VALUES ($1, $2, NOW())
			`, shortCode, originalURL)

			if err != nil {
				b.Errorf("Concurrent insert failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkIndexedVsSequentialScan compares indexed vs sequential scan
func BenchmarkIndexedVsSequentialScan(b *testing.B) {
	ctx := context.Background()

	// Setup: Insert 10000 URLs
	for i := 0; i < 10000; i++ {
		shortCode := randomString(8)
		db.ExecContext(ctx, `
			INSERT INTO urls (short_code, original_url, created_at)
			VALUES ($1, $2, NOW())
		`, shortCode, fmt.Sprintf("https://example.com/test-%d", i))
	}

	b.Run("Indexed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.QueryContext(ctx, `
				SELECT short_code, original_url
				FROM urls
				WHERE short_code LIKE 'a%'
				LIMIT 100
			`)
			if err != nil {
				b.Errorf("Indexed query failed: %v", err)
			}
			rows.Close()
		}
	})

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.QueryContext(ctx, `
				SELECT short_code, original_url
				FROM urls
				WHERE original_url LIKE '%test-%'
				LIMIT 100
			`)
			if err != nil {
				b.Errorf("Sequential query failed: %v", err)
			}
			rows.Close()
		}
	})
}

// CustomBenchmark runs a custom benchmark with detailed metrics
func CustomBenchmark(b *testing.B, name string, fn func() error) BenchmarkResult {
	var durations []time.Duration
	successes := 0
	failures := 0

	start := time.Now()

	for i := 0; i < b.N; i++ {
		opStart := time.Now()
		err := fn()
		duration := time.Since(opStart)

		durations = append(durations, duration)

		if err != nil {
			failures++
		} else {
			successes++
		}
	}

	totalTime := time.Since(start)

	// Calculate statistics
	result := BenchmarkResult{
		Name:       name,
		Iterations: b.N,
		TotalTime:  totalTime,
		Success:    successes,
		Failures:   failures,
	}

	if len(durations) > 0 {
		// Sort for percentile calculation
		sorted := make([]time.Duration, len(durations))
		copy(sorted, durations)

		// Simple insertion sort for small arrays
		for i := 1; i < len(sorted); i++ {
			key := sorted[i]
			j := i - 1
			for j >= 0 && sorted[j] > key {
				sorted[j+1] = sorted[j]
				j--
			}
			sorted[j+1] = key
		}

		result.MinTime = sorted[0]
		result.MaxTime = sorted[len(sorted)-1]
		result.P95Time = sorted[int(float64(len(sorted))*0.95)]
		result.P99Time = sorted[int(float64(len(sorted))*0.99)]

		var sum time.Duration
		for _, d := range durations {
			sum += d
		}
		result.AvgTime = sum / time.Duration(len(durations))
	}

	return result
}

// Helper function to generate random strings
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
