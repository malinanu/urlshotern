// cmd/migrate/main.go — database migration runner
// Usage:
//
//	go run ./cmd/migrate up        # apply all pending migrations
//	go run ./cmd/migrate down      # roll back the last applied migration
//	go run ./cmd/migrate status    # print applied / pending migrations
//	go run ./cmd/migrate up N      # apply exactly N pending migrations
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const migrationsDir = "configs/migrations"

var versionRe = regexp.MustCompile(`^(\d+)_.*_(up|down)\.sql$`)

func main() {
	godotenv.Load()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: migrate <up|down|status> [N]")
		os.Exit(1)
	}

	command := os.Args[1]

	db := connectDB()
	defer db.Close()

	ensureMigrationsTable(db)

	switch command {
	case "up":
		n := -1 // apply all
		if len(os.Args) >= 3 {
			var err error
			n, err = strconv.Atoi(os.Args[2])
			if err != nil || n < 1 {
				log.Fatalf("invalid count: %s", os.Args[2])
			}
		}
		migrateUp(db, n)
	case "down":
		migrateDown(db)
	case "status":
		printStatus(db)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		os.Exit(1)
	}
}

// connectDB opens a Postgres connection from environment variables.
func connectDB() *sql.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "urlshortener"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "urlshortener_db"),
		getEnv("DB_SSLMODE", "disable"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	return db
}

func ensureMigrationsTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version  BIGINT PRIMARY KEY,
		name     TEXT NOT NULL,
		applied_at TIMESTAMPTZ DEFAULT NOW()
	)`)
	if err != nil {
		log.Fatalf("create schema_migrations: %v", err)
	}
}

type migration struct {
	version  int64
	name     string
	upFile   string
	downFile string
}

func loadMigrations() []migration {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("read migrations dir: %v", err)
	}

	upFiles := map[int64]*migration{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := versionRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		ver, _ := strconv.ParseInt(m[1], 10, 64)
		direction := m[2]
		path := filepath.Join(migrationsDir, e.Name())
		if _, ok := upFiles[ver]; !ok {
			upFiles[ver] = &migration{version: ver}
		}
		if direction == "up" {
			upFiles[ver].upFile = path
			upFiles[ver].name = e.Name()
		} else {
			upFiles[ver].downFile = path
		}
	}

	migrations := make([]migration, 0, len(upFiles))
	for _, m := range upFiles {
		migrations = append(migrations, *m)
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	return migrations
}

func appliedVersions(db *sql.DB) map[int64]bool {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		log.Fatalf("query applied: %v", err)
	}
	defer rows.Close()
	applied := map[int64]bool{}
	for rows.Next() {
		var v int64
		rows.Scan(&v)
		applied[v] = true
	}
	return applied
}

func runSQL(db *sql.DB, file string) {
	content, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("read %s: %v", file, err)
	}
	if _, err := db.Exec(string(content)); err != nil {
		log.Fatalf("exec %s: %v", file, err)
	}
}

func migrateUp(db *sql.DB, limit int) {
	applied := appliedVersions(db)
	migrations := loadMigrations()

	count := 0
	for _, m := range migrations {
		if limit >= 0 && count >= limit {
			break
		}
		if applied[m.version] {
			continue
		}
		if m.upFile == "" {
			log.Printf("skip v%d: no up file", m.version)
			continue
		}
		log.Printf("applying %s ...", m.upFile)
		runSQL(db, m.upFile)
		_, err := db.Exec("INSERT INTO schema_migrations (version, name) VALUES ($1, $2)", m.version, m.name)
		if err != nil {
			log.Fatalf("record migration v%d: %v", m.version, err)
		}
		log.Printf("  ✓ v%d applied", m.version)
		count++
	}
	if count == 0 {
		log.Println("No pending migrations.")
	} else {
		log.Printf("%d migration(s) applied.", count)
	}
}

func migrateDown(db *sql.DB) {
	applied := appliedVersions(db)
	migrations := loadMigrations()

	// Find the highest applied version
	var last *migration
	for i := len(migrations) - 1; i >= 0; i-- {
		if applied[migrations[i].version] {
			last = &migrations[i]
			break
		}
	}
	if last == nil {
		log.Println("Nothing to roll back.")
		return
	}
	if last.downFile == "" {
		log.Fatalf("no down file for v%d (%s)", last.version, last.name)
	}
	log.Printf("rolling back %s ...", last.downFile)
	runSQL(db, last.downFile)
	_, err := db.Exec("DELETE FROM schema_migrations WHERE version = $1", last.version)
	if err != nil {
		log.Fatalf("remove migration record v%d: %v", last.version, err)
	}
	log.Printf("  ✓ v%d rolled back", last.version)
}

func printStatus(db *sql.DB) {
	applied := appliedVersions(db)
	migrations := loadMigrations()

	fmt.Printf("%-8s  %-6s  %s\n", "VERSION", "STATUS", "FILE")
	fmt.Println("--------  ------  ----")
	for _, m := range migrations {
		status := "pending"
		if applied[m.version] {
			status = "applied"
		}
		fmt.Printf("%-8d  %-6s  %s\n", m.version, status, m.name)
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
