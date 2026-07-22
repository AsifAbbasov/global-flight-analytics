package flightstate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

var (
	ErrServiceFlightIDRequired        = errors.New("flight state service flight identifier is required")
	ErrServiceICAO24Required          = errors.New("flight state service ICAO24 is required")
	ErrServiceRepositoryResultInvalid = errors.New("flight state service repository result is invalid")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	if err := dependency.Require("flight state repository", repository); err != nil {
		return nil, err
	}
	return &Service{repository: repository}, nil
}

func MustNewService(repository Repository) *Service {
	service, err := NewService(repository)
	if err != nil {
		panic(err)
	}
	return service
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
	for index, item := range items {
		if err := item.Validate(); err != nil {
			return nil, fmt.Errorf(
				"%w: index=%d: %w",
				ErrServiceRepositoryResultInvalid,
				index,
				err,
			)
		}
	}
	return items, nil
}

func (s *Service) GetLatestByICAO24(ctx context.Context, icao24 string) (FlightState, error) {
	normalized := strings.ToLower(strings.TrimSpace(icao24))
	if normalized == "" {
		return FlightState{}, ErrServiceICAO24Required
	}
	item, err := s.repository.GetLatestByICAO24(ctx, normalized)
	if err != nil {
		return FlightState{}, err
	}
	if err := item.Validate(); err != nil {
		return FlightState{}, fmt.Errorf(
			"%w: %w",
			ErrServiceRepositoryResultInvalid,
			err,
		)
	}
	return item, nil
}
