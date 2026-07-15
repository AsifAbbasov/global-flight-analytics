package historicalroute

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

func TestBuildHistoricalRouteMetrics(
	t *testing.T,
) {
	plan := routeTestPlan()
	start := plan.Buckets[0].StartTime

	complete := historicalRouteRecord(
		t,
		"route-1",
		"trajectory-1",
		routecontract.RouteStatusComplete,
		"UBBB",
		"UGTB",
		0.9,
		450,
		start.Add(10*time.Minute),
		start.Add(50*time.Minute),
		plan.AsOfTime,
	)
	partial := historicalRouteRecord(
		t,
		"route-2",
		"trajectory-2",
		routecontract.RouteStatusPartial,
		"UBBB",
		"UGTB",
		0.5,
		450,
		start.Add(70*time.Minute),
		start.Add(110*time.Minute),
		plan.AsOfTime,
	)
	unavailable := historicalRouteRecord(
		t,
		"route-3",
		"trajectory-3",
		routecontract.RouteStatusUnavailable,
		"",
		"",
		0,
		0,
		start.Add(80*time.Minute),
		start.Add(100*time.Minute),
		plan.AsOfTime,
	)

	snapshot := historicalread.Snapshot{
		Version: historicalread.Version,
		Routes: []historicalread.RouteRecord{
			complete,
			partial,
			unavailable,
		},
	}

	tests := []struct {
		metric historicalcontract.MetricName
		want   []float64
	}{
		{
			metric: historicalcontract.MetricNameActiveRoutes,
			want:   []float64{1, 1},
		},
		{
			metric: historicalcontract.MetricNameRouteObservations,
			want:   []float64{1, 2},
		},
		{
			metric: historicalcontract.MetricNameRouteConfidence,
			want:   []float64{0.9, 0.25},
		},
		{
			metric: historicalcontract.MetricNameCompleteRouteRatio,
			want:   []float64{1, 0},
		},
		{
			metric: historicalcontract.MetricNamePartialRouteRatio,
			want:   []float64{0, 0.5},
		},
		{
			metric: historicalcontract.MetricNameUnavailableRouteRatio,
			want:   []float64{0, 0.5},
		},
		{
			metric: historicalcontract.MetricNameGreatCircleDistanceKM,
			want:   []float64{450, 450},
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
						"build historical route metric: %v",
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
			},
		)
	}
}

func TestBuildHistoricalRouteMetricFiltersRouteScope(
	t *testing.T,
) {
	plan := routeTestPlan()
	start := plan.Buckets[0].StartTime
	snapshot := historicalread.Snapshot{
		Version: historicalread.Version,
		Routes: []historicalread.RouteRecord{
			historicalRouteRecord(
				t,
				"route-1",
				"trajectory-1",
				routecontract.RouteStatusComplete,
				"UBBB",
				"UGTB",
				0.9,
				450,
				start.Add(time.Minute),
				start.Add(30*time.Minute),
				plan.AsOfTime,
			),
			historicalRouteRecord(
				t,
				"route-2",
				"trajectory-2",
				routecontract.RouteStatusComplete,
				"UGTB",
				"UBBB",
				0.8,
				450,
				start.Add(10*time.Minute),
				start.Add(40*time.Minute),
				plan.AsOfTime,
			),
		},
	}

	result, err := Build(
		Request{
			Snapshot:            snapshot,
			Plan:                plan,
			OriginICAOCode:      "ubbb",
			DestinationICAOCode: "ugtb",
			MetricName: historicalcontract.
				MetricNameRouteObservations,
			GeneratedAt: plan.AsOfTime,
		},
	)
	if err != nil {
		t.Fatalf(
			"build route-scoped historical metric: %v",
			err,
		)
	}

	if result.Scope.Type !=
		historicalcontract.ScopeTypeRoute ||
		result.Scope.OriginICAOCode != "UBBB" ||
		result.Scope.DestinationICAOCode != "UGTB" {
		t.Fatalf(
			"unexpected route scope: %#v",
			result.Scope,
		)
	}
	if result.Points[0].Value != 1 {
		t.Fatalf(
			"expected one matching route observation, got %f",
			result.Points[0].Value,
		)
	}
}

func TestBuildHistoricalRouteMetricUsesLatestRecordPerTrajectory(
	t *testing.T,
) {
	plan := routeTestPlan()
	start := plan.Buckets[0].StartTime
	older := historicalRouteRecord(
		t,
		"older",
		"trajectory-1",
		routecontract.RouteStatusPartial,
		"UBBB",
		"UGTB",
		0.5,
		450,
		start.Add(time.Minute),
		start.Add(30*time.Minute),
		plan.AsOfTime.Add(-time.Minute),
	)
	newer := historicalRouteRecord(
		t,
		"newer",
		"trajectory-1",
		routecontract.RouteStatusComplete,
		"UBBB",
		"UGTB",
		0.9,
		450,
		start.Add(time.Minute),
		start.Add(30*time.Minute),
		plan.AsOfTime,
	)

	result, err := Build(
		Request{
			Snapshot: historicalread.Snapshot{
				Version: historicalread.Version,
				Routes: []historicalread.RouteRecord{
					older,
					newer,
				},
			},
			Plan: plan,
			MetricName: historicalcontract.
				MetricNameCompleteRouteRatio,
			GeneratedAt: plan.AsOfTime,
		},
	)
	if err != nil {
		t.Fatalf("build latest route metric: %v", err)
	}

	if result.Points[0].Value != 1 ||
		result.Points[0].SampleCount != 1 {
		t.Fatalf(
			"expected one latest complete route, got %#v",
			result.Points[0],
		)
	}
}

func historicalRouteRecord(
	t *testing.T,
	id string,
	trajectoryID string,
	status routecontract.RouteStatus,
	origin string,
	destination string,
	confidence float64,
	distance float64,
	startTime time.Time,
	endTime time.Time,
	asOfTime time.Time,
) historicalread.RouteRecord {
	t.Helper()

	var originEndpoint *routecontract.EndpointInference
	var destinationEndpoint *routecontract.EndpointInference
	if origin != "" {
		originEndpoint = &routecontract.EndpointInference{
			Role: routecontract.EndpointRoleOrigin,
			Airport: routecontract.AirportReference{
				ICAOCode: origin,
			},
		}
	}
	if destination != "" {
		destinationEndpoint =
			&routecontract.EndpointInference{
				Role: routecontract.EndpointRoleDestination,
				Airport: routecontract.AirportReference{
					ICAOCode: destination,
				},
			}
	}

	payload, err := json.Marshal(
		routecontract.Result{
			SchemaVersion: routecontract.SchemaVersionV1,
			Status:        status,
			TrajectoryID:  trajectoryID,
			Window: routecontract.RouteWindow{
				StartTime: startTime,
				EndTime:   endTime,
				AsOfTime:  asOfTime,
			},
			Origin:      originEndpoint,
			Destination: destinationEndpoint,
			Summary: routecontract.RouteSummary{
				GreatCircleDistanceKM: distance,
			},
			Confidence: routecontract.Confidence{
				Score: confidence,
				Level: routecontract.
					ConfidenceLevelForScore(confidence),
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
		Status:           string(status),
		RouteJSON:        payload,
		StoredAt:         asOfTime,
	}
}

func routeTestPlan() historicalwindow.Plan {
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
		Fingerprint:        "route-plan",
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
