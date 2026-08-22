// Applies migrations/*.sql to DATABASE_URL, tracking what has already run in a
// schema_migrations table so it is safe to re-run.
//
//	DATABASE_URL='postgres://...' go run ./cmd/migrate
//
// Set ADMIN_EMAIL and ADMIN_PASS to also seed an admin user; there is no signup
// route, so this is the only way to create the first one.
package main

import (
	"log"
	"os"
	"path/filepath"
	"sort"

	"portfolio/internal/auth"
	"portfolio/internal/database"
)

func main() {
	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("DATABASE_URL is required")
	}

	database.Connect()
	defer database.DB.Close()

	if _, err := database.DB.Exec(
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		)`,
	); err != nil {
		log.Fatalf("Failed to create schema_migrations: %v", err)
	}

	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		log.Fatalf("Failed to read migrations: %v", err)
	}
	if len(files) == 0 {
		log.Fatal("No migrations found — run this from the repository root")
	}
	sort.Strings(files)

	for _, file := range files {
		name := filepath.Base(file)

		var applied bool
		if err := database.DB.QueryRow(
			"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE filename = $1)", name,
		).Scan(&applied); err != nil {
			log.Fatalf("Failed to check %s: %v", name, err)
		}
		if applied {
			log.Printf("skip %s (already applied)", name)
			continue
		}

		sql, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read %s: %v", name, err)
		}

		tx, err := database.DB.Begin()
		if err != nil {
			log.Fatalf("Failed to begin transaction: %v", err)
		}
		if _, err := tx.Exec(string(sql)); err != nil {
			tx.Rollback()
			log.Fatalf("Failed to apply %s: %v", name, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations (filename) VALUES ($1)", name); err != nil {
			tx.Rollback()
			log.Fatalf("Failed to record %s: %v", name, err)
		}
		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit %s: %v", name, err)
		}

		log.Printf("applied %s", name)
	}

	seedAdmin()
}

func seedAdmin() {
	email, password := os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASS")
	if email == "" || password == "" {
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	if _, err := database.DB.Exec(
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, 'admin')
		 ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash`,
		email, hash,
	); err != nil {
		log.Fatalf("Failed to seed admin: %v", err)
	}

	log.Printf("seeded admin %s", email)
}
