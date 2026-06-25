package traffic

import (
	"context"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

type Service struct {
	repository    Repository
	regionService *region.Service
}

func NewService(repository Repository, regionService *region.Service) *Service {
	return &Service{
		repository:    repository,
		regionService: regionService,
	}
}

func (s *Service) GetCurrent(ctx context.Context) ([]CurrentTrafficItem, error) {
	return s.repository.GetCurrent(ctx)
}

func (s *Service) GetCurrentByRegion(
	ctx context.Context,
	regionCode string,
) ([]CurrentTrafficItem, error) {
	selectedRegion, err := s.regionService.GetByCode(regionCode)
	if err != nil {
		return nil, err
	}

	bounds := Bounds{
		MinLatitude:  selectedRegion.Bounds.MinLatitude,
		MaxLatitude:  selectedRegion.Bounds.MaxLatitude,
		MinLongitude: selectedRegion.Bounds.MinLongitude,
		MaxLongitude: selectedRegion.Bounds.MaxLongitude,
	}

	return s.repository.GetCurrentByBounds(ctx, bounds)
}
