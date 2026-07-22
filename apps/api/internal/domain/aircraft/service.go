package aircraft

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
)

var (
	ErrServiceICAO24Required          = errors.New("aircraft service ICAO24 is required")
	ErrServiceRepositoryResultInvalid = errors.New("aircraft service repository result is invalid")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	if err := dependency.Require("aircraft repository", repository); err != nil {
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

func (s *Service) List(ctx context.Context) ([]Aircraft, error) {
	items, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return make([]Aircraft, 0), nil
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

func (s *Service) GetByICAO24(ctx context.Context, icao24 string) (Aircraft, error) {
	normalized := strings.ToLower(strings.TrimSpace(icao24))
	if normalized == "" {
		return Aircraft{}, ErrServiceICAO24Required
	}
	item, err := s.repository.GetByICAO24(ctx, normalized)
	if err != nil {
		return Aircraft{}, err
	}
	if err := item.Validate(); err != nil {
		return Aircraft{}, fmt.Errorf(
			"%w: %w",
			ErrServiceRepositoryResultInvalid,
			err,
		)
	}
	return item, nil
}
