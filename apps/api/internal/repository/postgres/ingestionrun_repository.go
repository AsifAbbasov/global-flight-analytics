package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrIngestionRunNotFound = errors.New(
		"ingestion run not found",
	)
	ErrIngestionRunRepositoryPoolRequired = errors.New(
		"ingestion run repository pool is required",
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

	if ctx == nil {
		ctx = context.Background()
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

	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		UPDATE ingestion_runs
		SET
			finished_at = $2,
			status = $3,
			records_received = $4,
			records_inserted = $5,
			records_updated = $6,
			error_message = $7
		WHERE id = $1;
	`

	commandTag, err := r.db.Exec(
		ctx,
		query,
		id,
		finishedAt,
		string(status),
		recordsReceived,
		recordsInserted,
		recordsUpdated,
		nullableText(errorMessage),
	)
	if err != nil {
		return fmt.Errorf(
			"mark ingestion run %s: %w",
			status,
			err,
		)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrIngestionRunNotFound
	}

	return nil
}
