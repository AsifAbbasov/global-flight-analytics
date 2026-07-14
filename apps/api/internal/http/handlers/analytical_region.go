package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricexecution"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricquery"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

type regionalAnalyticalTrajectoryQueryService interface {
	RecentWithinBounds(
		ctx context.Context,
		request metricquery.RecentRequest,
		bounds metricquery.Bounds,
	) ([]trajectory.FlightTrajectory, error)
}

func resolveAnalyticalRegion(
	value string,
) (*region.Region, error) {
	regionCode := strings.ToLower(strings.TrimSpace(value))
	if regionCode == "" {
		return nil, nil
	}

	selectedRegion, err := region.NewService().GetByCode(regionCode)
	if err != nil {
		return nil, err
	}

	return &selectedRegion, nil
}

func analyticalRegionError(
	ctx *fiber.Ctx,
	err error,
) error {
	if errors.Is(err, region.ErrRegionNotFound) {
		return response.Error(
			ctx,
			fiber.StatusNotFound,
			"REGION_NOT_FOUND",
			"Region not found",
		)
	}

	return response.Error(
		ctx,
		fiber.StatusInternalServerError,
		"ANALYTICAL_REGION_CONFIGURATION_INVALID",
		"Analytical region configuration is invalid",
	)
}

func (handler *AnalyticalMetricsHandler) recentTrajectoriesForRegion(
	ctx context.Context,
	request metricquery.RecentRequest,
	selectedRegion *region.Region,
) ([]trajectory.FlightTrajectory, error) {
	if selectedRegion == nil || selectedRegion.Code == "world" {
		return handler.query.Recent(ctx, request)
	}

	regionalQuery, ok := handler.query.(regionalAnalyticalTrajectoryQueryService)
	if !ok {
		return nil, metricquery.ErrRegionalRepositoryUnsupported
	}

	return regionalQuery.RecentWithinBounds(
		ctx,
		request,
		metricQueryBounds(selectedRegion.Bounds),
	)
}

func trafficDensityAreaSquareKilometers(
	areaParameter string,
	selectedRegion *region.Region,
) (float64, error) {
	if selectedRegion == nil {
		return parseRequiredPositiveFloat(areaParameter)
	}

	return metricQueryBounds(
		selectedRegion.Bounds,
	).AreaSquareKilometers()
}

func trajectoryPublicationMetadataForRegion(
	items []trajectory.FlightTrajectory,
	resultLimit int,
	selectedRegion *region.Region,
) metricexecution.PublicationMetadata {
	metadata := trajectoryPublicationMetadata(items, resultLimit)
	if selectedRegion == nil {
		return metadata
	}

	notice := analyticalresult.Notice{
		Code: "regional_bounding_box",
		Message: fmt.Sprintf(
			"Region %q is represented by its configured rectangular geographic bounds.",
			selectedRegion.Name,
		),
	}
	if selectedRegion.Code == "world" {
		notice = analyticalresult.Notice{
			Code:    "global_region_scope",
			Message: "The World region uses the complete global analytical trajectory window.",
		}
	}

	metadata.Limitations = append(
		metadata.Limitations,
		notice,
	)

	return metadata
}

func metricQueryBounds(
	bounds region.Bounds,
) metricquery.Bounds {
	return metricquery.Bounds{
		MinLatitude:  bounds.MinLatitude,
		MaxLatitude:  bounds.MaxLatitude,
		MinLongitude: bounds.MinLongitude,
		MaxLongitude: bounds.MaxLongitude,
	}
}
