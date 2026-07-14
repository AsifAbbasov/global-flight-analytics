package historicaltraffic

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalseries"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func Build(
	request Request,
) (historicalcontract.Result, error) {
	if request.Snapshot.Version !=
		historicalread.Version {
		return historicalcontract.Result{},
			ErrSnapshotVersionInvalid
	}

	definition, ok := metricDefinition(
		request.MetricName,
	)
	if !ok {
		return historicalcontract.Result{},
			ErrMetricUnsupported
	}

	values := make(
		[]historicalseries.BucketValue,
		0,
		len(request.Plan.Buckets),
	)
	limitations := make(
		[]historicalcontract.Limitation,
		0,
		2,
	)

	relevantCount := 0
	latestSourceTime := time.Time{}
	sourceNames := []string{definition.sourceName}

	switch request.MetricName {
	case historicalcontract.MetricNameFlightCount:
		relevantCount = len(request.Snapshot.Flights)
		latestSourceTime = latestFlightUpdate(
			request.Snapshot.Flights,
			request.Plan.AsOfTime,
		)
		values = countFlightStarts(
			request.Plan.Buckets,
			request.Snapshot.Flights,
		)

	case historicalcontract.MetricNameTrajectoryCount:
		relevantCount =
			len(request.Snapshot.Trajectories)
		latestSourceTime = latestTrajectoryUpdate(
			request.Snapshot.Trajectories,
			request.Plan.AsOfTime,
		)
		values = countTrajectoryStarts(
			request.Plan.Buckets,
			request.Snapshot.Trajectories,
		)

	case historicalcontract.MetricNameObservationCount:
		relevantCount =
			len(request.Snapshot.Observations)
		latestSourceTime = latestObservationUpdate(
			request.Snapshot.Observations,
			request.Plan.AsOfTime,
		)
		values = countObservations(
			request.Plan.Buckets,
			request.Snapshot.Observations,
		)

	case historicalcontract.MetricNameActiveAircraft:
		relevantCount =
			len(request.Snapshot.Observations)
		latestSourceTime = latestObservationUpdate(
			request.Snapshot.Observations,
			request.Plan.AsOfTime,
		)
		var missingIdentityCount int
		values, missingIdentityCount =
			countActiveAircraft(
				request.Plan.Buckets,
				request.Snapshot.Observations,
			)
		if missingIdentityCount > 0 {
			limitations = append(
				limitations,
				historicalcontract.Limitation{
					Code: "active_aircraft_identity_unavailable",
					Message: fmt.Sprintf(
						"%d historical observations lacked both aircraft identifier and ICAO24 address and were excluded from unique-aircraft values.",
						missingIdentityCount,
					),
					Scope: "series",
				},
			)
		}

	case historicalcontract.MetricNameTrafficDensity:
		relevantCount =
			len(request.Snapshot.Observations)
		latestSourceTime = latestObservationUpdate(
			request.Snapshot.Observations,
			request.Plan.AsOfTime,
		)
		values = observationRate(
			request.Plan.Buckets,
			request.Snapshot.Observations,
		)
	}

	limitReached := definition.limitReached(
		request.Snapshot,
	)
	coverageRatio := conservativeCoverage(
		relevantCount,
		limitReached,
	)
	if limitReached {
		limitations = append(
			limitations,
			historicalcontract.Limitation{
				Code:    "historical_dataset_limit_reached",
				Message: "The bounded historical read reached its dataset limit; represented coverage is a conservative lower bound.",
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
			Scope: historicalcontract.Scope{
				Type: historicalcontract.ScopeTypeGlobal,
			},
			Plan:                  request.Plan,
			Values:                values,
			DataCoverageRatio:     coverageRatio,
			BuilderVersion:        Version,
			InputFingerprint:      trafficFingerprint(request),
			SourceNames:           sourceNames,
			LatestSourceUpdatedAt: latestSourceTime,
			GeneratedAt:           request.GeneratedAt,
			Limitations:           limitations,
		},
	)
}

type metricSpec struct {
	unit         string
	aggregation  historicalcontract.Aggregation
	sourceName   string
	limitReached func(historicalread.Snapshot) bool
}

func metricDefinition(
	name historicalcontract.MetricName,
) (metricSpec, bool) {
	switch name {
	case historicalcontract.MetricNameFlightCount:
		return metricSpec{
			unit:        "flights",
			aggregation: historicalcontract.AggregationCount,
			sourceName:  "flights",
			limitReached: func(
				snapshot historicalread.Snapshot,
			) bool {
				return snapshot.FlightLimitReached
			},
		}, true

	case historicalcontract.MetricNameTrajectoryCount:
		return metricSpec{
			unit:        "trajectories",
			aggregation: historicalcontract.AggregationCount,
			sourceName:  "flight_trajectories",
			limitReached: func(
				snapshot historicalread.Snapshot,
			) bool {
				return snapshot.TrajectoryLimitReached
			},
		}, true

	case historicalcontract.MetricNameObservationCount:
		return metricSpec{
			unit:        "observations",
			aggregation: historicalcontract.AggregationCount,
			sourceName:  "flight_states",
			limitReached: func(
				snapshot historicalread.Snapshot,
			) bool {
				return snapshot.ObservationLimitReached
			},
		}, true

	case historicalcontract.MetricNameActiveAircraft:
		return metricSpec{
			unit:        "aircraft",
			aggregation: historicalcontract.AggregationCount,
			sourceName:  "flight_states",
			limitReached: func(
				snapshot historicalread.Snapshot,
			) bool {
				return snapshot.ObservationLimitReached
			},
		}, true

	case historicalcontract.MetricNameTrafficDensity:
		return metricSpec{
			unit: "observations_per_hour",
			aggregation: historicalcontract.
				AggregationAverage,
			sourceName: "flight_states",
			limitReached: func(
				snapshot historicalread.Snapshot,
			) bool {
				return snapshot.ObservationLimitReached
			},
		}, true

	default:
		return metricSpec{}, false
	}
}

func countFlightStarts(
	buckets []historicalwindow.Bucket,
	records []historicalread.FlightRecord,
) []historicalseries.BucketValue {
	result := initializeValues(buckets)

	for _, record := range records {
		if index := bucketIndex(
			buckets,
			record.FirstSeenAt,
		); index >= 0 {
			result[index].Value++
			result[index].SampleCount++
		}
	}

	return result
}

func countTrajectoryStarts(
	buckets []historicalwindow.Bucket,
	records []historicalread.TrajectoryRecord,
) []historicalseries.BucketValue {
	result := initializeValues(buckets)

	for _, record := range records {
		if index := bucketIndex(
			buckets,
			record.StartTime,
		); index >= 0 {
			result[index].Value++
			result[index].SampleCount++
		}
	}

	return result
}

func countObservations(
	buckets []historicalwindow.Bucket,
	records []historicalread.ObservationRecord,
) []historicalseries.BucketValue {
	result := initializeValues(buckets)

	for _, record := range records {
		if index := bucketIndex(
			buckets,
			record.ObservedAt,
		); index >= 0 {
			result[index].Value++
			result[index].SampleCount++
		}
	}

	return result
}

func countActiveAircraft(
	buckets []historicalwindow.Bucket,
	records []historicalread.ObservationRecord,
) ([]historicalseries.BucketValue, int) {
	result := initializeValues(buckets)
	identities := make(
		[]map[string]struct{},
		len(buckets),
	)
	for index := range identities {
		identities[index] = make(map[string]struct{})
	}

	missingIdentityCount := 0
	for _, record := range records {
		index := bucketIndex(
			buckets,
			record.ObservedAt,
		)
		if index < 0 {
			continue
		}

		identity := strings.TrimSpace(
			record.AircraftID,
		)
		if identity == "" {
			identity = strings.ToUpper(
				strings.TrimSpace(record.ICAO24),
			)
		}
		if identity == "" {
			missingIdentityCount++
			continue
		}

		result[index].SampleCount++
		identities[index][identity] = struct{}{}
	}

	for index := range result {
		result[index].Value =
			float64(len(identities[index]))
	}

	return result, missingIdentityCount
}

func observationRate(
	buckets []historicalwindow.Bucket,
	records []historicalread.ObservationRecord,
) []historicalseries.BucketValue {
	result := countObservations(buckets, records)

	for index := range result {
		hours := result[index].Bucket.Duration().Hours()
		if hours > 0 {
			result[index].Value =
				result[index].Value / hours
		}
	}

	return result
}

func initializeValues(
	buckets []historicalwindow.Bucket,
) []historicalseries.BucketValue {
	result := make(
		[]historicalseries.BucketValue,
		len(buckets),
	)
	for index, bucket := range buckets {
		result[index].Bucket = bucket
	}
	return result
}

func bucketIndex(
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

func conservativeCoverage(
	recordCount int,
	limitReached bool,
) float64 {
	if !limitReached {
		return 1
	}
	if recordCount <= 0 {
		return 0.5
	}

	return float64(recordCount) /
		float64(recordCount+1)
}

func latestFlightUpdate(
	records []historicalread.FlightRecord,
	asOfTime time.Time,
) time.Time {
	result := time.Time{}
	for _, record := range records {
		result = boundedLatest(
			result,
			asOfTime,
			record.UpdatedAt,
			record.LastSeenAt,
		)
	}
	return result
}

func latestTrajectoryUpdate(
	records []historicalread.TrajectoryRecord,
	asOfTime time.Time,
) time.Time {
	result := time.Time{}
	for _, record := range records {
		result = boundedLatest(
			result,
			asOfTime,
			record.UpdatedAt,
			record.EndTime,
		)
	}
	return result
}

func latestObservationUpdate(
	records []historicalread.ObservationRecord,
	asOfTime time.Time,
) time.Time {
	result := time.Time{}
	for _, record := range records {
		result = boundedLatest(
			result,
			asOfTime,
			record.CreatedAt,
			record.ObservedAt,
		)
	}
	return result
}

func boundedLatest(
	current time.Time,
	asOfTime time.Time,
	candidates ...time.Time,
) time.Time {
	result := current
	cutoff := asOfTime.UTC()

	for _, candidate := range candidates {
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

	return result
}

func trafficFingerprint(
	request Request,
) string {
	records := []string{
		Version,
		string(request.MetricName),
		request.Plan.Fingerprint,
		request.Plan.AsOfTime.UTC().
			Format(time.RFC3339Nano),
	}

	switch request.MetricName {
	case historicalcontract.MetricNameFlightCount:
		records = append(
			records,
			fmt.Sprintf(
				"flight_limit_reached|%t",
				request.Snapshot.FlightLimitReached,
			),
		)
		for _, record := range request.Snapshot.Flights {
			records = append(
				records,
				fmt.Sprintf(
					"flight|%s|%s|%s|%s",
					record.ID,
					record.AircraftID,
					record.FirstSeenAt.UTC().
						Format(time.RFC3339Nano),
					record.UpdatedAt.UTC().
						Format(time.RFC3339Nano),
				),
			)
		}

	case historicalcontract.MetricNameTrajectoryCount:
		records = append(
			records,
			fmt.Sprintf(
				"trajectory_limit_reached|%t",
				request.Snapshot.TrajectoryLimitReached,
			),
		)
		for _, record := range request.Snapshot.Trajectories {
			records = append(
				records,
				fmt.Sprintf(
					"trajectory|%s|%s|%s|%s",
					record.ID,
					record.AircraftID,
					record.StartTime.UTC().
						Format(time.RFC3339Nano),
					record.UpdatedAt.UTC().
						Format(time.RFC3339Nano),
				),
			)
		}

	case historicalcontract.MetricNameObservationCount,
		historicalcontract.MetricNameActiveAircraft,
		historicalcontract.MetricNameTrafficDensity:
		records = append(
			records,
			fmt.Sprintf(
				"observation_limit_reached|%t",
				request.Snapshot.ObservationLimitReached,
			),
		)
		for _, record := range request.Snapshot.Observations {
			records = append(
				records,
				fmt.Sprintf(
					"observation|%s|%s|%s|%s|%s",
					record.ID,
					record.AircraftID,
					strings.ToUpper(record.ICAO24),
					record.ObservedAt.UTC().
						Format(time.RFC3339Nano),
					record.CreatedAt.UTC().
						Format(time.RFC3339Nano),
				),
			)
		}
	}

	sort.Strings(records)
	sum := sha256.Sum256(
		[]byte(strings.Join(records, "\n")),
	)
	return "sha256:" + hex.EncodeToString(sum[:])
}
