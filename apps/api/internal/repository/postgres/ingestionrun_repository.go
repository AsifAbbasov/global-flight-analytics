package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrIngestionRunNotFound = errors.New(
		"ingestion run not found",
	)
	ErrIngestionRunTransitionRejected = errors.New(
		"ingestion run transition rejected",
	)
	ErrIngestionRunRepositoryPoolRequired = errors.New(
		"ingestion run repository pool is required",
	)
	ErrIngestionRunRecoveryTimeInvalid = errors.New(
		"ingestion run recovery times are invalid",
	)
	ErrIngestionRunRecoveryMessageRequired = errors.New(
		"ingestion run recovery error message is required",
	)
)

type IngestionRunRepository struct {
	db *pgxpool.Pool
}

func NewIngestionRunRepository(
	db *pgxpool.Pool,
) *IngestionRunRepository {
	return &IngestionRunRepository{
		db: db,
	}
}

func (r *IngestionRunRepository) CreateRunning(
	ctx context.Context,
	sourceName string,
	regionID string,
	startedAt time.Time,
) (ingestionrun.Run, error) {
	if r == nil || r.db == nil {
		return ingestionrun.Run{},
			ErrIngestionRunRepositoryPoolRequired
	}

	if err := requireRepositoryContext(ctx); err != nil {
		return ingestionrun.Run{}, err
	}

	const query = `
		INSERT INTO ingestion_runs (
			source_name,
			region_id,
			started_at,
			status
		)
		VALUES (
			$1,
			$2,
			$3,
			$4
		)
		RETURNING
			id::text,
			created_at;
	`

	item := ingestionrun.Run{
		SourceName: sourceName,
		RegionID:   regionID,
		StartedAt:  startedAt,
		Status:     ingestionrun.StatusRunning,
	}

	err := r.db.QueryRow(
		ctx,
		query,
		sourceName,
		nullableUUID(regionID),
		startedAt,
		string(ingestionrun.StatusRunning),
	).Scan(
		&item.ID,
		&item.CreatedAt,
	)
	if err != nil {
		return ingestionrun.Run{}, fmt.Errorf(
			"create running ingestion run: %w",
			err,
		)
	}

	return item, nil
}

func (r *IngestionRunRepository) RecoverStaleRunning(
	ctx context.Context,
	staleBefore time.Time,
	recoveredAt time.Time,
	errorMessage string,
) (int64, error) {
	if r == nil || r.db == nil {
		return 0, ErrIngestionRunRepositoryPoolRequired
	}
	if err := requireRepositoryContext(ctx); err != nil {
		return 0, err
	}

	staleBefore = staleBefore.UTC()
	recoveredAt = recoveredAt.UTC()
	if staleBefore.IsZero() ||
		recoveredAt.IsZero() ||
		recoveredAt.Before(staleBefore) {
		return 0, ErrIngestionRunRecoveryTimeInvalid
	}

	normalizedErrorMessage := strings.TrimSpace(errorMessage)
	if normalizedErrorMessage == "" {
		return 0, ErrIngestionRunRecoveryMessageRequired
	}

	const query = `
		UPDATE ingestion_runs
		SET
			finished_at = $2,
			status = $3,
			error_message = $4
		WHERE status = $5
			AND finished_at IS NULL
			AND started_at < $1;
	`

	commandTag, err := r.db.Exec(
		ctx,
		query,
		staleBefore,
		recoveredAt,
		string(ingestionrun.StatusFailed),
		normalizedErrorMessage,
		string(ingestionrun.StatusRunning),
	)
	if err != nil {
		return 0, fmt.Errorf(
			"recover stale ingestion runs: %w",
			err,
		)
	}

	return commandTag.RowsAffected(), nil
}

func (r *IngestionRunRepository) MarkSuccess(
	ctx context.Context,
	id string,
	finishedAt time.Time,
	recordsReceived int,
	recordsInserted int,
	recordsUpdated int,
) error {
	return r.markFinished(
		ctx,
		id,
		finishedAt,
		ingestionrun.StatusSuccess,
		recordsReceived,
		recordsInserted,
		recordsUpdated,
		"",
	)
}

func (r *IngestionRunRepository) MarkFailed(
	ctx context.Context,
	id string,
	finishedAt time.Time,
	recordsReceived int,
	recordsInserted int,
	recordsUpdated int,
	errorMessage string,
) error {
	return r.markFinished(
		ctx,
		id,
		finishedAt,
		ingestionrun.StatusFailed,
		recordsReceived,
		recordsInserted,
		recordsUpdated,
		errorMessage,
	)
}

func (r *IngestionRunRepository) markFinished(
	ctx context.Context,
	id string,
	finishedAt time.Time,
	status ingestionrun.Status,
	recordsReceived int,
	recordsInserted int,
	recordsUpdated int,
	errorMessage string,
) error {
	if r == nil || r.db == nil {
		return ErrIngestionRunRepositoryPoolRequired
	}

	if err := requireRepositoryContext(ctx); err != nil {
		return err
	}

	normalizedErrorMessage, validationErr := validateIngestionRunCompletion(
		status,
		recordsReceived,
		recordsInserted,
		recordsUpdated,
		errorMessage,
	)
	if validationErr != nil {
		return validationErr
	}

	const query = `
		WITH updated AS (
			UPDATE ingestion_runs
			SET
				finished_at = $2,
				status = $3,
				records_received = $4,
				records_inserted = $5,
				records_updated = $6,
				error_message = $7
			WHERE id = $1
				AND status = $8
			RETURNING id
		)
		SELECT CASE
			WHEN EXISTS (SELECT 1 FROM updated)
				THEN 'updated'
			WHEN EXISTS (
				SELECT 1
				FROM ingestion_runs
				WHERE id = $1
			)
				THEN 'transition_rejected'
			ELSE 'not_found'
		END;
	`

	var outcome string

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		finishedAt,
		string(status),
		recordsReceived,
		recordsInserted,
		recordsUpdated,
		nullableText(normalizedErrorMessage),
		string(ingestionrun.StatusRunning),
	).Scan(
		&outcome,
	)
	if err != nil {
		return fmt.Errorf(
			"mark ingestion run %s: %w",
			status,
			err,
		)
	}

	switch outcome {
	case "updated":
		return nil

	case "transition_rejected":
		return ErrIngestionRunTransitionRejected

	case "not_found":
		return ErrIngestionRunNotFound

	default:
		return fmt.Errorf(
			"mark ingestion run %s returned unknown outcome %q",
			status,
			outcome,
		)
	}
}
