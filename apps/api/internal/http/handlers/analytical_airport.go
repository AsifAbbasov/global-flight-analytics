package handlers

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricexecution"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricquery"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

const airportActivityMinimumQueryRadiusKilometers = 60.0

func parseAirportActivityRadius(
	value string,
) (float64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return metricexecution.DefaultAirportActivityRadiusKilometers, nil
	}

	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil ||
		math.IsNaN(parsed) ||
		math.IsInf(parsed, 0) ||
		parsed <= 0 ||
		parsed > metricexecution.MaximumAirportActivityRadiusKilometers {
		return 0, metricexecution.ErrAirportActivityRadiusInvalid
	}
	return parsed, nil
}

func analyticalAirportError(
	ctx *fiber.Ctx,
	err error,
) error {
	switch {
	case errors.Is(err, airport.ErrNotFound):
		return response.Error(
			ctx,
			fiber.StatusNotFound,
			"AIRPORT_NOT_FOUND",
			"Airport not found",
		)
	case errors.Is(err, airport.ErrServiceICAORequired):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"AIRPORT_ICAO_REQUIRED",
			"airport_icao is required",
		)
	default:
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			"ANALYTICAL_AIRPORT_QUERY_FAILED",
			"Failed to load airport data",
		)
	}
}

func airportActivityQueryBounds(
	selectedAirport airport.Airport,
	radiusKilometers float64,
) (metricquery.Bounds, error) {
	queryRadius := math.Max(
		airportActivityMinimumQueryRadiusKilometers,
		radiusKilometers*4,
	)

	latitudeDelta := queryRadius / 111.32
	minimumLatitude := math.Max(
		-90,
		selectedAirport.Latitude-latitudeDelta,
	)
	maximumLatitude := math.Min(
		90,
		selectedAirport.Latitude+latitudeDelta,
	)

	cosineLatitude := math.Abs(
		math.Cos(selectedAirport.Latitude * math.Pi / 180),
	)
	minimumLongitude := -180.0
	maximumLongitude := 180.0

	if cosineLatitude > 0.01 {
		longitudeDelta := queryRadius / (111.32 * cosineLatitude)
		candidateMinimum := selectedAirport.Longitude - longitudeDelta
		candidateMaximum := selectedAirport.Longitude + longitudeDelta
		if candidateMinimum >= -180 && candidateMaximum <= 180 {
			minimumLongitude = candidateMinimum
			maximumLongitude = candidateMaximum
		}
	}

	bounds := metricquery.Bounds{
		MinLatitude:  minimumLatitude,
		MaxLatitude:  maximumLatitude,
		MinLongitude: minimumLongitude,
		MaxLongitude: maximumLongitude,
	}
	if err := bounds.Validate(); err != nil {
		return metricquery.Bounds{}, err
	}
	return bounds, nil
}

func airportActivityPublicationMetadata(
	items []trajectory.FlightTrajectory,
	resultLimit int,
	selectedAirport airport.Airport,
	radiusKilometers float64,
) metricexecution.PublicationMetadata {
	metadata := trajectoryPublicationMetadata(
		items,
		resultLimit,
	)
	metadata.Limitations = append(
		metadata.Limitations,
		analyticalresult.Notice{
			Code: "airport_activity_geofence",
			Message: fmt.Sprintf(
				"Airport activity for %s is classified by trajectory crossings of a %.1f kilometer radius geofence.",
				selectedAirport.ICAOCode,
				radiusKilometers,
			),
		},
	)
	return metadata
}
