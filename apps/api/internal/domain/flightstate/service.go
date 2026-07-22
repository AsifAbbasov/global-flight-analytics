package flightstate

import (
	"context"
	"errors"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

var (
	ErrServiceFlightIDRequired = errors.New("flight state service flight identifier is required")
	ErrServiceICAO24Required   = errors.New("flight state service ICAO24 is required")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	dependency.Must("flight state repository", repository)
	return &Service{repository: repository}
}

func (s *Service) ListByFlightID(ctx context.Context, flightID string) ([]FlightState, error) {
	normalized := strings.TrimSpace(flightID)
	if normalized == "" {
		return nil, ErrServiceFlightIDRequired
	}

	items, err := s.repository.ListByFlightID(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return make([]FlightState, 0), nil
	}
	return items, nil
}

func (s *Service) GetLatestByICAO24(ctx context.Context, icao24 string) (FlightState, error) {
	normalized := strings.ToLower(strings.TrimSpace(icao24))
	if normalized == "" {
		return FlightState{}, ErrServiceICAO24Required
	}
	return s.repository.GetLatestByICAO24(ctx, normalized)
}
