package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DataQualityRepository struct {
	db *pgxpool.Pool
}

func NewDataQualityRepository(db *pgxpool.Pool) *DataQualityRepository {
	return &DataQualityRepository{db: db}
}

func (repository *DataQualityRepository) SaveFlightStateQuality(
	ctx context.Context,
	state flightstate.FlightState,
	quality dataquality.DataQuality,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	warningsJSON, err := json.Marshal(quality.Warnings)
	if err != nil {
		return fmt.Errorf("marshal data quality warnings: %w", err)
	}

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
		VALUES (
			$1,
			(
				SELECT persisted_state.id
				FROM flight_states AS persisted_state
				WHERE persisted_state.id = $1
			),
			$2,
			$3,
			$4,
			$5,
			$6,
			$7::jsonb
		);
	`

	_, err = repository.db.Exec(
		ctx,
		query,
		nullableUUID(state.ID),
		string(quality.ValidationStatus),
		string(quality.Completeness),
		string(quality.Confidence),
		quality.Score,
		quality.MissingFields,
		string(warningsJSON),
	)
	if err != nil {
		return fmt.Errorf("insert flight state quality report: %w", err)
	}

	return nil
}
