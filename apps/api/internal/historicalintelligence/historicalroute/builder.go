package historicalroute

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalseries"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

var airportICAOPattern = regexp.MustCompile(
	`^[A-Z0-9]{4}$`,
)

func Build(
	request Request,
) (historicalcontract.Result, error) {
	if request.Snapshot.Version !=
		historicalread.Version {
		return historicalcontract.Result{},
			ErrSnapshotVersionInvalid
	}

	scope, originICAOCode, destinationICAOCode, err :=
		normalizeScope(
			request.OriginICAOCode,
			request.DestinationICAOCode,
		)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	definition, ok := routeMetricDefinition(
		request.MetricName,
	)
	if !ok {
		return historicalcontract.Result{},
			ErrMetricUnsupported
	}

	selected := latestRouteRecords(
		request.Snapshot.Routes,
		request.Plan.AsOfTime,
	)

	decoded := make(
		[]routecontract.Result,
		0,
		len(selected),
	)
	invalidPayloadCount := 0
	for _, record := range selected {
		result, valid := decodeRouteResult(
			record,
			request.Plan.AsOfTime,
		)
		if !valid {
			invalidPayloadCount++
			continue
		}

		if !matchesRouteScope(
			result,
			originICAOCode,
			destinationICAOCode,
		) {
			continue
		}

		decoded = append(decoded, result)
	}

	values := routeValues(
		request.Plan.Buckets,
		decoded,
		request.MetricName,
	)

	coverageRatio := routeCoverage(
		len(selected),
		len(selected)-invalidPayloadCount,
		request.Snapshot.RouteLimitReached,
	)
	limitations := []historicalcontract.Limitation{
		{
			Code:    "probable_route_intelligence_only",
			Message: "Historical route metrics are derived from probable Route Intelligence results rather than filed flight-plan data.",
			Scope:   "series",
		},
	}

	if invalidPayloadCount > 0 {
		limitations = append(
			limitations,
			historicalcontract.Limitation{
				Code: "historical_route_payload_invalid",
				Message: fmt.Sprintf(
					"%d persisted Route Intelligence payloads were invalid, unsupported, or exceeded the analytical as-of time.",
					invalidPayloadCount,
				),
				Scope: "series",
			},
		)
	}
	if request.Snapshot.RouteLimitReached {
		limitations = append(
			limitations,
			historicalcontract.Limitation{
				Code:    "historical_route_dataset_limit_reached",
				Message: "The bounded historical route read reached its dataset limit; represented route coverage is a conservative lower bound.",
				Scope:   "series",
			},
		)
	}

	return historicalseries.Build(
		historicalseries.BuildRequest{
			Metric: historicalcontract.Metric{
				Name:        request.MetricName,
				Unit:        definition.unit,
				Aggregation: definition.aggregation,
			},
			Scope:                 scope,
			Plan:                  request.Plan,
			Values:                values,
			DataCoverageRatio:     coverageRatio,
			BuilderVersion:        Version,
			InputFingerprint:      routeFingerprint(request, originICAOCode, destinationICAOCode),
			SourceNames:           []string{"flight_route_results", "route_intelligence"},
			LatestSourceUpdatedAt: latestRouteUpdate(selected, request.Plan.AsOfTime),
			GeneratedAt:           request.GeneratedAt,
			Limitations:           limitations,
		},
	)
}

type routeMetricSpec struct {
	unit        string
	aggregation historicalcontract.Aggregation
}

