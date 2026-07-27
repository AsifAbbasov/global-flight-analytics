package historicalairport

import (
	"crypto/sha256"
	"encoding/hex"
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

	canonicalPlan, err := historicalwindow.CanonicalizePlan(
		request.Plan,
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}
	request.Plan = canonicalPlan

	airportICAOCode := strings.ToUpper(
		strings.TrimSpace(request.AirportICAOCode),
	)
	if !airportICAOPattern.MatchString(
		airportICAOCode,
	) {
		return historicalcontract.Result{},
			ErrAirportICAOInvalid
	}

	definition, ok := historicalcontract.MetricSpecFor(request.MetricName)
	if !ok || definition.Family != historicalcontract.MetricFamilyAirport {
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
		result, valid := record.ResultAt(
			request.Plan.AsOfTime,
		)
		if !valid {
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

	routeMatchedCount := request.Snapshot.TotalForSource(
		historicalread.DatasetRoutes,
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
				Message: "The bounded historical route read reached its dataset limit; airport activity coverage uses the exact matched-route denominator.",
				Scope:   "series",
			},
		)
	}
	if request.Snapshot.RouteByteLimitReached {
		limitations = append(
			limitations,
			historicalcontract.Limitation{
				Code:    "historical_route_payload_byte_limit_reached",
				Message: "The bounded historical route read reached its payload byte budget; unrepresented routes remain explicitly incomplete.",
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

	coverageState := historicalseries.DatasetReadComplete
	if request.Snapshot.RouteLimitReached ||
		request.Snapshot.RouteByteLimitReached ||
		invalidPayloadCount > 0 ||
		(missingIdentityCount > 0 &&
			request.MetricName ==
				historicalcontract.MetricNameUniqueAircraft) {
		coverageState = historicalseries.DatasetReadIncomplete
	}
	values, err = historicalseries.BindDatasetCoverage(
		values,
		historicalseries.DatasetCoverage{
			State:        coverageState,
			MatchedCount: routeMatchedCount,
		},
	)
	if err != nil {
		return historicalcontract.Result{}, err
	}

	return historicalseries.Build(
		historicalseries.BuildRequest{
			Metric: historicalcontract.Metric{
				Name:        request.MetricName,
				Unit:        definition.Unit,
				Aggregation: definition.Aggregation,
			},
			Scope: historicalcontract.Scope{
				Type:            historicalcontract.ScopeTypeAirport,
				AirportICAOCode: airportICAOCode,
			},
			Plan:             request.Plan,
			Values:           values,
			BuilderVersion:   Version,
			InputFingerprint: airportFingerprint(request, airportICAOCode),
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
		record.Result = record.Result.Clone()
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
	matchedCount int64,
) float64 {
	if decodedCount < 0 || decodedCount > selectedCount {
		return 0
	}
	return historicalread.RepresentedCoverage(
		decodedCount,
		matchedCount,
	)
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
		fmt.Sprintf(
			"route_byte_limit_reached|%t",
			request.Snapshot.RouteByteLimitReached,
		),
		fmt.Sprintf(
			"route_matched_count|%d",
			request.Snapshot.TotalForSource(
				historicalread.DatasetRoutes,
			),
		),
		fmt.Sprintf(
			"route_payload_bytes|%d|%d",
			request.Snapshot.RoutePayloadBytes,
			request.Snapshot.RouteTotalPayloadBytes,
		),
	}

	for _, record := range request.Snapshot.Routes {
		records = append(
			records,
			fmt.Sprintf(
				"route|%s|%s|%s|%s|%s",
				record.ID,
				record.TrajectoryID,
				record.AsOfTime.UTC().
					Format(time.RFC3339Nano),
				record.InputFingerprint,
				record.PayloadDigest(),
			),
		)
	}

	sort.Strings(records)
	sum := sha256.Sum256(
		[]byte(strings.Join(records, "\n")),
	)
	return "sha256:" + hex.EncodeToString(sum[:])
}
