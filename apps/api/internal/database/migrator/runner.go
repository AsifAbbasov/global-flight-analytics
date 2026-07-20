package migrator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrMigrationPoolRequired = errors.New("migration database pool is required")
	ErrMigrationDirRequired  = errors.New("migration directory is required")
)

const (
	migrationAdvisoryLockNamespace int32 = 0x474641
	migrationAdvisoryLockResource  int32 = 0x4D494752
	migrationLockReleaseTimeout          = 5 * time.Second
)

var nestedTransactionControlPattern = regexp.MustCompile(
	`(?mi)^\s*(BEGIN(?:\s+TRANSACTION)?|COMMIT(?:\s+TRANSACTION)?|ROLLBACK(?:\s+TRANSACTION)?)\s*;`,
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

type migrationExecutor interface {
	Exec(
		context.Context,
		string,
		...any,
	) (pgconn.CommandTag, error)
	Query(
		context.Context,
		string,
		...any,
	) (pgx.Rows, error)
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
	return runner.ensureSchemaMigrations(ctx, runner.pool)
}

func (
	runner *Runner,
) ensureSchemaMigrations(
	ctx context.Context,
	executor migrationExecutor,
) error {
	_, err := executor.Exec(ctx, `
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

	if err := validateUniqueMigrationVersions(migrations); err != nil {
		return nil, err
	}

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

func (runner *Runner) ApplyPending(ctx context.Context) ([]Migration, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	appliedNow := make([]Migration, 0)

	err := runner.withMigrationLock(
		ctx,
		func(conn *pgxpool.Conn) error {
			if err := runner.ensureSchemaMigrations(ctx, conn); err != nil {
				return err
			}

			migrations, err := runner.ListMigrations()
			if err != nil {
				return err
			}

			applied, err := runner.appliedMigrationsWith(ctx, conn)
			if err != nil {
				return err
			}

			for _, migration := range migrations {
				record, ok := applied[migration.Version]
				if ok {
					if record.checksum != migration.Checksum {
						return fmt.Errorf(
							"migration %s checksum mismatch",
							migration.Version,
						)
					}

					continue
				}

				if err := runner.applyMigrationAtomically(
					ctx,
					conn,
					migration,
				); err != nil {
					return err
				}

				appliedNow = append(appliedNow, migration)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return appliedNow, nil
}

func (
	runner *Runner,
) applyMigrationAtomically(
	ctx context.Context,
	conn *pgxpool.Conn,
	migration Migration,
) error {
	sqlBytes, err := os.ReadFile(migration.Path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", migration.Version, err)
	}

	sqlBody, err := prepareMigrationSQL(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("prepare migration %s: %w", migration.Version, err)
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin migration %s transaction: %w", migration.Version, err)
	}

	committed := false
	defer func() {
		if !committed {
			rollbackMigrationTransaction(tx)
		}
	}()

	if _, err := tx.Exec(ctx, sqlBody); err != nil {
		return fmt.Errorf("apply migration %s: %w", migration.Version, err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO schema_migrations (version, name, checksum)
		VALUES ($1, $2, $3);
	`, migration.Version, migration.Name, migration.Checksum); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.Version, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.Version, err)
	}

	committed = true
	return nil
}

func (
	runner *Runner,
) withMigrationLock(
	ctx context.Context,
	operation func(*pgxpool.Conn) error,
) (resultErr error) {
	if ctx == nil {
		ctx = context.Background()
	}

	conn, err := runner.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration lock connection: %w", err)
	}

	if _, err := conn.Exec(ctx, `
		SELECT pg_advisory_lock($1, $2);
	`, migrationAdvisoryLockNamespace, migrationAdvisoryLockResource); err != nil {
		conn.Release()
		return fmt.Errorf("acquire migration advisory lock: %w", err)
	}

	defer func() {
		unlockErr := releaseMigrationLock(conn)
		if unlockErr == nil {
			conn.Release()
		} else {
			destroyLockedConnection(conn)
		}

		if resultErr == nil && unlockErr != nil {
			resultErr = unlockErr
		}
	}()

	return operation(conn)
}

func releaseMigrationLock(conn *pgxpool.Conn) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		migrationLockReleaseTimeout,
	)
	defer cancel()

	var unlocked bool
	if err := conn.QueryRow(ctx, `
		SELECT pg_advisory_unlock($1, $2);
	`, migrationAdvisoryLockNamespace, migrationAdvisoryLockResource).Scan(
		&unlocked,
	); err != nil {
		return fmt.Errorf("release migration advisory lock: %w", err)
	}

	if !unlocked {
		return errors.New("migration advisory lock was not held by the connection")
	}

	return nil
}

func destroyLockedConnection(conn *pgxpool.Conn) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		migrationLockReleaseTimeout,
	)
	defer cancel()

	_ = conn.Hijack().Close(ctx)
}

func rollbackMigrationTransaction(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		migrationLockReleaseTimeout,
	)
	defer cancel()

	_ = tx.Rollback(ctx)
}

func prepareMigrationSQL(rawSQL string) (string, error) {
	trimmed := strings.TrimSpace(
		strings.TrimPrefix(rawSQL, "\ufeff"),
	)
	if trimmed == "" {
		return "", errors.New("migration SQL is empty")
	}

	beginLength := transactionBeginPrefixLength(trimmed)
	commitLength := transactionCommitSuffixLength(trimmed)

	if (beginLength > 0) != (commitLength > 0) {
		return "", errors.New(
			"migration transaction envelope must contain both BEGIN and COMMIT",
		)
	}

	if beginLength == 0 {
		if nestedTransactionControlPattern.MatchString(trimmed) {
			return "", errors.New(
				"migration SQL contains unsupported transaction control statements",
			)
		}

		return trimmed, nil
	}

	body := strings.TrimSpace(
		trimmed[beginLength : len(trimmed)-commitLength],
	)
	if body == "" {
		return "", errors.New("migration transaction body is empty")
	}
	if nestedTransactionControlPattern.MatchString(body) {
		return "", errors.New(
			"migration transaction body contains nested transaction control statements",
		)
	}

	return body, nil
}

func transactionBeginPrefixLength(sql string) int {
	upper := strings.ToUpper(sql)

	for _, prefix := range []string{
		"BEGIN;",
		"BEGIN TRANSACTION;",
	} {
		if strings.HasPrefix(upper, prefix) {
			return len(prefix)
		}
	}

	return 0
}

func transactionCommitSuffixLength(sql string) int {
	upper := strings.ToUpper(sql)

	for _, suffix := range []string{
		"COMMIT;",
		"COMMIT TRANSACTION;",
	} {
		if strings.HasSuffix(upper, suffix) {
			return len(suffix)
		}
	}

	return 0
}

type appliedMigrationRecord struct {
	checksum  string
	appliedAt time.Time
}

func (runner *Runner) appliedMigrations(
	ctx context.Context,
) (map[string]appliedMigrationRecord, error) {
	return runner.appliedMigrationsWith(ctx, runner.pool)
}

func (
	runner *Runner,
) appliedMigrationsWith(
	ctx context.Context,
	executor migrationExecutor,
) (map[string]appliedMigrationRecord, error) {
	rows, err := executor.Query(ctx, `
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
