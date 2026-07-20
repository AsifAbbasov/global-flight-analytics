package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDataQualityFlightStateNotPersisted = errors.New(
	"data quality report requires a persisted flight state",
)

type DataQualityRepository struct {
	db *pgxpool.Pool
}

func NewDataQualityRepository(
	db *pgxpool.Pool,
) *DataQualityRepository {
	return &DataQualityRepository{
		db: db,
	}
}

func (repository *DataQualityRepository) SaveFlightStateQuality(
	ctx context.Context,
	state flightstate.FlightState,
	quality dataquality.DataQuality,
) error {
	return repository.saveFlightStateQuality(
		ctx,
		"",
		0,
		state,
		quality,
	)
}

func (repository *DataQualityRepository) SaveReconciledFlightStateQuality(
	ctx context.Context,
	taskID string,
	attemptCount int,
	state flightstate.FlightState,
	quality dataquality.DataQuality,
) error {
	normalizedTaskID := reconciliation.NormalizeTaskID(
		taskID,
	)
	if normalizedTaskID == "" {
		return reconciliation.ErrTaskIDRequired
	}

	if attemptCount <= 0 {
		return reconciliation.ErrAttemptCountInvalid
	}

	return repository.saveFlightStateQuality(
		ctx,
		normalizedTaskID,
		attemptCount,
		state,
		quality,
	)
}

func (repository *DataQualityRepository) saveFlightStateQuality(
	ctx context.Context,
	reconciliationTaskID string,
	attemptCount int,
	state flightstate.FlightState,
	quality dataquality.DataQuality,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	warningsJSON, err := json.Marshal(
		quality.Warnings,
	)
	if err != nil {
		return fmt.Errorf(
			"marshal data quality warnings: %w",
			err,
		)
	}

	if reconciliationTaskID == "" &&
		quality.ValidationStatus == dataquality.ValidationStatusInvalid {
		return repository.insertRejectedFlightStateQuality(
			ctx,
			state,
			quality,
			string(warningsJSON),
		)
	}

	if reconciliationTaskID == "" {
		return repository.insertFlightStateQuality(
			ctx,
			state,
			quality,
			string(warningsJSON),
		)
	}

	return repository.upsertReconciledFlightStateQuality(
		ctx,
		reconciliationTaskID,
		attemptCount,
		state,
		quality,
		string(warningsJSON),
	)
}

func (repository *DataQualityRepository) insertFlightStateQuality(
	ctx context.Context,
	state flightstate.FlightState,
	quality dataquality.DataQuality,
	warningsJSON string,
) error {
	const query = `
		INSERT INTO data_quality_reports (
			state_id,
			flight_state_id,
			validation_status,
			completeness,
			confidence,
			score,
			missing_fields,
			warnings_json
		)
		SELECT
			persisted_state.id,
			persisted_state.id,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7::jsonb
		FROM flight_states AS persisted_state
		WHERE persisted_state.id = $1
		RETURNING id::text;
	`

	var reportID string

	err := repository.db.QueryRow(
		ctx,
		query,
		nullableUUID(state.ID),
		string(quality.ValidationStatus),
		string(quality.Completeness),
		string(quality.Confidence),
		quality.Score,
		quality.MissingFields,
		warningsJSON,
	).Scan(
		&reportID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDataQualityFlightStateNotPersisted
	}
	if err != nil {
		return fmt.Errorf(
			"insert flight state quality report: %w",
			err,
		)
	}

	return nil
}

func (repository *DataQualityRepository) insertRejectedFlightStateQuality(
	ctx context.Context,
	state flightstate.FlightState,
	quality dataquality.DataQuality,
	warningsJSON string,
) error {
	const query = `
		INSERT INTO rejected_flight_state_quality_reports (
			state_id,
			icao24,
			callsign,
			observed_at,
			source_name,
			ingestion_run_id,
			validation_status,
			completeness,
			confidence,
			score,
			missing_fields,
			warnings_json
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12::jsonb
		);
	`

	_, err := repository.db.Exec(
		ctx,
		query,
		nullableUUID(state.ID),
		state.ICAO24,
		state.Callsign,
		state.ObservedAt,
		state.SourceName,
		nullableUUID(state.IngestionRunID),
		string(quality.ValidationStatus),
		string(quality.Completeness),
		string(quality.Confidence),
		quality.Score,
		quality.MissingFields,
		warningsJSON,
	)
	if err != nil {
		return fmt.Errorf(
			"insert rejected flight state quality report: %w",
			err,
		)
	}

	return nil
}

func (repository *DataQualityRepository) upsertReconciledFlightStateQuality(
	ctx context.Context,
	taskID string,
	attemptCount int,
	state flightstate.FlightState,
	quality dataquality.DataQuality,
	warningsJSON string,
) error {
	const query = `
		WITH owned_task AS (
			SELECT id
			FROM derived_reconciliation_tasks
			WHERE id = $8
				AND status = 'processing'
				AND attempt_count = $9
			FOR UPDATE
		),
		persisted_state AS (
			SELECT id
			FROM flight_states
			WHERE id = $1
		),
		upserted_report AS (
			INSERT INTO data_quality_reports (
				state_id,
				flight_state_id,
				validation_status,
				completeness,
				confidence,
				score,
				missing_fields,
				warnings_json,
				reconciliation_task_id
			)
			SELECT
				persisted_state.id,
				persisted_state.id,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7::jsonb,
				owned_task.id
			FROM owned_task
			CROSS JOIN persisted_state
			ON CONFLICT (reconciliation_task_id)
				WHERE reconciliation_task_id IS NOT NULL
			DO UPDATE SET
				state_id = EXCLUDED.state_id,
				flight_state_id = EXCLUDED.flight_state_id,
				validation_status = EXCLUDED.validation_status,
				completeness = EXCLUDED.completeness,
				confidence = EXCLUDED.confidence,
				score = EXCLUDED.score,
				missing_fields = EXCLUDED.missing_fields,
				warnings_json = EXCLUDED.warnings_json,
				calculated_at = now()
			RETURNING id::text AS id
		)
		SELECT
			EXISTS (SELECT 1 FROM owned_task),
			EXISTS (SELECT 1 FROM persisted_state),
			COALESCE((SELECT id FROM upserted_report), '');
	`

	var taskOwned bool
	var statePersisted bool
	var reportID string

	err := repository.db.QueryRow(
		ctx,
		query,
		nullableUUID(state.ID),
		string(quality.ValidationStatus),
		string(quality.Completeness),
		string(quality.Confidence),
		quality.Score,
		quality.MissingFields,
		warningsJSON,
		taskID,
		attemptCount,
	).Scan(
		&taskOwned,
		&statePersisted,
		&reportID,
	)
	if err != nil {
		return fmt.Errorf(
			"upsert reconciled flight state quality report: %w",
			err,
		)
	}
	if !taskOwned {
		return reconciliation.ErrTaskTransitionRejected
	}
	if !statePersisted {
		return ErrDataQualityFlightStateNotPersisted
	}
	if reportID == "" {
		return errors.New(
			"upsert reconciled flight state quality report returned no report",
		)
	}

	return nil
}