func routeMetricDefinition(
	name historicalcontract.MetricName,
) (routeMetricSpec, bool) {
	switch name {
	case historicalcontract.MetricNameActiveRoutes:
		return routeMetricSpec{
			unit:        "routes",
			aggregation: historicalcontract.AggregationCount,
		}, true

	case historicalcontract.MetricNameRouteObservations:
		return routeMetricSpec{
			unit:        "route_results",
			aggregation: historicalcontract.AggregationCount,
		}, true

	case historicalcontract.MetricNameRouteConfidence:
		return routeMetricSpec{
			unit:        "ratio",
			aggregation: historicalcontract.AggregationAverage,
		}, true

	case historicalcontract.MetricNameCompleteRouteRatio,
		historicalcontract.MetricNamePartialRouteRatio,
		historicalcontract.MetricNameUnavailableRouteRatio:
		return routeMetricSpec{
			unit:        "ratio",
			aggregation: historicalcontract.AggregationRatio,
		}, true

	case historicalcontract.MetricNameGreatCircleDistanceKM:
		return routeMetricSpec{
			unit:        "kilometres",
			aggregation: historicalcontract.AggregationAverage,
		}, true

	default:
		return routeMetricSpec{}, false
	}
}

func normalizeScope(
	origin string,
	destination string,
) (
	historicalcontract.Scope,
	string,
	string,
	error,
) {
	origin = strings.ToUpper(
		strings.TrimSpace(origin),
	)
	destination = strings.ToUpper(
		strings.TrimSpace(destination),
	)

	if origin == "" && destination == "" {
		return historicalcontract.Scope{
				Type: historicalcontract.ScopeTypeGlobal,
			},
			"",
			"",
			nil
	}
	if origin == "" || destination == "" {
		return historicalcontract.Scope{},
			"",
			"",
			ErrRouteScopeIncomplete
	}
	if !airportICAOPattern.MatchString(origin) {
		return historicalcontract.Scope{},
			"",
			"",
			ErrOriginICAOInvalid
	}
	if !airportICAOPattern.MatchString(destination) {
		return historicalcontract.Scope{},
			"",
			"",
			ErrDestinationICAOInvalid
	}

	return historicalcontract.Scope{
			Type:                historicalcontract.ScopeTypeRoute,
			OriginICAOCode:      origin,
			DestinationICAOCode: destination,
		},
		origin,
		destination,
		nil
}

func decodeRouteResult(
	record historicalread.RouteRecord,
	asOfTime time.Time,
) (routecontract.Result, bool) {
	var result routecontract.Result
	if err := json.Unmarshal(
		record.RouteJSON,
		&result,
	); err != nil {
		return routecontract.Result{}, false
	}

	if result.SchemaVersion !=
		routecontract.SchemaVersionV1 ||
		!knownRouteStatus(result.Status) ||
		result.Window.StartTime.IsZero() ||
		result.Window.EndTime.IsZero() ||
		result.Window.AsOfTime.IsZero() ||
		!result.Window.StartTime.Before(
			result.Window.EndTime,
		) ||
		result.Window.EndTime.After(
			result.Window.AsOfTime,
		) ||
		result.Window.AsOfTime.After(
			asOfTime.UTC(),
		) ||
		math.IsNaN(result.Confidence.Score) ||
		math.IsInf(result.Confidence.Score, 0) ||
		result.Confidence.Score < 0 ||
		result.Confidence.Score > 1 {
		return routecontract.Result{}, false
	}

	result.Window.StartTime =
		result.Window.StartTime.UTC()
	result.Window.EndTime =
		result.Window.EndTime.UTC()
	result.Window.AsOfTime =
		result.Window.AsOfTime.UTC()

	return result, true
}

func knownRouteStatus(
	status routecontract.RouteStatus,
) bool {
	switch status {
	case routecontract.RouteStatusComplete,
		routecontract.RouteStatusPartial,
		routecontract.RouteStatusUnavailable:
		return true
	default:
		return false
	}
}

func matchesRouteScope(
	result routecontract.Result,
	origin string,
	destination string,
) bool {
	if origin == "" && destination == "" {
		return true
	}
	if result.Origin == nil ||
		result.Destination == nil {
		return false
	}

	return normalizedAirportCode(
		result.Origin.Airport.ICAOCode,
	) == origin &&
		normalizedAirportCode(
			result.Destination.Airport.ICAOCode,
		) == destination
}

