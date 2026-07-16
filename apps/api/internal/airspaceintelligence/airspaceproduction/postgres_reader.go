package airspaceproduction

import (
	"context"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const postgresObservationSourceFallback = "postgres_flight_states"

type postgresQueryer interface {
	Query(
		context.Context,
		string,
		...any,
	) (pgx.Rows, error)
}

type PostgresObservationReader struct {
	queryer postgresQueryer
}

func NewPostgresObservationReader(
	pool *pgxpool.Pool,
) (*PostgresObservationReader, error) {
	if pool == nil {
		return nil, ErrPostgresPoolRequired
	}
	return &PostgresObservationReader{
		queryer: pool,
	}, nil
}

func (reader *PostgresObservationReader) ListAirspaceObservations(
	ctx context.Context,
	query ObservationQuery,
) ([]Observation, error) {
	if reader == nil || reader.queryer == nil {
		return nil, ErrPostgresPoolRequired
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := validateObservationQuery(query); err != nil {
		return nil, err
	}

	const statement = `
		SELECT
			fs.id::text,
			COALESCE(fs.flight_id::text, ''),
			COALESCE(fs.aircraft_id::text, ''),
			fs.icao24,
			COALESCE(fs.callsign, ''),
			fs.latitude,
			fs.longitude,
			fs.barometric_altitude_m::double precision,
			fs.barometric_altitude_status,
			fs.geometric_altitude_m::double precision,
			fs.geometric_altitude_status,
			COALESCE(fs.velocity_mps, 0),
			COALESCE(fs.heading_degrees, 0),
			COALESCE(fs.vertical_rate_mps, 0),
			COALESCE(fs.on_ground, false),
			fs.observed_at,
			COALESCE(NULLIF(BTRIM(fs.source_name), ''), $8)
		FROM flight_states fs
		JOIN ingestion_runs ir
		  ON ir.id = fs.ingestion_run_id
		 AND ir.status = $9
		WHERE fs.observed_at >= $1
		  AND fs.observed_at <= $2
		  AND fs.latitude BETWEEN $3 AND $4
		  AND fs.longitude BETWEEN $5 AND $6
		ORDER BY fs.observed_at ASC, fs.icao24 ASC, fs.id ASC
		LIMIT $7;
	`

	rows, err := reader.queryer.Query(
		ctx,
		statement,
		query.WindowStart.UTC(),
		query.WindowEnd.UTC(),
		query.Bounds.MinLatitude,
		query.Bounds.MaxLatitude,
		query.Bounds.MinLongitude,
		query.Bounds.MaxLongitude,
		query.Limit,
		postgresObservationSourceFallback,
		string(ingestionrun.StatusSuccess),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load PostgreSQL airspace observations: %w",
			err,
		)
	}
	defer rows.Close()

	observations := make([]Observation, 0)
	for rows.Next() {
		var observation Observation
		var barometricAltitude pgtype.Float8
		var geometricAltitude pgtype.Float8
		var barometricStatus string
		var geometricStatus string

		if err := rows.Scan(
			&observation.StateID,
			&observation.FlightID,
			&observation.AircraftID,
			&observation.ICAO24,
			&observation.Callsign,
			&observation.Latitude,
			&observation.Longitude,
			&barometricAltitude,
			&barometricStatus,
			&geometricAltitude,
			&geometricStatus,
			&observation.VelocityMetersPerSecond,
			&observation.HeadingDegrees,
			&observation.VerticalRateMetersPerSecond,
			&observation.OnGround,
			&observation.ObservedAt,
			&observation.SourceName,
		); err != nil {
			return nil, fmt.Errorf(
				"scan PostgreSQL airspace observation: %w",
				err,
			)
		}

		observation.ICAO24 = strings.ToUpper(
			strings.TrimSpace(observation.ICAO24),
		)
		observation.Callsign = strings.ToUpper(
			strings.TrimSpace(observation.Callsign),
		)
		observation.SourceName = strings.TrimSpace(observation.SourceName)
		observation.ObservedAt = observation.ObservedAt.UTC()
		observation.AltitudeMeters,
			observation.AltitudeReference = selectAltitude(
			barometricAltitude,
			flightstate.AltitudeStatus(barometricStatus),
			geometricAltitude,
			flightstate.AltitudeStatus(geometricStatus),
		)

		observations = append(
			observations,
			observation,
		)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate PostgreSQL airspace observations: %w",
			err,
		)
	}
	return observations, nil
}

func validateObservationQuery(query ObservationQuery) error {
	if query.WindowStart.IsZero() ||
		query.WindowEnd.IsZero() ||
		!query.WindowStart.Before(query.WindowEnd) ||
		query.Limit < 1 ||
		query.Bounds.MinLatitude < -90 ||
		query.Bounds.MaxLatitude > 90 ||
		query.Bounds.MinLatitude > query.Bounds.MaxLatitude ||
		query.Bounds.MinLongitude < -180 ||
		query.Bounds.MaxLongitude > 180 ||
		query.Bounds.MinLongitude > query.Bounds.MaxLongitude {
		return fmt.Errorf(
			"%w: PostgreSQL observation query",
			ErrInvalidRequest,
		)
	}
	return nil
}

func selectAltitude(
	barometric pgtype.Float8,
	barometricStatus flightstate.AltitudeStatus,
	geometric pgtype.Float8,
	geometricStatus flightstate.AltitudeStatus,
) (*float64, interactiongraph.AltitudeReference) {
	if geometric.Valid &&
		geometricStatus == flightstate.AltitudeStatusObserved {
		value := geometric.Float64
		return &value, interactiongraph.AltitudeReferenceGeometric
	}
	if barometric.Valid &&
		barometricStatus == flightstate.AltitudeStatusObserved {
		value := barometric.Float64
		return &value, interactiongraph.AltitudeReferenceBarometric
	}
	return nil, interactiongraph.AltitudeReferenceUnknown
}
