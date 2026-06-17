package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
    dsn := "host=localhost port=5432 user=postgres password=malin123456 dbname=urlshortener_db sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
        fmt.Println("Ping failed, trying with user urlshortener_user")
        dsn = "host=localhost port=5432 user=urlshortener_user password=malin123456 dbname=urlshortener_db sslmode=disable"
        db2, err2 := sql.Open("postgres", dsn)
        if err2 != nil {
            log.Fatalf("Failed second try: %v", err2)
        }
        if err2 := db2.Ping(); err2 != nil {
            log.Fatalf("Ping failed second try: %v", err2)
        }
        db = db2
	}

	fmt.Println("Connected successfully!")
    
    // Check if we can read the static_pages
    rows, err := db.Query("SELECT slug FROM static_pages WHERE slug IN ('privacy', 'terms')")
    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }
    defer rows.Close()
    
    count := 0
    for rows.Next() {
        var slug string
        rows.Scan(&slug)
        fmt.Println("Found slug:", slug)
        count++
    }
    
    if count == 0 {
        fmt.Println("Did not find privacy or terms pages.")
    }
}
