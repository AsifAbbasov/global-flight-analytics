package historicalroute

import (
	"math"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

const meanEarthRadiusKM = 6371.0088

type routeEvidence struct {
	record      historicalread.RouteRecord
	result      routecontract.Result
	origin      string
	destination string
	distanceKM  float64
}

func decodeRouteEvidenceSet(
	records []historicalread.RouteRecord,
	cutoff time.Time,
) ([]routeEvidence, error) {
	items := make([]routeEvidence, 0, len(records))
	for _, record := range records {
		item, err := decodeRouteEvidence(record, cutoff)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func decodeRouteEvidence(
	record historicalread.RouteRecord,
	cutoff time.Time,
) (routeEvidence, error) {
	result, err := record.ValidatedResultAt(cutoff)
	if err != nil {
		return routeEvidence{}, &EvidenceError{
			RecordID: record.ID,
			Err:      err,
		}
	}

	item := routeEvidence{
		record: record,
		result: result,
	}
	if result.Origin != nil {
		item.origin = normalizedAirportCode(result.Origin.Airport.ICAOCode)
	}
	if result.Destination != nil {
		item.destination = normalizedAirportCode(result.Destination.Airport.ICAOCode)
	}
	if result.Status == routecontract.RouteStatusComplete {
		if result.Summary.SameAirport {
			item.distanceKM = 0
		} else {
			item.distanceKM = greatCircleDistanceKM(
				result.Origin.Airport.Latitude,
				result.Origin.Airport.Longitude,
				result.Destination.Airport.Latitude,
				result.Destination.Airport.Longitude,
			)
		}
	}
	return item, nil
}

func filterRouteEvidence(
	items []routeEvidence,
	origin string,
	destination string,
) []routeEvidence {
	if origin == "" && destination == "" {
		return append([]routeEvidence(nil), items...)
	}

	filtered := make([]routeEvidence, 0, len(items))
	for _, item := range items {
		if item.result.Status != routecontract.RouteStatusComplete {
			continue
		}
		if item.origin == origin && item.destination == destination {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func greatCircleDistanceKM(
	originLatitude float64,
	originLongitude float64,
	destinationLatitude float64,
	destinationLongitude float64,
) float64 {
	toRadians := math.Pi / 180
	originLatitude *= toRadians
	originLongitude *= toRadians
	destinationLatitude *= toRadians
	destinationLongitude *= toRadians

	latitudeDelta := destinationLatitude - originLatitude
	longitudeDelta := destinationLongitude - originLongitude
	halfLatitude := math.Sin(latitudeDelta / 2)
	halfLongitude := math.Sin(longitudeDelta / 2)
	haversine := halfLatitude*halfLatitude +
		math.Cos(originLatitude)*math.Cos(destinationLatitude)*
			halfLongitude*halfLongitude
	if haversine < 0 {
		haversine = 0
	}
	if haversine > 1 {
		haversine = 1
	}
	return 2 * meanEarthRadiusKM * math.Asin(math.Sqrt(haversine))
}
