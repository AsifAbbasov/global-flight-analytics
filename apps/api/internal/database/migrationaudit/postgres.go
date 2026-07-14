package migrationaudit

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStateLoader struct {
	pool *pgxpool.Pool
}

func NewPostgresStateLoader(
	pool *pgxpool.Pool,
) (*PostgresStateLoader, error) {
	if pool == nil {
		return nil, ErrPostgresPoolRequired
	}

	return &PostgresStateLoader{
		pool: pool,
	}, nil
}

func (loader *PostgresStateLoader) Load(
	ctx context.Context,
) (DatabaseState, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return DatabaseState{}, err
	}

	var tableName *string
	if err := loader.pool.QueryRow(
		ctx,
		`SELECT to_regclass('schema_migrations')::text;`,
	).Scan(&tableName); err != nil {
		return DatabaseState{},
			&DatabaseInspectionError{
				Operation: "detect schema_migrations table",
				Err:       err,
			}
	}

	if tableName == nil ||
		strings.TrimSpace(*tableName) == "" {
		return DatabaseState{
			SchemaMigrationsTableExists: false,
			AppliedMigrations:           []AppliedMigration{},
		}, nil
	}

	rows, err := loader.pool.Query(
		ctx,
		`
			SELECT
				version,
				name,
				checksum,
				applied_at
			FROM schema_migrations
			ORDER BY version, name, applied_at;
		`,
	)
	if err != nil {
		return DatabaseState{},
			&DatabaseInspectionError{
				Operation: "query schema_migrations rows",
				Err:       err,
			}
	}
	defer rows.Close()

	applied := make(
		[]AppliedMigration,
		0,
	)
	for rows.Next() {
		var migration AppliedMigration
		if err := rows.Scan(
			&migration.Version,
			&migration.Name,
			&migration.Checksum,
			&migration.AppliedAt,
		); err != nil {
			return DatabaseState{},
				&DatabaseInspectionError{
					Operation: "scan schema_migrations row",
					Err:       err,
				}
		}

		migration.Version = strings.TrimSpace(
			migration.Version,
		)
		migration.Name = strings.TrimSpace(
			migration.Name,
		)
		migration.Checksum = strings.TrimSpace(
			migration.Checksum,
		)
		migration.AppliedAt =
			migration.AppliedAt.UTC()

		applied = append(
			applied,
			migration,
		)
	}
	if err := rows.Err(); err != nil {
		return DatabaseState{},
			&DatabaseInspectionError{
				Operation: "iterate schema_migrations rows",
				Err:       err,
			}
	}

	return DatabaseState{
		SchemaMigrationsTableExists: true,
		AppliedMigrations:           applied,
	}, nil
}
