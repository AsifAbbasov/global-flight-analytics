package migrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrMigrationPoolRequired = errors.New("migration database pool is required")
	ErrMigrationDirRequired  = errors.New("migration directory is required")
)

type Migration struct {
	Version  string
	Name     string
	Path     string
	Checksum string
}

type MigrationStatus struct {
	Migration Migration
	Applied   bool
	AppliedAt time.Time
}

type Runner struct {
	pool          *pgxpool.Pool
	migrationsDir string
}

func NewRunner(pool *pgxpool.Pool, migrationsDir string) (*Runner, error) {
	if pool == nil {
		return nil, ErrMigrationPoolRequired
	}

	normalizedDir := strings.TrimSpace(migrationsDir)
	if normalizedDir == "" {
		return nil, ErrMigrationDirRequired
	}

	return &Runner{
		pool:          pool,
		migrationsDir: normalizedDir,
	}, nil
}

func (runner *Runner) EnsureSchemaMigrations(ctx context.Context) error {
	_, err := runner.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version text PRIMARY KEY,
			name text NOT NULL,
			checksum text NOT NULL,
			applied_at timestamptz NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	return nil
}

func (runner *Runner) ListMigrations() ([]Migration, error) {
	entries, err := os.ReadDir(runner.migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	migrations := make([]Migration, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if !strings.HasSuffix(fileName, ".sql") {
			continue
		}

		version, name, err := parseMigrationFileName(fileName)
		if err != nil {
			return nil, err
		}

		path := filepath.Join(runner.migrationsDir, fileName)

		checksum, err := calculateFileChecksum(path)
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, Migration{
			Version:  version,
			Name:     name,
			Path:     path,
			Checksum: checksum,
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		if migrations[i].Version == migrations[j].Version {
			return migrations[i].Name < migrations[j].Name
		}

		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (runner *Runner) Status(ctx context.Context) ([]MigrationStatus, error) {
	if err := runner.EnsureSchemaMigrations(ctx); err != nil {
		return nil, err
	}

	migrations, err := runner.ListMigrations()
	if err != nil {
		return nil, err
	}

	applied, err := runner.appliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	statuses := make([]MigrationStatus, 0, len(migrations))

	for _, migration := range migrations {
		record, ok := applied[migration.Version]

		statuses = append(statuses, MigrationStatus{
			Migration: migration,
			Applied:   ok,
			AppliedAt: record.appliedAt,
		})
	}

	return statuses, nil
}

func (runner *Runner) Baseline(ctx context.Context) ([]Migration, error) {
	if err := runner.EnsureSchemaMigrations(ctx); err != nil {
		return nil, err
	}

	migrations, err := runner.ListMigrations()
	if err != nil {
		return nil, err
	}

	applied, err := runner.appliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	baselined := make([]Migration, 0)

	for _, migration := range migrations {
		record, ok := applied[migration.Version]
		if ok {
			if record.checksum != migration.Checksum {
				return nil, fmt.Errorf("migration %s checksum mismatch", migration.Version)
			}

			continue
		}

		_, err := runner.pool.Exec(ctx, `
			INSERT INTO schema_migrations (version, name, checksum)
			VALUES ($1, $2, $3);
		`, migration.Version, migration.Name, migration.Checksum)
		if err != nil {
			return nil, fmt.Errorf("baseline migration %s: %w", migration.Version, err)
		}

		baselined = append(baselined, migration)
	}

	return baselined, nil
}

func (runner *Runner) ApplyPending(ctx context.Context) ([]Migration, error) {
	if err := runner.EnsureSchemaMigrations(ctx); err != nil {
		return nil, err
	}

	migrations, err := runner.ListMigrations()
	if err != nil {
		return nil, err
	}

	applied, err := runner.appliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	appliedNow := make([]Migration, 0)

	for _, migration := range migrations {
		record, ok := applied[migration.Version]
		if ok {
			if record.checksum != migration.Checksum {
				return nil, fmt.Errorf("migration %s checksum mismatch", migration.Version)
			}

			continue
		}

		sqlBytes, err := os.ReadFile(migration.Path)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", migration.Version, err)
		}

		if _, err := runner.pool.Exec(ctx, string(sqlBytes)); err != nil {
			return nil, fmt.Errorf("apply migration %s: %w", migration.Version, err)
		}

		_, err = runner.pool.Exec(ctx, `
			INSERT INTO schema_migrations (version, name, checksum)
			VALUES ($1, $2, $3);
		`, migration.Version, migration.Name, migration.Checksum)
		if err != nil {
			return nil, fmt.Errorf("record migration %s: %w", migration.Version, err)
		}

		appliedNow = append(appliedNow, migration)
	}

	return appliedNow, nil
}

type appliedMigrationRecord struct {
	checksum  string
	appliedAt time.Time
}

func (runner *Runner) appliedMigrations(ctx context.Context) (map[string]appliedMigrationRecord, error) {
	rows, err := runner.pool.Query(ctx, `
		SELECT version, checksum, applied_at
		FROM schema_migrations;
	`)
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	result := make(map[string]appliedMigrationRecord)

	for rows.Next() {
		var version string
		var record appliedMigrationRecord

		if err := rows.Scan(&version, &record.checksum, &record.appliedAt); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}

		result[version] = record
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return result, nil
}

func parseMigrationFileName(fileName string) (string, string, error) {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return "", "", errors.New("migration file name is empty")
	}

	if !strings.HasSuffix(trimmed, ".sql") {
		return "", "", fmt.Errorf("migration file %s must have .sql extension", fileName)
	}

	nameWithoutExtension := strings.TrimSuffix(trimmed, ".sql")
	parts := strings.SplitN(nameWithoutExtension, "_", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("migration file %s must use format 001_name.sql", fileName)
	}

	version := strings.TrimSpace(parts[0])
	name := strings.TrimSpace(parts[1])

	if version == "" || name == "" {
		return "", "", fmt.Errorf("migration file %s has invalid version or name", fileName)
	}

	return version, name, nil
}

func calculateFileChecksum(path string) (string, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(fileBytes)

	return hex.EncodeToString(hash[:]), nil
}
