package airport

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

var (
	ErrServiceICAORequired            = errors.New("airport service ICAO code is required")
	ErrServiceRepositoryResultInvalid = errors.New("airport service repository result is invalid")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	if err := dependency.Require("airport repository", repository); err != nil {
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

func (s *Service) List(ctx context.Context) ([]Airport, error) {
	items, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return make([]Airport, 0), nil
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

func (s *Service) GetByICAO(ctx context.Context, icao string) (Airport, error) {
	normalized := strings.ToUpper(strings.TrimSpace(icao))
	if normalized == "" {
		return Airport{}, ErrServiceICAORequired
	}
	item, err := s.repository.GetByICAO(ctx, normalized)
	if err != nil {
		return Airport{}, err
	}
	if err := item.Validate(); err != nil {
		return Airport{}, fmt.Errorf(
			"%w: %w",
			ErrServiceRepositoryResultInvalid,
			err,
		)
	}
	return item, nil
}
