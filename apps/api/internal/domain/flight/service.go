package flight

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

var (
	ErrServiceFlightIDRequired        = errors.New("flight service identifier is required")
	ErrServiceRepositoryResultInvalid = errors.New("flight service repository result is invalid")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	if err := dependency.Require("flight repository", repository); err != nil {
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

func (s *Service) List(ctx context.Context) ([]Flight, error) {
	items, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return make([]Flight, 0), nil
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

func (s *Service) GetByID(ctx context.Context, id string) (Flight, error) {
	normalized := strings.TrimSpace(id)
	if normalized == "" {
		return Flight{}, ErrServiceFlightIDRequired
	}
	item, err := s.repository.GetByID(ctx, normalized)
	if err != nil {
		return Flight{}, err
	}
	if err := item.Validate(); err != nil {
		return Flight{}, fmt.Errorf(
			"%w: %w",
			ErrServiceRepositoryResultInvalid,
			err,
		)
	}
	return item, nil
}