func normalizedAirportCode(
	value string,
) string {
	return strings.ToUpper(
		strings.TrimSpace(value),
	)
}

type bucketAccumulator struct {
	routeKeys        map[string]struct{}
	observationCount int
	confidenceTotal  float64
	completeCount    int
	partialCount     int
	unavailableCount int
	distanceTotal    float64
	distanceCount    int
}

func routeValues(
	buckets []historicalwindow.Bucket,
	routes []routecontract.Result,
	metricName historicalcontract.MetricName,
) []historicalseries.BucketValue {
	values := make(
		[]historicalseries.BucketValue,
		len(buckets),
	)
	accumulators := make(
		[]bucketAccumulator,
		len(buckets),
	)
	for index, bucket := range buckets {
		values[index].Bucket = bucket
		accumulators[index].routeKeys =
			make(map[string]struct{})
	}

	for _, result := range routes {
		index := routeBucketIndex(
			buckets,
			result.Window.StartTime,
		)
		if index < 0 {
			continue
		}

		accumulator := &accumulators[index]
		accumulator.observationCount++
		accumulator.confidenceTotal +=
			result.Confidence.Score

		switch result.Status {
		case routecontract.RouteStatusComplete:
			accumulator.completeCount++
		case routecontract.RouteStatusPartial:
			accumulator.partialCount++
		case routecontract.RouteStatusUnavailable:
			accumulator.unavailableCount++
		}

		if result.Origin != nil &&
			result.Destination != nil {
			origin := normalizedAirportCode(
				result.Origin.Airport.ICAOCode,
			)
			destination := normalizedAirportCode(
				result.Destination.Airport.ICAOCode,
			)
			if airportICAOPattern.MatchString(origin) &&
				airportICAOPattern.MatchString(destination) {
				accumulator.routeKeys[origin+"-"+destination] = struct{}{}
			}
		}

		distance := result.Summary.
			GreatCircleDistanceKM
		if result.Origin != nil &&
			result.Destination != nil &&
			!math.IsNaN(distance) &&
			!math.IsInf(distance, 0) &&
			distance >= 0 {
			accumulator.distanceTotal += distance
			accumulator.distanceCount++
		}
	}

	for index, accumulator := range accumulators {
		switch metricName {
		case historicalcontract.MetricNameActiveRoutes:
			values[index].Value =
				float64(len(accumulator.routeKeys))
			values[index].SampleCount =
				accumulator.observationCount

		case historicalcontract.MetricNameRouteObservations:
			values[index].Value =
				float64(accumulator.observationCount)
			values[index].SampleCount =
				accumulator.observationCount

		case historicalcontract.MetricNameRouteConfidence:
			if accumulator.observationCount > 0 {
				values[index].Value =
					accumulator.confidenceTotal /
						float64(
							accumulator.observationCount,
						)
			}
			values[index].SampleCount =
				accumulator.observationCount

		case historicalcontract.MetricNameCompleteRouteRatio:
			values[index].Value = statusRatio(
				accumulator.completeCount,
				accumulator.observationCount,
			)
			values[index].SampleCount =
				accumulator.observationCount

		case historicalcontract.MetricNamePartialRouteRatio:
			values[index].Value = statusRatio(
				accumulator.partialCount,
				accumulator.observationCount,
			)
			values[index].SampleCount =
				accumulator.observationCount

		case historicalcontract.MetricNameUnavailableRouteRatio:
			values[index].Value = statusRatio(
				accumulator.unavailableCount,
				accumulator.observationCount,
			)
			values[index].SampleCount =
				accumulator.observationCount

		case historicalcontract.MetricNameGreatCircleDistanceKM:
			if accumulator.distanceCount > 0 {
				values[index].Value =
					accumulator.distanceTotal /
						float64(
							accumulator.distanceCount,
						)
			}
			values[index].SampleCount =
				accumulator.distanceCount
		}
	}

	return values
}

