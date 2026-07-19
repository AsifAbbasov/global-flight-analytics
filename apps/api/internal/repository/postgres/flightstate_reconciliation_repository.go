package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/jackc/pgx/v5/pgtype"
)

func (repository *FlightStateRepository) ListByReconciliationScope(
	ctx context.Context,
	icao24 string,
	ingestionRunID string,
	observedFrom time.Time,
	observedTo time.Time,
) ([]flightstate.FlightState, error) {
	if repository == nil || repository.db == nil {
		return nil, ErrFlightStateRepositoryPoolRequired
	}

	normalizedICAO24 := strings.ToUpper(
		strings.TrimSpace(
			icao24,
		),
	)
	if normalizedICAO24 == "" {
		return nil, fmt.Errorf(
			"list reconciliation flight states: icao24 is required",
		)
	}

	if observedFrom.IsZero() || observedTo.IsZero() {
		return nil, fmt.Errorf(
			"list reconciliation flight states: observed range is required",
		)
	}

	observedFrom = observedFrom.UTC()
	observedTo = observedTo.UTC()

	if observedFrom.After(
		observedTo,
	) {
		return nil, fmt.Errorf(
			"list reconciliation flight states: observed range is invalid",
		)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		SELECT
			id::text,
			COALESCE(flight_id::text, ''),
			COALESCE(aircraft_id::text, ''),
			COALESCE(ingestion_run_id::text, ''),
			icao24,
			COALESCE(callsign, ''),
			latitude::double precision,
			longitude::double precision,
			barometric_altitude_m::double precision,
			barometric_altitude_status,
			geometric_altitude_m::double precision,
			geometric_altitude_status,
			velocity_mps::double precision,
			heading_degrees::double precision,
			vertical_rate_mps::double precision,
			on_ground,
			COALESCE(origin_country, ''),
			observed_at,
			source_name
		FROM flight_states
		WHERE icao24 = $1
			AND latitude IS NOT NULL
			AND longitude IS NOT NULL
			AND observed_at >= $2
			AND observed_at <= $3
			AND (
				$4::uuid IS NULL
				OR ingestion_run_id = $4::uuid
			)
		ORDER BY
			observed_at ASC,
			id ASC;
	`

	rows, err := repository.db.Query(
		ctx,
		query,
		normalizedICAO24,
		observedFrom,
		observedTo,
		nullableUUID(
			strings.TrimSpace(
				ingestionRunID,
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query reconciliation flight states: %w",
			err,
		)
	}
	defer rows.Close()

	items := make(
		[]flightstate.FlightState,
		0,
	)

	for rows.Next() {
		var item flightstate.FlightState
		var barometricAltitude pgtype.Float8
		var geometricAltitude pgtype.Float8
		var velocity pgtype.Float8
		var heading pgtype.Float8
		var verticalRate pgtype.Float8
		var onGround pgtype.Bool
		var barometricStatus string
		var geometricStatus string

		if err := rows.Scan(
			&item.ID,
			&item.FlightID,
			&item.AircraftID,
			&item.IngestionRunID,
			&item.ICAO24,
			&item.Callsign,
			&item.Latitude,
			&item.Longitude,
			&barometricAltitude,
			&barometricStatus,
			&geometricAltitude,
			&geometricStatus,
			&velocity,
			&heading,
			&verticalRate,
			&onGround,
			&item.OriginCountry,
			&item.ObservedAt,
			&item.SourceName,
		); err != nil {
			return nil, fmt.Errorf(
				"scan reconciliation flight state: %w",
				err,
			)
		}

		applyAltitudeDatabaseValues(
			&item,
			barometricAltitude,
			barometricStatus,
			geometricAltitude,
			geometricStatus,
		)
		applyTelemetryDatabaseValues(
			&item,
			velocity,
			heading,
			verticalRate,
			onGround,
		)

		item.ObservedAt = item.ObservedAt.UTC()

		items = append(
			items,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate reconciliation flight states: %w",
			err,
		)
	}

	return items, nil
}
