package main

import (
	"context"
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

	if err := ensureSchemaMigrationsTable(context.Background(), db); err != nil {
		log.Fatalf("ensure schema_migrations: %v", err)
	}

	migrations, err := loadMigrationFiles("migrations")
	if err != nil {
		log.Fatalf("load migrations: %v", err)
	}

	if err := applyPendingMigrations(context.Background(), db, migrations); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	log.Println("migrations are up to date")
}

type migrationFile struct {
	version  string
	contents string
}

func ensureSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func applyPendingMigrations(ctx context.Context, db *sql.DB, migrations []migrationFile) error {
	for _, migration := range migrations {
		applied, err := isApplied(ctx, db, migration.version)
		if err != nil {
			return fmt.Errorf("check migration state for %s: %w", migration.version, err)
		}

		if applied {
			continue
		}

		log.Printf("applying migration %s", migration.version)

		if err := applyMigration(ctx, db, migration); err != nil {
			return err
		}
	}

	return nil
}

func applyMigration(ctx context.Context, db *sql.DB, migration migrationFile) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", migration.version, err)
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, migration.contents); err != nil {
		return fmt.Errorf("apply migration %s: %w", migration.version, err)
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, migration.version); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.version, err)
	}

	return nil
}

func isApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists)
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
