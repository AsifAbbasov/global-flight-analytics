package postgres

import (
	"context"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (
	repository *TrajectoryRepository,
) listTrajectoryPoints(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) ([]trajectory.TrackPoint4D, error) {
	if err := requireRepositoryContext(ctx); err != nil {
		return nil, err
	}

	query := trajectoryPointsByICAO24AndWindowQuery
	identity := strings.ToUpper(strings.TrimSpace(item.ICAO24))
	if strings.TrimSpace(item.FlightID) != "" {
		query = trajectoryPointsByFlightIDAndWindowQuery
		identity = strings.TrimSpace(item.FlightID)
	}

	rows, err := repository.trajectoryReadExecutor().Query(
		ctx,
		query,
		identity,
		item.StartTime.UTC(),
		item.EndTime.UTC(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTrajectoryPointRows(rows)
}
