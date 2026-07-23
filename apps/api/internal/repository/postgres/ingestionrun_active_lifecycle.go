package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
)

var ErrIngestionRunSourceNameRequired = errors.New(
	"ingestion run source name is required",
)

func (r *IngestionRunRepository) UpdateRunningSource(
	ctx context.Context,
	id string,
	sourceName string,
) error {
	if r == nil || r.db == nil {
		return ErrIngestionRunRepositoryPoolRequired
	}
	if err := requireRepositoryContext(ctx); err != nil {
		return err
	}

	normalizedSourceName := strings.TrimSpace(sourceName)
	if normalizedSourceName == "" {
		return ErrIngestionRunSourceNameRequired
	}

	const query = `
		WITH updated AS (
			UPDATE ingestion_runs
			SET source_name = $2
			WHERE id = $1
				AND status = $3
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
	if err := r.db.QueryRow(
		ctx,
		query,
		id,
		normalizedSourceName,
		string(ingestionrun.StatusRunning),
	).Scan(&outcome); err != nil {
		return fmt.Errorf(
			"update running ingestion run source: %w",
			err,
		)
	}

	return ingestionRunActiveMutationOutcome(
		"update source",
		outcome,
	)
}

func (r *IngestionRunRepository) DeleteRunning(
	ctx context.Context,
	id string,
) error {
	if r == nil || r.db == nil {
		return ErrIngestionRunRepositoryPoolRequired
	}
	if err := requireRepositoryContext(ctx); err != nil {
		return err
	}

	const query = `
		WITH deleted AS (
			DELETE FROM ingestion_runs
			WHERE id = $1
				AND status = $2
				AND NOT EXISTS (
					SELECT 1
					FROM flight_states
					WHERE ingestion_run_id = $1
				)
			RETURNING id
		)
		SELECT CASE
			WHEN EXISTS (SELECT 1 FROM deleted)
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
	if err := r.db.QueryRow(
		ctx,
		query,
		id,
		string(ingestionrun.StatusRunning),
	).Scan(&outcome); err != nil {
		return fmt.Errorf(
			"delete running ingestion run: %w",
			err,
		)
	}

	return ingestionRunActiveMutationOutcome(
		"delete",
		outcome,
	)
}

func (r *IngestionRunRepository) MarkPartial(
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
		ingestionrun.StatusPartial,
		recordsReceived,
		recordsInserted,
		recordsUpdated,
		errorMessage,
	)
}

func ingestionRunActiveMutationOutcome(
	operation string,
	outcome string,
) error {
	switch outcome {
	case "updated":
		return nil
	case "transition_rejected":
		return ErrIngestionRunTransitionRejected
	case "not_found":
		return ErrIngestionRunNotFound
	default:
		return fmt.Errorf(
			"ingestion run %s returned unknown outcome %q",
			operation,
			outcome,
		)
	}
}
