package flight

import (
	"context"
	"errors"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

var ErrServiceFlightIDRequired = errors.New("flight service identifier is required")

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	dependency.Must("flight repository", repository)
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context) ([]Flight, error) {
	items, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return make([]Flight, 0), nil
	}
	return items, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Flight, error) {
	normalized := strings.TrimSpace(id)
	if normalized == "" {
		return Flight{}, ErrServiceFlightIDRequired
	}
	return s.repository.GetByID(ctx, normalized)
}
