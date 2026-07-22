package aircraft

import (
	"context"
	"errors"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

var ErrServiceICAO24Required = errors.New("aircraft service ICAO24 is required")

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	dependency.Must("aircraft repository", repository)
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context) ([]Aircraft, error) {
	items, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return make([]Aircraft, 0), nil
	}
	return items, nil
}

func (s *Service) GetByICAO24(ctx context.Context, icao24 string) (Aircraft, error) {
	normalized := strings.ToLower(strings.TrimSpace(icao24))
	if normalized == "" {
		return Aircraft{}, ErrServiceICAO24Required
	}
	return s.repository.GetByICAO24(ctx, normalized)
}
