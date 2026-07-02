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
	return &DataQualityRepository{
		db: db,
	}
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
			object_type,
			object_id,
			validation_status,
			completeness,
			confidence,
			score,
			missing_fields,
			warnings_json
		)
		VALUES (
			'flight_state',
			$1,
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
		return fmt.Errorf("insert data quality report: %w", err)
	}

	return nil
}
