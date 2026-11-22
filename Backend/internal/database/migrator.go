package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Migration represents a single version-controlled migration step.
type Migration struct {
	Version  int64
	Name     string
	UpFile   string
	DownFile string
}

// MigrationStatus reports the status of a specific migration.
type MigrationStatus struct {
	Version   int64      `json:"version"`
	Name      string     `json:"name"`
	Applied   bool       `json:"applied"`
	AppliedAt *time.Time `json:"applied_at,omitempty"`
}

// Migrator manages and executes database migrations.
type Migrator struct {
	sqlDB         *sql.DB
	migrationsDir string
}

// NewMigrator initializes a new Migrator instance.
func NewMigrator(sqlDB *sql.DB, migrationsDir string) *Migrator {
	return &Migrator{
		sqlDB:         sqlDB,
		migrationsDir: migrationsDir,
	}
}

// EnsureMigrationsTable creates the schema_migrations tracking table if it does not exist.
func (m *Migrator) EnsureMigrationsTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`
	_, err := m.sqlDB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}
	return nil
}

// LoadMigrations parses all .up.sql and .down.sql files from the migrations directory.
func (m *Migrator) LoadMigrations() ([]Migration, error) {
	files, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory %s: %w", m.migrationsDir, err)
	}

	migrationsMap := make(map[int64]*Migration)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}

		if _, exists := migrationsMap[version]; !exists {
			migrationsMap[version] = &Migration{
				Version: version,
			}
		}

		fullPath := filepath.Join(m.migrationsDir, name)
		if strings.HasSuffix(name, ".up.sql") {
			baseName := strings.TrimSuffix(parts[1], ".up.sql")
			migrationsMap[version].Name = baseName
			migrationsMap[version].UpFile = fullPath
		} else if strings.HasSuffix(name, ".down.sql") {
			migrationsMap[version].DownFile = fullPath
		}
	}

	var migrations []Migration
	for _, mig := range migrationsMap {
		migrations = append(migrations, *mig)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// GetAppliedVersions retrieves the list of currently applied migration versions.
func (m *Migrator) GetAppliedVersions(ctx context.Context) (map[int64]time.Time, error) {
	if err := m.EnsureMigrationsTable(ctx); err != nil {
		return nil, err
	}

	rows, err := m.sqlDB.QueryContext(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]time.Time)
	for rows.Next() {
		var v int64
		var t time.Time
		if err := rows.Scan(&v, &t); err != nil {
			return nil, fmt.Errorf("failed to scan applied migration row: %w", err)
		}
		applied[v] = t
	}

	return applied, nil
}

// MigrateUp applies all pending migrations in ascending version order.
func (m *Migrator) MigrateUp(ctx context.Context) (int, error) {
	if err := m.EnsureMigrationsTable(ctx); err != nil {
		return 0, err
	}

	allMigrations, err := m.LoadMigrations()
	if err != nil {
		return 0, err
	}

	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return 0, err
	}

	appliedCount := 0
	for _, mig := range allMigrations {
		if _, alreadyApplied := applied[mig.Version]; alreadyApplied {
			continue
		}

		if mig.UpFile == "" {
			return appliedCount, fmt.Errorf("missing up migration file for version %d (%s)", mig.Version, mig.Name)
		}

		content, err := os.ReadFile(mig.UpFile)
		if err != nil {
			return appliedCount, fmt.Errorf("failed to read migration file %s: %w", mig.UpFile, err)
		}

		tx, err := m.sqlDB.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return appliedCount, fmt.Errorf("failed to begin migration transaction for version %d: %w", mig.Version, err)
		}

		// Execute migration SQL
		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return appliedCount, fmt.Errorf("failed to execute migration %06d_%s: %w", mig.Version, mig.Name, err)
		}

		// Record applied migration
		recordQuery := `INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, NOW())`
		if _, err := tx.ExecContext(ctx, recordQuery, mig.Version, mig.Name); err != nil {
			_ = tx.Rollback()
			return appliedCount, fmt.Errorf("failed to record migration %06d in schema_migrations: %w", mig.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return appliedCount, fmt.Errorf("failed to commit migration %06d: %w", mig.Version, err)
		}

		log.Printf("Applied migration: %06d_%s", mig.Version, mig.Name)
		appliedCount++
	}

	return appliedCount, nil
}

// MigrateDown rolls back up to `steps` migrations in descending version order.
func (m *Migrator) MigrateDown(ctx context.Context, steps int) (int, error) {
	if err := m.EnsureMigrationsTable(ctx); err != nil {
		return 0, err
	}

	if steps <= 0 {
		steps = 1
	}

	allMigrations, err := m.LoadMigrations()
	if err != nil {
		return 0, err
	}

	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return 0, err
	}

	// Filter and sort applied migrations descending
	var appliedMigrations []Migration
	for _, mig := range allMigrations {
		if _, ok := applied[mig.Version]; ok {
			appliedMigrations = append(appliedMigrations, mig)
		}
	}

	sort.Slice(appliedMigrations, func(i, j int) bool {
		return appliedMigrations[i].Version > appliedMigrations[j].Version
	})

	rolledBackCount := 0
	for _, mig := range appliedMigrations {
		if rolledBackCount >= steps {
			break
		}

		if mig.DownFile == "" {
			return rolledBackCount, fmt.Errorf("missing down migration file for version %d (%s)", mig.Version, mig.Name)
		}

		content, err := os.ReadFile(mig.DownFile)
		if err != nil {
			return rolledBackCount, fmt.Errorf("failed to read rollback file %s: %w", mig.DownFile, err)
		}

		tx, err := m.sqlDB.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return rolledBackCount, fmt.Errorf("failed to begin rollback transaction for version %d: %w", mig.Version, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return rolledBackCount, fmt.Errorf("failed to execute rollback %06d_%s: %w", mig.Version, mig.Name, err)
		}

		if _, err := tx.ExecContext(ctx, "DELETE FROM schema_migrations WHERE version = $1", mig.Version); err != nil {
			_ = tx.Rollback()
			return rolledBackCount, fmt.Errorf("failed to remove migration %06d from schema_migrations: %w", mig.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return rolledBackCount, fmt.Errorf("failed to commit rollback %06d: %w", mig.Version, err)
		}

		log.Printf("Rolled back migration: %06d_%s", mig.Version, mig.Name)
		rolledBackCount++
	}

	return rolledBackCount, nil
}

// GetStatus returns the migration status across all discovered migrations.
func (m *Migrator) GetStatus(ctx context.Context) ([]MigrationStatus, error) {
	allMigrations, err := m.LoadMigrations()
	if err != nil {
		return nil, err
	}

	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	var statuses []MigrationStatus
	for _, mig := range allMigrations {
		status := MigrationStatus{
			Version: mig.Version,
			Name:    mig.Name,
		}
		if t, ok := applied[mig.Version]; ok {
			status.Applied = true
			status.AppliedAt = &t
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}
