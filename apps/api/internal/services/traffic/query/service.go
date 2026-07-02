package query

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

var (
	ErrTrajectoryRepositoryRequired = errors.New("trajectory repository is required")
	ErrInvalidICAO24                = errors.New("invalid icao24")
	ErrInvalidTrajectoryID          = errors.New("invalid trajectory id")
)

var icao24Pattern = regexp.MustCompile(`^[A-F0-9]{6}$`)

type TrajectoryReadRepository interface {
	GetLatestTrajectoryByICAO24(ctx context.Context, icao24 string) (trajectory.FlightTrajectory, error)
	GetTrajectoryByID(ctx context.Context, trajectoryID string) (trajectory.FlightTrajectory, error)
}

type Config struct {
	TrajectoryRepository TrajectoryReadRepository
}

type Service struct {
	trajectoryRepository TrajectoryReadRepository
}

func New(config Config) *Service {
	return &Service{
		trajectoryRepository: config.TrajectoryRepository,
	}
}

func (service *Service) GetLatestTrajectoryByICAO24(
	ctx context.Context,
	icao24 string,
) (trajectory.FlightTrajectory, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if service.trajectoryRepository == nil {
		return trajectory.FlightTrajectory{}, ErrTrajectoryRepositoryRequired
	}

	normalizedICAO24 := normalizeICAO24(icao24)
	if normalizedICAO24 == "" {
		return trajectory.FlightTrajectory{}, ErrInvalidICAO24
	}

	if !icao24Pattern.MatchString(normalizedICAO24) {
		return trajectory.FlightTrajectory{}, ErrInvalidICAO24
	}

	item, err := service.trajectoryRepository.GetLatestTrajectoryByICAO24(ctx, normalizedICAO24)
	if err != nil {
		return trajectory.FlightTrajectory{}, fmt.Errorf(
			"get latest trajectory by icao24 %s: %w",
			normalizedICAO24,
			err,
		)
	}

	return item, nil
}

func (service *Service) GetTrajectoryByID(
	ctx context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if service.trajectoryRepository == nil {
		return trajectory.FlightTrajectory{}, ErrTrajectoryRepositoryRequired
	}

	normalizedTrajectoryID := strings.TrimSpace(trajectoryID)
	if normalizedTrajectoryID == "" {
		return trajectory.FlightTrajectory{}, ErrInvalidTrajectoryID
	}

	item, err := service.trajectoryRepository.GetTrajectoryByID(ctx, normalizedTrajectoryID)
	if err != nil {
		return trajectory.FlightTrajectory{}, fmt.Errorf(
			"get trajectory by id %s: %w",
			normalizedTrajectoryID,
			err,
		)
	}

	return item, nil
}

func normalizeICAO24(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
