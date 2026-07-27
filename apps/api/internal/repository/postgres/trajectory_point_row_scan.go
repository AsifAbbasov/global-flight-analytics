package postgres

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func scanTrajectoryPoint(
	scanner postgresRowScanner,
) (trajectory.TrackPoint4D, error) {
	var state flightstate.FlightState
	var latitude pgtype.Float8
	var longitude pgtype.Float8
	var barometricAltitude pgtype.Float8
	var geometricAltitude pgtype.Float8
	var velocity pgtype.Float8
	var heading pgtype.Float8
	var verticalRate pgtype.Float8
	var onGround pgtype.Bool
	var barometricStatus string
	var geometricStatus string

	if err := scanner.Scan(
		&state.ID,
		&state.FlightID,
		&state.AircraftID,
		&state.ICAO24,
		&state.Callsign,
		&latitude,
		&longitude,
		&barometricAltitude,
		&barometricStatus,
		&geometricAltitude,
		&geometricStatus,
		&velocity,
		&heading,
		&verticalRate,
		&onGround,
		&state.OriginCountry,
		&state.ObservedAt,
		&state.SourceName,
	); err != nil {
		return trajectory.TrackPoint4D{}, err
	}
	if !latitude.Valid || !longitude.Valid {
		return trajectory.TrackPoint4D{}, fmt.Errorf(
			"trajectory point %q is missing required position",
			state.ID,
		)
	}
	state.Latitude = latitude.Float64
	state.Longitude = longitude.Float64
	applyAltitudeDatabaseValues(
		&state,
		barometricAltitude,
		barometricStatus,
		geometricAltitude,
		geometricStatus,
	)
	applyTelemetryDatabaseValues(
		&state,
		velocity,
		heading,
		verticalRate,
		onGround,
	)
	return trajectory.TrackPoint4DFromFlightState(state), nil
}

func scanTrajectoryPointRows(
	rows pgx.Rows,
) ([]trajectory.TrackPoint4D, error) {
	items := make([]trajectory.TrackPoint4D, 0)
	for rows.Next() {
		item, err := scanTrajectoryPoint(rows)
		if err != nil {
			return nil, fmt.Errorf("scan trajectory point: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trajectory points: %w", err)
	}
	return items, nil
}
