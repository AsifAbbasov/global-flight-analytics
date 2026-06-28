package traffic

import (
	"context"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

type Service struct {
	provider Provider
}

func NewService(provider Provider) *Service {
	return &Service{
		provider: provider,
	}
}

func (s *Service) LoadByCallsign(
	ctx context.Context,
	callsign string,
) ([]flightstate.FlightState, error) {
	result, err := s.provider.LoadByCallsign(ctx, callsign)
	if err != nil {
		return nil, fmt.Errorf("load traffic by callsign: %w", err)
	}

	return result, nil
}
