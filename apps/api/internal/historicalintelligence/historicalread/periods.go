package historicalread

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (repository *PostgresRepository) ReadPeriods(
	ctx context.Context,
	queries PeriodQueries,
) (PeriodSnapshots, error) {
	if ctx == nil {
		return PeriodSnapshots{}, ErrContextRequired
	}
	if err := ctx.Err(); err != nil {
		return PeriodSnapshots{}, err
	}

	normalized, err := normalizePeriodQueries(queries)
	if err != nil {
		return PeriodSnapshots{}, err
	}
	if repository == nil ||
		(repository.beginner == nil &&
			repository.client == nil) {
		return PeriodSnapshots{},
			ErrPostgresClientRequired
	}

	if repository.beginner == nil {
		return repository.readPeriodSnapshots(
			ctx,
			repository.client,
			normalized,
			repository.isolationLevel,
		)
	}

	transaction, err :=
		repository.beginner.BeginSnapshot(ctx)
	if err != nil {
		return PeriodSnapshots{},
			databaseError(
				"begin repeatable-read period snapshot",
				err,
			)
	}

	snapshots, readErr :=
		repository.readPeriodSnapshots(
			ctx,
			transaction,
			normalized,
			SnapshotIsolationRepeatableRead,
		)
	if readErr != nil {
		rollbackErr := transaction.Rollback(
			context.Background(),
		)
		if rollbackErr != nil &&
			!errors.Is(
				rollbackErr,
				pgx.ErrTxClosed,
			) {
			return PeriodSnapshots{},
				errors.Join(
					readErr,
					databaseError(
						"rollback repeatable-read period snapshot",
						rollbackErr,
					),
				)
		}
		return PeriodSnapshots{}, readErr
	}

	if err := transaction.Commit(ctx); err != nil {
		return PeriodSnapshots{},
			databaseError(
				"commit repeatable-read period snapshot",
				err,
			)
	}

	return snapshots.Clone(), nil
}

func (repository *PostgresRepository) readPeriodSnapshots(
	ctx context.Context,
	client postgresClient,
	queries PeriodQueries,
	isolationLevel string,
) (PeriodSnapshots, error) {
	previous, err := repository.readSnapshot(
		ctx,
		client,
		queries.Previous,
		isolationLevel,
	)
	if err != nil {
		return PeriodSnapshots{}, err
	}
	current, err := repository.readSnapshot(
		ctx,
		client,
		queries.Current,
		isolationLevel,
	)
	if err != nil {
		return PeriodSnapshots{}, err
	}
	if err := ctx.Err(); err != nil {
		return PeriodSnapshots{}, err
	}

	return PeriodSnapshots{
		Previous: previous.Clone(),
		Current:  current.Clone(),
	}.Clone(), nil
}

func normalizePeriodQueries(
	queries PeriodQueries,
) (PeriodQueries, error) {
	previous, err := normalizeQuery(
		queries.Previous,
	)
	if err != nil {
		return PeriodQueries{}, err
	}
	current, err := normalizeQuery(
		queries.Current,
	)
	if err != nil {
		return PeriodQueries{}, err
	}

	if !previous.Window.AsOfTime.Equal(
		current.Window.AsOfTime,
	) {
		return PeriodQueries{},
			ErrPeriodAsOfTimeMismatch
	}
	if !previous.Window.EndTime.Equal(
		current.Window.StartTime,
	) {
		return PeriodQueries{},
			ErrPeriodWindowsNotAdjacent
	}

	return PeriodQueries{
		Previous: previous,
		Current:  current,
	}, nil
}
