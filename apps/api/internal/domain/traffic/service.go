package traffic

import (
	"context"

	"errors"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dependency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

type RegionResolver interface {
	GetByCode(string) (region.Region, error)
}

type Service struct {
	repository     Repository
	regionResolver RegionResolver
}

var ErrServiceRegionCodeRequired = errors.New("traffic service region code is required")

func NewService(repository Repository, regionResolver RegionResolver) *Service {
	dependency.Must("traffic repository", repository)
	dependency.Must("traffic region resolver", regionResolver)
	return &Service{
		repository:     repository,
		regionResolver: regionResolver,
	}
}

func (s *Service) GetCurrent(ctx context.Context) ([]CurrentTrafficItem, error) {
	items, err := s.repository.GetCurrent(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return make([]CurrentTrafficItem, 0), nil
	}
	return items, nil
}

func (s *Service) GetCurrentByRegion(
	ctx context.Context,
	regionCode string,
) ([]CurrentTrafficItem, error) {
	normalizedRegionCode := strings.ToLower(strings.TrimSpace(regionCode))
	if normalizedRegionCode == "" {
		return nil, ErrServiceRegionCodeRequired
	}

	selectedRegion, err := s.regionResolver.GetByCode(normalizedRegionCode)
	if err != nil {
		return nil, err
	}

	bounds := Bounds{
		MinLatitude:  selectedRegion.Bounds.MinLatitude,
		MaxLatitude:  selectedRegion.Bounds.MaxLatitude,
		MinLongitude: selectedRegion.Bounds.MinLongitude,
		MaxLongitude: selectedRegion.Bounds.MaxLongitude,
	}

	items, err := s.repository.GetCurrentByBounds(ctx, bounds)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return make([]CurrentTrafficItem, 0), nil
	}
	return items, nil
}
