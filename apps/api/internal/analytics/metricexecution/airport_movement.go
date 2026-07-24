package metricexecution

import (
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const airportMovementEarthRadiusKilometers = 6371.0088

type movementRole string

const (
	movementRoleAmbiguous movementRole = "ambiguous"
	movementRoleArrival   movementRole = "arrival"
	movementRoleDeparture movementRole = "departure"
	movementRoleUnrelated movementRole = "unrelated"
)

type geographicEndpoint struct {
	latitude  float64
	longitude float64
}

func classifyAirportMovement(
	item trajectory.FlightTrajectory,
	airportLatitude float64,
	airportLongitude float64,
	radiusKilometers float64,
) movementRole {
	start, end, available := trajectoryGeographicEndpoints(item)
	if !available {
		return movementRoleAmbiguous
	}

	startInside := geographicDistanceKilometers(
		start.latitude,
		start.longitude,
		airportLatitude,
		airportLongitude,
	) <= radiusKilometers
	endInside := geographicDistanceKilometers(
		end.latitude,
		end.longitude,
		airportLatitude,
		airportLongitude,
	) <= radiusKilometers

	switch {
	case !startInside && endInside:
		return movementRoleArrival
	case startInside && !endInside:
		return movementRoleDeparture
	case !startInside && !endInside:
		return movementRoleUnrelated
	default:
		return movementRoleAmbiguous
	}
}

func trajectoryGeographicEndpoints(
	item trajectory.FlightTrajectory,
) (
	geographicEndpoint,
	geographicEndpoint,
	bool,
) {
	if len(item.Points) > 0 {
		first := item.Points[0]
		last := item.Points[0]

		for _, point := range item.Points[1:] {
			if point.ObservedAt.Before(first.ObservedAt) {
				first = point
			}
			if point.ObservedAt.After(last.ObservedAt) {
				last = point
			}
		}

		start := geographicEndpoint{
			latitude:  first.Latitude,
			longitude: first.Longitude,
		}
		end := geographicEndpoint{
			latitude:  last.Latitude,
			longitude: last.Longitude,
		}
		return start, end,
			validGeographicEndpoint(start) &&
				validGeographicEndpoint(end)
	}

	if len(item.Segments) > 0 {
		first := item.Segments[0]
		last := item.Segments[0]

		for _, segment := range item.Segments[1:] {
			if segment.StartTime.Before(first.StartTime) {
				first = segment
			}
			if segment.EndTime.After(last.EndTime) {
				last = segment
			}
		}

		start := geographicEndpoint{
			latitude:  first.StartLatitude,
			longitude: first.StartLongitude,
		}
		end := geographicEndpoint{
			latitude:  last.EndLatitude,
			longitude: last.EndLongitude,
		}
		return start, end,
			validGeographicEndpoint(start) &&
				validGeographicEndpoint(end)
	}

	return geographicEndpoint{},
		geographicEndpoint{},
		false
}

func validGeographicEndpoint(
	value geographicEndpoint,
) bool {
	return !math.IsNaN(value.latitude) &&
		!math.IsInf(value.latitude, 0) &&
		value.latitude >= -90 &&
		value.latitude <= 90 &&
		!math.IsNaN(value.longitude) &&
		!math.IsInf(value.longitude, 0) &&
		value.longitude >= -180 &&
		value.longitude <= 180
}

func geographicDistanceKilometers(
	latitudeOne float64,
	longitudeOne float64,
	latitudeTwo float64,
	longitudeTwo float64,
) float64 {
	latitudeOneRadians := latitudeOne * math.Pi / 180
	latitudeTwoRadians := latitudeTwo * math.Pi / 180
	latitudeDelta := (latitudeTwo - latitudeOne) * math.Pi / 180
	longitudeDelta := (longitudeTwo - longitudeOne) * math.Pi / 180

	sineLatitude := math.Sin(latitudeDelta / 2)
	sineLongitude := math.Sin(longitudeDelta / 2)

	haversine := sineLatitude*sineLatitude +
		math.Cos(latitudeOneRadians)*
			math.Cos(latitudeTwoRadians)*
			sineLongitude*
			sineLongitude

	return 2 *
		airportMovementEarthRadiusKilometers *
		math.Asin(
			math.Sqrt(
				math.Min(1, haversine),
			),
		)
}
