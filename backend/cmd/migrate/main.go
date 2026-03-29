package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "up" {
		log.Fatal("usage: go run ./cmd/migrate up")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://tradelab:tradelab@localhost:5432/tradelab?sslmode=disable"
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	defer db.Close()

	if err := ensureSchemaMigrationsTable(db); err != nil {
		log.Fatalf("ensure schema_migrations: %v", err)
	}

	migrations, err := loadMigrationFiles("migrations")
	if err != nil {
		log.Fatalf("load migrations: %v", err)
	}

	for _, migration := range migrations {
		applied, err := isApplied(db, migration.version)
		if err != nil {
			log.Fatalf("check migration state: %v", err)
		}

		if applied {
			continue
		}

		log.Printf("applying migration %s", migration.version)

		if _, err := db.Exec(migration.contents); err != nil {
			log.Fatalf("apply migration %s: %v", migration.version, err)
		}

		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, migration.version); err != nil {
			log.Fatalf("record migration %s: %v", migration.version, err)
		}
	}

	log.Println("migrations are up to date")
}

type migrationFile struct {
	version  string
	contents string
}

func ensureSchemaMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func isApplied(db *sql.DB, version string) (bool, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists)
	return exists, err
}

func loadMigrationFiles(root string) ([]migrationFile, error) {
	var files []migrationFile

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || !strings.HasSuffix(path, ".up.sql") {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		version := strings.TrimSuffix(filepath.Base(path), ".up.sql")
		files = append(files, migrationFile{
			version:  version,
			contents: string(contents),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].version < files[j].version
	})

	if len(files) == 0 {
		return nil, fmt.Errorf("no migration files found in %s", root)
	}

	return files, nil
}
