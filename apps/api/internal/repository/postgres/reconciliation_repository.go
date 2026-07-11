package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReconciliationRepository struct {
	db *pgxpool.Pool
}

func NewReconciliationRepository(
	db *pgxpool.Pool,
) *ReconciliationRepository {
	return &ReconciliationRepository{
		db: db,
	}
}

func (repository *ReconciliationRepository) MarkPendingDerivation(
	ctx context.Context,
	task reconciliation.PendingDerivation,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	normalized := task.Normalize()
	if err := normalized.Validate(); err != nil {
		return err
	}

	const query = `
		INSERT INTO derived_reconciliation_tasks (
			deduplication_key,
			ingestion_run_id,
			icao24,
			derivation_type,
			status,
			observed_from,
			observed_to,
			last_error
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			'pending',
			$5,
			$6,
			$7
		)
		ON CONFLICT (deduplication_key)
		DO UPDATE SET
			status = 'pending',
			observed_from = EXCLUDED.observed_from,
			observed_to = EXCLUDED.observed_to,
			last_error = EXCLUDED.last_error,
			updated_at = now(),
			completed_at = NULL;
	`

	_, err := repository.db.Exec(
		ctx,
		query,
		normalized.DeduplicationKey(),
		nullableUUID(normalized.IngestionRunID),
		normalized.ICAO24,
		string(normalized.DerivationType),
		nullableReconciliationTime(normalized.ObservedFrom),
		nullableReconciliationTime(normalized.ObservedTo),
		normalized.LastError,
	)
	if err != nil {
		return fmt.Errorf(
			"insert pending derived reconciliation task: %w",
			err,
		)
	}

	return nil
}

func nullableReconciliationTime(
	value time.Time,
) *time.Time {
	if value.IsZero() {
		return nil
	}

	normalized := value.UTC()

	return &normalized
}
