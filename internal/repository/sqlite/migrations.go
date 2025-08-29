package sqlite

import (
	"context"
	"embed"
	"fmt"
	"path"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// runMigrations applies all pending migrations to the database
func (r *Repository) runMigrations(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := r.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all available migrations
	migrations, err := r.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations
	appliedVersions, err := r.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if _, applied := appliedVersions[migration.Version]; !applied {
			if err := r.applyMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (r *Repository) createMigrationsTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// loadMigrations loads all migration files from the embedded filesystem
func (r *Repository) loadMigrations() ([]Migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Parse version from filename (e.g., "001_create_urls_table.sql" -> 1)
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) != 2 {
			continue
		}

		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			continue
		}

		// Read migration content
		content, err := migrationsFS.ReadFile(path.Join("migrations", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    strings.TrimSuffix(parts[1], ".sql"),
			SQL:     string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedMigrations returns a map of applied migration versions
func (r *Repository) getAppliedMigrations(ctx context.Context) (map[int]bool, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// applyMigration applies a single migration
func (r *Repository) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	if _, err := tx.ExecContext(ctx, 
		"INSERT INTO schema_migrations (version) VALUES (?)", 
		migration.Version); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}