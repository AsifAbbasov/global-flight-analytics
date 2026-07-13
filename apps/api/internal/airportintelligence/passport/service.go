package passport

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

type AirportReader interface {
	GetByICAO(ctx context.Context, icao string) (airport.Airport, error)
}

type AnalyticsReader interface {
	GetByICAO(ctx context.Context, icao string) (AnalyticsInput, error)
}

type Clock func() time.Time

type Service struct {
	airports  AirportReader
	analytics AnalyticsReader
	builder   Builder
	clock     Clock
}

func NewService(
	airports AirportReader,
	analytics AnalyticsReader,
	clock Clock,
) (*Service, error) {
	if airports == nil {
		return nil, fmt.Errorf("%w: airport reader is required", ErrInvalidServiceConfiguration)
	}
	if analytics == nil {
		return nil, fmt.Errorf("%w: analytics reader is required", ErrInvalidServiceConfiguration)
	}
	if clock == nil {
		clock = time.Now
	}

	return &Service{
		airports:  airports,
		analytics: analytics,
		builder:   NewBuilder(),
		clock:     clock,
	}, nil
}

func (service *Service) GetByICAO(
	ctx context.Context,
	icao string,
) (Passport, error) {
	normalizedICAO := strings.ToUpper(strings.TrimSpace(icao))
	if normalizedICAO == "" {
		return Passport{}, fmt.Errorf("%w: ICAO code is required", ErrInvalidIdentity)
	}

	source, err := service.airports.GetByICAO(ctx, normalizedICAO)
	if err != nil {
		return Passport{}, fmt.Errorf("load airport %s: %w", normalizedICAO, err)
	}

	analytics, err := service.analytics.GetByICAO(ctx, normalizedICAO)
	if err != nil {
		return Passport{}, fmt.Errorf("load airport analytics %s: %w", normalizedICAO, err)
	}

	result, err := service.builder.Build(source, analytics, service.clock())
	if err != nil {
		return Passport{}, fmt.Errorf("build airport passport %s: %w", normalizedICAO, err)
	}

	return result, nil
}
