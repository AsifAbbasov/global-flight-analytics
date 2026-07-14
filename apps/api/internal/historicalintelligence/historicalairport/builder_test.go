package historicalairport

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestBuildHistoricalAirportMetrics(
	t *testing.T,
) {
	plan := airportTestPlan()
	start := plan.Buckets[0].StartTime
	first := airportRouteRecord(
		t,
		"route-1",
		"trajectory-1",
		"aircraft-1",
		"UBBB",
		"UGTB",
		start.Add(10*time.Minute),
		start.Add(70*time.Minute),
		plan.AsOfTime,
	)
	second := airportRouteRecord(
		t,
		"route-2",
		"trajectory-2",
		"aircraft-2",
		"UGTB",
		"UBBB",
		start.Add(80*time.Minute),
		start.Add(90*time.Minute),
		plan.AsOfTime,
	)

	tests := []struct {
		metric historicalcontract.MetricName
		want   []float64
	}{
		{
			metric: historicalcontract.
				MetricNameAirportDepartures,
			want: []float64{1, 0},
		},
		{
			metric: historicalcontract.
				MetricNameAirportArrivals,
			want: []float64{0, 1},
		},
		{
			metric: historicalcontract.
				MetricNameAirportOperations,
			want: []float64{1, 1},
		},
		{
			metric: historicalcontract.
				MetricNameUniqueAircraft,
			want: []float64{1, 1},
		},
	}

	for _, test := range tests {
		t.Run(
			string(test.metric),
			func(t *testing.T) {
				result, err := Build(
					Request{
						Snapshot: historicalread.Snapshot{
							Version: historicalread.Version,
							Routes: []historicalread.RouteRecord{
								first,
								second,
							},
						},
						Plan:            plan,
						AirportICAOCode: "ubbb",
						MetricName:      test.metric,
						GeneratedAt:     plan.AsOfTime,
					},
				)
				if err != nil {
					t.Fatalf(
						"build airport metric: %v",
						err,
					)
				}

				got := make(
					[]float64,
					0,
					len(result.Points),
				)
				for _, point := range result.Points {
					got = append(got, point.Value)
				}
				if !reflect.DeepEqual(
					got,
					test.want,
				) {
					t.Fatalf(
						"values = %#v, want %#v",
						got,
						test.want,
					)
				}
				if result.Scope.AirportICAOCode !=
					"UBBB" {
					t.Fatalf(
						"expected normalized airport code, got %q",
						result.Scope.AirportICAOCode,
					)
				}
			},
		)
	}
}

func TestBuildHistoricalAirportMetricMarksInvalidPayloadPartial(
	t *testing.T,
) {
	plan := airportTestPlan()
	result, err := Build(
		Request{
			Snapshot: historicalread.Snapshot{
				Version: historicalread.Version,
				Routes: []historicalread.RouteRecord{
					{
						ID:           "broken",
						TrajectoryID: "trajectory-broken",
						AsOfTime:     plan.AsOfTime,
						RouteJSON:    []byte("{"),
						StoredAt:     plan.AsOfTime,
					},
				},
			},
			Plan:            plan,
			AirportICAOCode: "UBBB",
			MetricName: historicalcontract.
				MetricNameAirportOperations,
			GeneratedAt: plan.AsOfTime,
		},
	)
	if err != nil {
		t.Fatalf(
			"build partial airport metric: %v",
			err,
		)
	}

	if result.Status !=
		historicalcontract.SeriesStatusPartial {
		t.Fatalf(
			"expected partial result, got %s",
			result.Status,
		)
	}
	if result.Points[0].CoverageRatio != 0 {
		t.Fatalf(
			"expected zero decoded coverage, got %f",
			result.Points[0].CoverageRatio,
		)
	}
}

