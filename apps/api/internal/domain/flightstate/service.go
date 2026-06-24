package flightstate

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) ListByFlightID(ctx context.Context, flightID string) ([]FlightState, error) {
	return s.repository.ListByFlightID(ctx, flightID)
}

func (s *Service) GetLatestByICAO24(ctx context.Context, icao24 string) (FlightState, error) {
	return s.repository.GetLatestByICAO24(ctx, icao24)
}
