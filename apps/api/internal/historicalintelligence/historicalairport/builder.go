package historicalairport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

	airportICAOCode := strings.ToUpper(
		strings.TrimSpace(request.AirportICAOCode),
	)
	if !airportICAOPattern.MatchString(
		airportICAOCode,
	) {
		return historicalcontract.Result{},
			ErrAirportICAOInvalid
	}

	definition, ok := airportMetricDefinition(
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
		var result routecontract.Result
		if err := json.Unmarshal(
			record.RouteJSON,
			&result,
		); err != nil {
			invalidPayloadCount++
			continue
		}

		if result.Window.AsOfTime.After(
			request.Plan.AsOfTime,
		) {
			invalidPayloadCount++
			continue
		}

		decoded = append(decoded, result)
	}

	values, missingIdentityCount := airportValues(
		request.Plan.Buckets,
		decoded,
		airportICAOCode,
		request.MetricName,
	)

	coverageRatio := routeCoverage(
		len(selected),
		len(decoded),
		request.Snapshot.RouteLimitReached,
	)
	limitations := []historicalcontract.Limitation{
		{
			Code:    "probable_airport_activity_only",
			Message: "Historical airport activity is derived from probable Route Intelligence endpoints rather than filed flight-plan data.",
			Scope:   "series",
		},
	}

	if invalidPayloadCount > 0 {
		limitations = append(
			limitations,
			historicalcontract.Limitation{
				Code: "historical_route_payload_invalid",
				Message: fmt.Sprintf(
					"%d persisted Route Intelligence payloads could not be decoded or exceeded the analytical as-of time.",
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
				Message: "The bounded historical route read reached its dataset limit; airport activity coverage is a conservative lower bound.",
				Scope:   "series",
			},
		)
	}
	if missingIdentityCount > 0 &&
		request.MetricName ==
			historicalcontract.MetricNameUniqueAircraft {
		limitations = append(
			limitations,
			historicalcontract.Limitation{
				Code: "historical_aircraft_identity_unavailable",
				Message: fmt.Sprintf(
					"%d airport activity events lacked aircraft identity and were excluded from unique-aircraft values.",
					missingIdentityCount,
				),
				Scope: "series",
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
			Scope: historicalcontract.Scope{
				Type:            historicalcontract.ScopeTypeAirport,
				AirportICAOCode: airportICAOCode,
			},
			Plan:              request.Plan,
			Values:            values,
			DataCoverageRatio: coverageRatio,
			BuilderVersion:    Version,
			InputFingerprint:  airportFingerprint(request, airportICAOCode),
			SourceNames: []string{
				"flight_route_results",
				"route_intelligence",
			},
			LatestSourceUpdatedAt: latestRouteUpdate(
				selected,
				request.Plan.AsOfTime,
			),
			GeneratedAt: request.GeneratedAt,
			Limitations: limitations,
		},
	)
}

type airportMetricSpec struct {
	unit        string
	aggregation historicalcontract.Aggregation
}

func airportMetricDefinition(
	name historicalcontract.MetricName,
) (airportMetricSpec, bool) {
	switch name {
	case historicalcontract.MetricNameAirportDepartures:
		return airportMetricSpec{
			unit:        "departures",
			aggregation: historicalcontract.AggregationCount,
		}, true

	case historicalcontract.MetricNameAirportArrivals:
		return airportMetricSpec{
			unit:        "arrivals",
			aggregation: historicalcontract.AggregationCount,
		}, true

	case historicalcontract.MetricNameAirportOperations:
		return airportMetricSpec{
			unit:        "operations",
			aggregation: historicalcontract.AggregationCount,
		}, true

	case historicalcontract.MetricNameUniqueAircraft:
		return airportMetricSpec{
			unit:        "aircraft",
			aggregation: historicalcontract.AggregationCount,
		}, true

	default:
		return airportMetricSpec{}, false
	}
}

type airportEvent struct {
	observedAt time.Time
	identity   string
}

func airportValues(
	buckets []historicalwindow.Bucket,
	routes []routecontract.Result,
	airportICAOCode string,
	metricName historicalcontract.MetricName,
) ([]historicalseries.BucketValue, int) {
	values := make(
		[]historicalseries.BucketValue,
		len(buckets),
	)
	for index, bucket := range buckets {
		values[index].Bucket = bucket
	}

	events := make([]airportEvent, 0)
	for _, route := range routes {
		identity := strings.TrimSpace(route.AircraftID)
		if identity == "" {
			identity = strings.ToUpper(
				strings.TrimSpace(route.ICAO24),
			)
		}

		if metricName ==
			historicalcontract.MetricNameAirportDepartures ||
			metricName ==
				historicalcontract.MetricNameAirportOperations ||
			metricName ==
				historicalcontract.MetricNameUniqueAircraft {
			if route.Origin != nil &&
				strings.ToUpper(
					strings.TrimSpace(
						route.Origin.Airport.ICAOCode,
					),
				) == airportICAOCode {
				events = append(
					events,
					airportEvent{
						observedAt: route.Window.StartTime,
						identity:   identity,
					},
				)
			}
		}

		if metricName ==
			historicalcontract.MetricNameAirportArrivals ||
			metricName ==
				historicalcontract.MetricNameAirportOperations ||
			metricName ==
				historicalcontract.MetricNameUniqueAircraft {
			if route.Destination != nil &&
				strings.ToUpper(
					strings.TrimSpace(
						route.Destination.Airport.ICAOCode,
					),
				) == airportICAOCode {
				events = append(
					events,
					airportEvent{
						observedAt: route.Window.EndTime,
						identity:   identity,
					},
				)
			}
		}
	}

	missingIdentityCount := 0
	uniqueAircraft := make(
		[]map[string]struct{},
		len(buckets),
	)
	for index := range uniqueAircraft {
		uniqueAircraft[index] =
			make(map[string]struct{})
	}

	for _, event := range events {
		index := airportBucketIndex(
			buckets,
			event.observedAt,
		)
		if index < 0 {
			continue
		}

		if metricName ==
			historicalcontract.MetricNameUniqueAircraft {
			if event.identity == "" {
				missingIdentityCount++
				continue
			}

			values[index].SampleCount++
			uniqueAircraft[index][event.identity] =
				struct{}{}
			continue
		}

		values[index].Value++
		values[index].SampleCount++
	}

	if metricName ==
		historicalcontract.MetricNameUniqueAircraft {
		for index := range values {
			values[index].Value =
				float64(len(uniqueAircraft[index]))
		}
	}

	return values, missingIdentityCount
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

		key := strings.TrimSpace(record.TrajectoryID)
		if key == "" {
			key = strings.TrimSpace(record.ID)
		}
		if key == "" {
			continue
		}

		current, exists := latest[key]
		if !exists ||
			record.AsOfTime.After(current.AsOfTime) ||
			(record.AsOfTime.Equal(current.AsOfTime) &&
				record.ID < current.ID) {
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

func airportBucketIndex(
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

	if ratio < 0 {
		return 0
	}
	if ratio > 1 {
		return 1
	}
	return ratio
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

func airportFingerprint(
	request Request,
	airportICAOCode string,
) string {
	records := []string{
		Version,
		string(request.MetricName),
		airportICAOCode,
		request.Plan.Fingerprint,
		request.Plan.AsOfTime.UTC().
			Format(time.RFC3339Nano),
		fmt.Sprintf(
			"route_limit_reached|%t",
			request.Snapshot.RouteLimitReached,
		),
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