func TestHistoricalAirportFingerprintIgnoresRouteOrder(
	t *testing.T,
) {
	plan := airportTestPlan()
	start := plan.Buckets[0].StartTime
	first := airportRouteRecord(
		t,
		"a",
		"trajectory-a",
		"aircraft-a",
		"UBBB",
		"UGTB",
		start.Add(time.Minute),
		start.Add(30*time.Minute),
		plan.AsOfTime,
	)
	second := airportRouteRecord(
		t,
		"b",
		"trajectory-b",
		"aircraft-b",
		"UGTB",
		"UBBB",
		start.Add(70*time.Minute),
		start.Add(80*time.Minute),
		plan.AsOfTime,
	)

	build := func(
		routes []historicalread.RouteRecord,
	) historicalcontract.Result {
		t.Helper()
		result, err := Build(
			Request{
				Snapshot: historicalread.Snapshot{
					Version: historicalread.Version,
					Routes:  routes,
				},
				Plan:            plan,
				AirportICAOCode: "UBBB",
				MetricName: historicalcontract.
					MetricNameAirportOperations,
				GeneratedAt: plan.AsOfTime,
			},
		)
		if err != nil {
			t.Fatalf("build airport result: %v", err)
		}
		return result
	}

	left := build(
		[]historicalread.RouteRecord{
			first,
			second,
		},
	)
	right := build(
		[]historicalread.RouteRecord{
			second,
			first,
		},
	)

	if left.Provenance.InputFingerprint !=
		right.Provenance.InputFingerprint {
		t.Fatal(
			"expected order-independent airport fingerprint",
		)
	}
}

func airportRouteRecord(
	t *testing.T,
	id string,
	trajectoryID string,
	aircraftID string,
	origin string,
	destination string,
	startTime time.Time,
	endTime time.Time,
	asOfTime time.Time,
) historicalread.RouteRecord {
	t.Helper()

	payload, err := json.Marshal(
		routecontract.Result{
			SchemaVersion: routecontract.SchemaVersionV1,
			Status:        routecontract.RouteStatusComplete,
			TrajectoryID:  trajectoryID,
			AircraftID:    aircraftID,
			ICAO24:        "ABC123",
			Window: routecontract.RouteWindow{
				StartTime: startTime,
				EndTime:   endTime,
				AsOfTime:  asOfTime,
			},
			Origin: &routecontract.EndpointInference{
				Role: routecontract.EndpointRoleOrigin,
				Airport: routecontract.AirportReference{
					ICAOCode: origin,
				},
			},
			Destination: &routecontract.EndpointInference{
				Role: routecontract.EndpointRoleDestination,
				Airport: routecontract.AirportReference{
					ICAOCode: destination,
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("marshal route result: %v", err)
	}

	return historicalread.RouteRecord{
		ID:               id,
		TrajectoryID:     trajectoryID,
		AsOfTime:         asOfTime,
		InputFingerprint: "sha256:test",
		Status:           "complete",
		RouteJSON:        payload,
		StoredAt:         asOfTime,
	}
}

func airportTestPlan() historicalwindow.Plan {
	start := time.Date(
		2026,
		time.July,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	end := start.Add(2 * time.Hour)
	window := historicalcontract.TimeWindow{
		StartTime: start,
		EndTime:   end,
		AsOfTime:  end,
	}

	return historicalwindow.Plan{
		Version:            historicalwindow.Version,
		Fingerprint:        "airport-plan",
		RequestedStartTime: start,
		RequestedEndTime:   end,
		AsOfTime:           end,
		Granularity: historicalcontract.
			GranularityHour,
		EffectiveWindow: &window,
		Buckets: []historicalwindow.Bucket{
			{
				Key:       "bucket-0",
				Sequence:  0,
				StartTime: start,
				EndTime:   start.Add(time.Hour),
			},
			{
				Key:       "bucket-1",
				Sequence:  1,
				StartTime: start.Add(time.Hour),
				EndTime:   end,
			},
		},
		MaximumBucketCount: 100,
	}
}
