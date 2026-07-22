package airport

import (
	"context"
	"errors"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

var ErrServiceICAORequired = errors.New("airport service ICAO code is required")

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	dependency.Must("airport repository", repository)
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context) ([]Airport, error) {
	items, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return make([]Airport, 0), nil
	}
	return items, nil
}

func (s *Service) GetByICAO(ctx context.Context, icao string) (Airport, error) {
	normalized := strings.ToUpper(strings.TrimSpace(icao))
	if normalized == "" {
		return Airport{}, ErrServiceICAORequired
	}
	return s.repository.GetByICAO(ctx, normalized)
}