func statusRatio(
	count int,
	total int,
) float64 {
	if total <= 0 {
		return 0
	}

	return float64(count) / float64(total)
}

func latestRouteRecords(
	records []historicalread.RouteRecord,
	asOfTime time.Time,
) []historicalread.RouteRecord {
	latest := make(
		map[string]historicalread.RouteRecord,
	)
	cutoff := asOfTime.UTC()

	for _, record := range records {
		if record.AsOfTime.IsZero() ||
			record.AsOfTime.After(cutoff) {
			continue
		}

		key := strings.TrimSpace(
			record.TrajectoryID,
		)
		if key == "" {
			key = strings.TrimSpace(record.ID)
		}
		if key == "" {
			continue
		}

		current, exists := latest[key]
		if !exists ||
			record.AsOfTime.After(current.AsOfTime) ||
			(record.AsOfTime.Equal(
				current.AsOfTime,
			) && record.ID < current.ID) {
			latest[key] = record
		}
	}

	result := make(
		[]historicalread.RouteRecord,
		0,
		len(latest),
	)
	for _, record := range latest {
		record.RouteJSON = append(
			[]byte(nil),
			record.RouteJSON...,
		)
		result = append(result, record)
	}

	sort.SliceStable(
		result,
		func(left int, right int) bool {
			if !result[left].AsOfTime.Equal(
				result[right].AsOfTime,
			) {
				return result[left].AsOfTime.Before(
					result[right].AsOfTime,
				)
			}
			return result[left].ID <
				result[right].ID
		},
	)

	return result
}

func routeBucketIndex(
	buckets []historicalwindow.Bucket,
	value time.Time,
) int {
	if value.IsZero() {
		return -1
	}
	normalized := value.UTC()

	index := sort.Search(
		len(buckets),
		func(index int) bool {
			return buckets[index].EndTime.
				After(normalized)
		},
	)
	if index >= len(buckets) ||
		!buckets[index].Contains(normalized) {
		return -1
	}

	return index
}

func routeCoverage(
	selectedCount int,
	decodedCount int,
	limitReached bool,
) float64 {
	if selectedCount == 0 {
		if limitReached {
			return 0.5
		}
		return 1
	}

	ratio := float64(decodedCount) /
		float64(selectedCount)
	if limitReached {
		ratio *= float64(selectedCount) /
			float64(selectedCount+1)
	}

	switch {
	case ratio < 0:
		return 0
	case ratio > 1:
		return 1
	default:
		return ratio
	}
}

func latestRouteUpdate(
	records []historicalread.RouteRecord,
	asOfTime time.Time,
) time.Time {
	result := time.Time{}
	cutoff := asOfTime.UTC()

	for _, record := range records {
		for _, candidate := range []time.Time{
			record.StoredAt,
			record.AsOfTime,
		} {
			if candidate.IsZero() {
				continue
			}
			normalized := candidate.UTC()
			if normalized.After(cutoff) {
				continue
			}
			if normalized.After(result) {
				result = normalized
			}
		}
	}

	return result
}

func routeFingerprint(
	request Request,
	origin string,
	destination string,
) string {
	records := []string{
		Version,
		string(request.MetricName),
		origin,
		destination,
		request.Plan.Fingerprint,
		request.Plan.AsOfTime.UTC().
			Format(time.RFC3339Nano),
	}

	for _, record := range request.Snapshot.Routes {
		payloadHash := sha256.Sum256(
			record.RouteJSON,
		)
		records = append(
			records,
			fmt.Sprintf(
				"route|%s|%s|%s|%s|%s",
				record.ID,
				record.TrajectoryID,
				record.AsOfTime.UTC().
					Format(time.RFC3339Nano),
				record.InputFingerprint,
				hex.EncodeToString(payloadHash[:]),
			),
		)
	}

	sort.Strings(records)
	sum := sha256.Sum256(
		[]byte(strings.Join(records, "\n")),
	)
	return "sha256:" + hex.EncodeToString(sum[:])
}
