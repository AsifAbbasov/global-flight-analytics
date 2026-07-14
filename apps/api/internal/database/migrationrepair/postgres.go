package migrationrepair

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresInspector struct {
	pool *pgxpool.Pool
}

func NewPostgresInspector(
	pool *pgxpool.Pool,
) (*PostgresInspector, error) {
	if pool == nil {
		return nil, ErrPostgresPoolRequired
	}

	return &PostgresInspector{
		pool: pool,
	}, nil
}

func (inspector *PostgresInspector) Load(
	ctx context.Context,
) (State, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return State{}, err
	}

	tableExists, err := inspector.regclassExists(
		ctx,
		"schema_migrations",
	)
	if err != nil {
		return State{}, err
	}
	if !tableExists {
		return State{
			SchemaMigrationsTableExists: false,
			AppliedMigrations:           []AppliedMigration{},
		}, nil
	}

	applied, err := inspector.loadAppliedMigrations(
		ctx,
	)
	if err != nil {
		return State{}, err
	}

	state := State{
		SchemaMigrationsTableExists: true,
		AppliedMigrations:           applied,
	}

	if state.FlightTrajectoryReconciliationTaskIDColumnExists, err =
		inspector.columnExists(
			ctx,
			"flight_trajectories",
			"reconciliation_task_id",
		); err != nil {
		return State{}, err
	}
	if state.DataQualityReconciliationTaskIDColumnExists, err =
		inspector.columnExists(
			ctx,
			"data_quality_reports",
			"reconciliation_task_id",
		); err != nil {
		return State{}, err
	}
	if state.FlightTrajectoryReconciliationForeignKeyExists, err =
		inspector.constraintExists(
			ctx,
			"flight_trajectories",
			"flight_trajectories_reconciliation_task_fk",
		); err != nil {
		return State{}, err
	}
	if state.DataQualityReconciliationForeignKeyExists, err =
		inspector.constraintExists(
			ctx,
			"data_quality_reports",
			"data_quality_reports_reconciliation_task_fk",
		); err != nil {
		return State{}, err
	}
	if state.FlightTrajectoryReconciliationUniqueIndexExists, err =
		inspector.regclassExists(
			ctx,
			"flight_trajectories_reconciliation_task_unique",
		); err != nil {
		return State{}, err
	}
	if state.DataQualityReconciliationUniqueIndexExists, err =
		inspector.regclassExists(
			ctx,
			"data_quality_reports_reconciliation_task_unique",
		); err != nil {
		return State{}, err
	}

	if state.IdentityKeyColumnExists, err =
		inspector.columnExists(
			ctx,
			"flight_trajectories",
			"identity_key",
		); err != nil {
		return State{}, err
	}
	if state.IdentityBasisColumnExists, err =
		inspector.columnExists(
			ctx,
			"flight_trajectories",
			"identity_basis",
		); err != nil {
		return State{}, err
	}
	if state.SplitReasonColumnExists, err =
		inspector.columnExists(
			ctx,
			"flight_trajectories",
			"split_reason",
		); err != nil {
		return State{}, err
	}
	if state.IdentityCompletenessCheckExists, err =
		inspector.constraintExists(
			ctx,
			"flight_trajectories",
			"flight_trajectories_identity_completeness_check",
		); err != nil {
		return State{}, err
	}
	if state.IdentityKeyCheckExists, err =
		inspector.constraintExists(
			ctx,
			"flight_trajectories",
			"flight_trajectories_identity_key_check",
		); err != nil {
		return State{}, err
	}
	if state.IdentityBasisCheckExists, err =
		inspector.constraintExists(
			ctx,
			"flight_trajectories",
			"flight_trajectories_identity_basis_check",
		); err != nil {
		return State{}, err
	}
	if state.SplitReasonCheckExists, err =
		inspector.constraintExists(
			ctx,
			"flight_trajectories",
			"flight_trajectories_split_reason_check",
		); err != nil {
		return State{}, err
	}
	if state.IdentityKeyTimeIndexExists, err =
		inspector.regclassExists(
			ctx,
			"flight_trajectories_identity_key_time_idx",
		); err != nil {
		return State{}, err
	}

	return state, nil
}

func (inspector *PostgresInspector) loadAppliedMigrations(
	ctx context.Context,
) ([]AppliedMigration, error) {
	rows, err := inspector.pool.Query(
		ctx,
		`
			SELECT version, name, checksum
			FROM schema_migrations
			WHERE version IN ('010', '011', '012')
			ORDER BY version, name;
		`,
	)
	if err != nil {
		return nil, &InspectionError{
			Operation: "load migration history for versions 010 through 012",
			Err:       err,
		}
	}
	defer rows.Close()

	result := make(
		[]AppliedMigration,
		0,
		3,
	)
	for rows.Next() {
		var migration AppliedMigration
		if err := rows.Scan(
			&migration.Version,
			&migration.Name,
			&migration.Checksum,
		); err != nil {
			return nil, &InspectionError{
				Operation: "scan migration history row",
				Err:       err,
			}
		}

		migration.Version =
			strings.TrimSpace(
				migration.Version,
			)
		migration.Name =
			strings.TrimSpace(
				migration.Name,
			)
		migration.Checksum =
			strings.TrimSpace(
				migration.Checksum,
			)

		result = append(
			result,
			migration,
		)
	}
	if err := rows.Err(); err != nil {
		return nil, &InspectionError{
			Operation: "iterate migration history rows",
			Err:       err,
		}
	}

	return result, nil
}

func (inspector *PostgresInspector) columnExists(
	ctx context.Context,
	tableName string,
	columnName string,
) (bool, error) {
	var exists bool
	if err := inspector.pool.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = current_schema()
				  AND table_name = $1
				  AND column_name = $2
			);
		`,
		tableName,
		columnName,
	).Scan(&exists); err != nil {
		return false, &InspectionError{
			Operation: "inspect column " +
				tableName + "." + columnName,
			Err: err,
		}
	}

	return exists, nil
}

func (inspector *PostgresInspector) constraintExists(
	ctx context.Context,
	tableName string,
	constraintName string,
) (bool, error) {
	var exists bool
	if err := inspector.pool.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM pg_constraint AS constraint_record
				INNER JOIN pg_class AS table_record
					ON table_record.oid =
						constraint_record.conrelid
				INNER JOIN pg_namespace AS namespace_record
					ON namespace_record.oid =
						table_record.relnamespace
				WHERE namespace_record.nspname =
						current_schema()
				  AND table_record.relname = $1
				  AND constraint_record.conname = $2
			);
		`,
		tableName,
		constraintName,
	).Scan(&exists); err != nil {
		return false, &InspectionError{
			Operation: "inspect constraint " +
				constraintName,
			Err: err,
		}
	}

	return exists, nil
}

func (inspector *PostgresInspector) regclassExists(
	ctx context.Context,
	objectName string,
) (bool, error) {
	var exists bool
	if err := inspector.pool.QueryRow(
		ctx,
		`SELECT to_regclass($1) IS NOT NULL;`,
		objectName,
	).Scan(&exists); err != nil {
		return false, &InspectionError{
			Operation: "inspect database object " +
				objectName,
			Err: err,
		}
	}

	return exists, nil
}
