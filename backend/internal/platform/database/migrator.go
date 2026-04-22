package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const schemaMigrationsTable = "schema_migrations"

type Migrator struct {
	dir string
}

func NewMigrator(dir string) Migrator {
	return Migrator{dir: dir}
}

func DiscoverMigrations(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}

	sort.Strings(files)

	return files, nil
}

func (m Migrator) Apply(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("sql db is required")
	}
	if m.dir == "" {
		return errors.New("migration directory is required")
	}

	files, err := DiscoverMigrations(m.dir)
	if err != nil {
		return err
	}

	if err := ensureSchemaMigrationsTable(ctx, db); err != nil {
		return err
	}

	applied, err := loadAppliedMigrations(ctx, db)
	if err != nil {
		return err
	}

	for _, file := range files {
		name := filepath.Base(file)
		if _, ok := applied[name]; ok {
			continue
		}

		if err := applyFile(ctx, db, file, name); err != nil {
			return err
		}
	}

	return nil
}

func ensureSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		create table if not exists schema_migrations (
			name varchar(255) primary key,
			applied_at timestamptz not null default now()
		)
	`)
	if err != nil {
		return fmt.Errorf("ensure schema migrations table: %w", err)
	}

	return nil
}

func loadAppliedMigrations(ctx context.Context, db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, "select name from "+schemaMigrationsTable)
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[name] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return applied, nil
}

func applyFile(ctx context.Context, db *sql.DB, filePath string, name string) error {
	payload, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}

	if strings.TrimSpace(string(payload)) != "" {
		if _, err := tx.ExecContext(ctx, string(payload)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", name, err)
		}
	}

	if _, err := tx.ExecContext(ctx, "insert into "+schemaMigrationsTable+" (name) values ($1)", name); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration %s: %w", name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}

	return nil
}
