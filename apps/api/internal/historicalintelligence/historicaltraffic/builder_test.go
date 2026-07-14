package historicaltraffic

import (
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func TestBuildTrafficMetrics(
	t *testing.T,
) {
	plan := trafficTestPlan()
	start := plan.Buckets[0].StartTime
	snapshot := historicalread.Snapshot{
		Version: historicalread.Version,
		Flights: []historicalread.FlightRecord{
			{
				ID:          "flight-1",
				AircraftID:  "aircraft-1",
				FirstSeenAt: start.Add(10 * time.Minute),
				LastSeenAt:  start.Add(70 * time.Minute),
				UpdatedAt:   start.Add(70 * time.Minute),
			},
			{
				ID:          "flight-2",
				AircraftID:  "aircraft-2",
				FirstSeenAt: start.Add(80 * time.Minute),
				LastSeenAt:  start.Add(90 * time.Minute),
				UpdatedAt:   start.Add(90 * time.Minute),
			},
		},
		Trajectories: []historicalread.TrajectoryRecord{
			{
				ID:         "trajectory-1",
				AircraftID: "aircraft-1",
				StartTime:  start.Add(5 * time.Minute),
				EndTime:    start.Add(65 * time.Minute),
				UpdatedAt:  start.Add(65 * time.Minute),
			},
		},
		Observations: []historicalread.ObservationRecord{
			{
				ID:         "observation-1",
				AircraftID: "aircraft-1",
				ObservedAt: start.Add(10 * time.Minute),
				CreatedAt:  start.Add(11 * time.Minute),
			},
			{
				ID:         "observation-2",
				AircraftID: "aircraft-1",
				ObservedAt: start.Add(20 * time.Minute),
				CreatedAt:  start.Add(21 * time.Minute),
			},
			{
				ID:         "observation-3",
				ICAO24:     "ABC123",
				ObservedAt: start.Add(70 * time.Minute),
				CreatedAt:  start.Add(71 * time.Minute),
			},
		},
	}

	tests := []struct {
		metric historicalcontract.MetricName
		want   []float64
	}{
		{
			metric: historicalcontract.MetricNameFlightCount,
			want:   []float64{1, 1},
		},
		{
			metric: historicalcontract.MetricNameTrajectoryCount,
			want:   []float64{1, 0},
		},
		{
			metric: historicalcontract.MetricNameObservationCount,
			want:   []float64{2, 1},
		},
		{
			metric: historicalcontract.MetricNameActiveAircraft,
			want:   []float64{1, 1},
		},
		{
			metric: historicalcontract.MetricNameTrafficDensity,
			want:   []float64{2, 1},
		},
	}

	for _, test := range tests {
		t.Run(
			string(test.metric),
			func(t *testing.T) {
				result, err := Build(
					Request{
						Snapshot:    snapshot,
						Plan:        plan,
						MetricName:  test.metric,
						GeneratedAt: plan.AsOfTime,
					},
				)
				if err != nil {
					t.Fatalf(
						"build traffic metric: %v",
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
				if result.Status !=
					historicalcontract.SeriesStatusComplete {
					t.Fatalf(
						"expected complete result, got %s",
						result.Status,
					)
				}
			},
		)
	}
}

func TestBuildTrafficMetricMarksLimitAsPartial(
	t *testing.T,
) {
	plan := trafficTestPlan()
	start := plan.Buckets[0].StartTime
	result, err := Build(
		Request{
			Snapshot: historicalread.Snapshot{
				Version: historicalread.Version,
				Observations: []historicalread.ObservationRecord{
					{
						ID:         "observation-1",
						ObservedAt: start.Add(time.Minute),
						CreatedAt:  start.Add(2 * time.Minute),
					},
				},
				ObservationLimitReached: true,
			},
			Plan: plan,
			MetricName: historicalcontract.
				MetricNameObservationCount,
			GeneratedAt: plan.AsOfTime,
		},
	)
	if err != nil {
		t.Fatalf(
			"build partial traffic metric: %v",
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
	if result.Points[0].CoverageRatio != 0.5 {
		t.Fatalf(
			"expected conservative coverage 0.5, got %f",
			result.Points[0].CoverageRatio,
		)
	}
}

func TestTrafficFingerprintIgnoresRecordOrder(
	t *testing.T,
) {
	plan := trafficTestPlan()
	start := plan.Buckets[0].StartTime
	first := historicalread.ObservationRecord{
		ID:         "a",
		ObservedAt: start.Add(time.Minute),
	}
	second := historicalread.ObservationRecord{
		ID:         "b",
		ObservedAt: start.Add(2 * time.Minute),
	}

	left, err := Build(
		Request{
			Snapshot: historicalread.Snapshot{
				Version: historicalread.Version,
				Observations: []historicalread.ObservationRecord{
					first,
					second,
				},
			},
			Plan: plan,
			MetricName: historicalcontract.
				MetricNameObservationCount,
			GeneratedAt: plan.AsOfTime,
		},
	)
	if err != nil {
		t.Fatalf("build left result: %v", err)
	}
	right, err := Build(
		Request{
			Snapshot: historicalread.Snapshot{
				Version: historicalread.Version,
				Observations: []historicalread.ObservationRecord{
					second,
					first,
				},
			},
			Plan: plan,
			MetricName: historicalcontract.
				MetricNameObservationCount,
			GeneratedAt: plan.AsOfTime,
		},
	)
	if err != nil {
		t.Fatalf("build right result: %v", err)
	}

	if left.Provenance.InputFingerprint !=
		right.Provenance.InputFingerprint {
		t.Fatal(
			"expected order-independent traffic fingerprint",
		)
	}
}

func trafficTestPlan() historicalwindow.Plan {
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
		Fingerprint:        "traffic-plan",
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
