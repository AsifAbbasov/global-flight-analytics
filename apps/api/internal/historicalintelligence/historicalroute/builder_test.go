package historicalroute

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestBuildHistoricalRouteMetricsFromValidatedEvidence(t *testing.T) {
	plan := routeTestPlan()
	start := plan.Buckets[0].StartTime
	complete := historicalRouteRecord(
		t,
		"route-complete",
		routeTestResult(
			routecontract.RouteStatusComplete,
			"trajectory-complete",
			"UBBB",
			"UGTB",
			0.9,
			999,
			start.Add(10*time.Minute),
			start.Add(50*time.Minute),
			plan.AsOfTime,
			[]string{"opensky"},
		),
	)
	partial := historicalRouteRecord(
		t,
		"route-partial",
		routeTestResult(
			routecontract.RouteStatusPartial,
			"trajectory-partial",
			"UBBB",
			"",
			0.5,
			0,
			start.Add(70*time.Minute),
			start.Add(110*time.Minute),
			plan.AsOfTime,
			[]string{"opensky"},
		),
	)
	unavailable := historicalRouteRecord(
		t,
		"route-unavailable",
		routeTestResult(
			routecontract.RouteStatusUnavailable,
			"trajectory-unavailable",
			"",
			"",
			0,
			0,
			start.Add(80*time.Minute),
			start.Add(100*time.Minute),
			plan.AsOfTime,
			[]string{"opensky"},
		),
	)
	snapshot := historicalread.Snapshot{
		Version: historicalread.Version,
		Routes:  []historicalread.RouteRecord{complete, partial, unavailable},
	}

	distance := greatCircleDistanceKM(40.4675, 50.0467, 41.6692, 44.9547)
	tests := []struct {
		metric historicalcontract.MetricName
		want   []float64
	}{
		{historicalcontract.MetricNameActiveRoutes, []float64{1, 0}},
		{historicalcontract.MetricNameRouteObservations, []float64{1, 2}},
		{historicalcontract.MetricNameRouteConfidence, []float64{0.9, 0.25}},
		{historicalcontract.MetricNameCompleteRouteRatio, []float64{1, 0}},
		{historicalcontract.MetricNamePartialRouteRatio, []float64{0, 0.5}},
		{historicalcontract.MetricNameUnavailableRouteRatio, []float64{0, 0.5}},
		{historicalcontract.MetricNameGreatCircleDistanceKM, []float64{distance, 0}},
	}

	for _, test := range tests {
		t.Run(string(test.metric), func(t *testing.T) {
			result, err := Build(Request{
				Snapshot:    snapshot,
				Plan:        plan,
				MetricName:  test.metric,
				GeneratedAt: plan.AsOfTime,
			})
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}
			if len(result.Points) != len(test.want) {
				t.Fatalf("point count = %d, want %d", len(result.Points), len(test.want))
			}
			for index, want := range test.want {
				if math.Abs(result.Points[index].Value-want) > 1e-9 {
					t.Fatalf("point %d value = %.12f, want %.12f", index, result.Points[index].Value, want)
				}
			}
			if test.metric == historicalcontract.MetricNameActiveRoutes &&
				result.Metric.Unit != "route_pairs" {
				t.Fatalf("active routes unit = %q", result.Metric.Unit)
			}
			if !reflect.DeepEqual(
				result.Provenance.SourceNames,
				[]string{"flight_route_results", "opensky"},
			) {
				t.Fatalf("source names = %#v", result.Provenance.SourceNames)
			}
		})
	}
}

func TestBuildRejectsRoutePairScopeForStatusRatios(t *testing.T) {
	plan := routeTestPlan()
	record := historicalRouteRecord(
		t,
		"route-scope",
		routeTestResult(
			routecontract.RouteStatusComplete,
			"trajectory-scope",
			"UBBB",
			"UGTB",
			0.9,
			450,
			plan.Buckets[0].StartTime,
			plan.Buckets[0].EndTime.Add(-time.Minute),
			plan.AsOfTime,
			[]string{"opensky"},
		),
	)
	for _, metricName := range []historicalcontract.MetricName{
		historicalcontract.MetricNameCompleteRouteRatio,
		historicalcontract.MetricNamePartialRouteRatio,
		historicalcontract.MetricNameUnavailableRouteRatio,
	} {
		_, err := Build(Request{
			Snapshot: historicalread.Snapshot{
				Version: historicalread.Version,
				Routes:  []historicalread.RouteRecord{record},
			},
			Plan:                plan,
			OriginICAOCode:      "ubbb",
			DestinationICAOCode: "ugtb",
			MetricName:          metricName,
			GeneratedAt:         plan.AsOfTime,
		})
		if !errors.Is(err, ErrMetricScopeUnsupported) {
			t.Fatalf("metric %q error = %v, want ErrMetricScopeUnsupported", metricName, err)
		}
	}
}

func TestBuildRoutePairScopeUsesOnlyMatchingCompleteEvidence(t *testing.T) {
	plan := routeTestPlan()
	start := plan.Buckets[0].StartTime
	matching := historicalRouteRecord(t, "matching", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-matching",
		"UBBB",
		"UGTB",
		0.9,
		450,
		start.Add(time.Minute),
		start.Add(30*time.Minute),
		plan.AsOfTime.Add(-10*time.Minute),
		[]string{"opensky-a"},
	))
	matching.StoredAt = plan.AsOfTime.Add(-5 * time.Minute)
	reverse := historicalRouteRecord(t, "reverse", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-reverse",
		"UGTB",
		"UBBB",
		0.8,
		450,
		start.Add(2*time.Minute),
		start.Add(31*time.Minute),
		plan.AsOfTime.Add(-10*time.Minute),
		[]string{"opensky-b"},
	))
	reverse.StoredAt = plan.AsOfTime

	result, err := Build(Request{
		Snapshot: historicalread.Snapshot{
			Version: historicalread.Version,
			Routes:  []historicalread.RouteRecord{matching, reverse},
		},
		Plan:                plan,
		OriginICAOCode:      "ubbb",
		DestinationICAOCode: "ugtb",
		MetricName:          historicalcontract.MetricNameRouteObservations,
		GeneratedAt:         plan.AsOfTime,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Points[0].Value != 1 || result.Points[0].SampleCount != 1 {
		t.Fatalf("matching point = %#v", result.Points[0])
	}
	if !reflect.DeepEqual(
		result.Provenance.SourceNames,
		[]string{"flight_route_results", "opensky-a"},
	) {
		t.Fatalf("source names = %#v", result.Provenance.SourceNames)
	}
	if !result.Provenance.LatestSourceUpdatedAt.Equal(matching.StoredAt) {
		t.Fatalf(
			"latest source update = %s, want %s",
			result.Provenance.LatestSourceUpdatedAt,
			matching.StoredAt,
		)
	}
}

func TestBuildRejectsInvalidContract(t *testing.T) {
	plan := routeTestPlan()
	result := routeTestResult(
		routecontract.RouteStatusPartial,
		"trajectory-invalid-contract",
		"UBBB",
		"",
		0.5,
		0,
		plan.Buckets[0].StartTime,
		plan.Buckets[0].EndTime.Add(-time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	)
	result.Destination = routeTestEndpoint(
		routecontract.EndpointRoleDestination,
		"UGTB",
		41.6692,
		44.9547,
		result.Window.EndTime,
		0.5,
		"opensky",
	)
	result.Confidence.EvidenceCount = 2
	record := historicalRouteRecordUnchecked(t, "invalid-contract", result)

	_, err := Build(Request{
		Snapshot: historicalread.Snapshot{
			Version: historicalread.Version,
			Routes:  []historicalread.RouteRecord{record},
		},
		Plan:        plan,
		MetricName:  historicalcontract.MetricNameRouteObservations,
		GeneratedAt: plan.AsOfTime,
	})
	if !errors.Is(err, historicalread.ErrRouteContractInvalid) {
		t.Fatalf("Build() error = %v, want ErrRouteContractInvalid", err)
	}
}

func TestBuildRejectsInvalidJSON(t *testing.T) {
	plan := routeTestPlan()
	record := historicalRouteRecord(t, "invalid-json", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-invalid-json",
		"UBBB",
		"UGTB",
		0.9,
		450,
		plan.Buckets[0].StartTime,
		plan.Buckets[0].EndTime.Add(-time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	))
	record.Result = routecontract.Result{}
	record.ResultAvailable = false
	record.RouteJSON = []byte("{")

	_, err := Build(Request{
		Snapshot: historicalread.Snapshot{
			Version: historicalread.Version,
			Routes:  []historicalread.RouteRecord{record},
		},
		Plan:        plan,
		MetricName:  historicalcontract.MetricNameRouteObservations,
		GeneratedAt: plan.AsOfTime,
	})
	if !errors.Is(err, historicalread.ErrRoutePayloadDecode) {
		t.Fatalf("Build() error = %v, want ErrRoutePayloadDecode", err)
	}
}

func TestBuildRejectsPersistenceMetadataMismatch(t *testing.T) {
	plan := routeTestPlan()
	record := historicalRouteRecord(t, "metadata", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-metadata",
		"UBBB",
		"UGTB",
		0.9,
		450,
		plan.Buckets[0].StartTime,
		plan.Buckets[0].EndTime.Add(-time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	))
	record.Status = string(routecontract.RouteStatusPartial)

	_, err := Build(Request{
		Snapshot: historicalread.Snapshot{
			Version: historicalread.Version,
			Routes:  []historicalread.RouteRecord{record},
		},
		Plan:        plan,
		MetricName:  historicalcontract.MetricNameRouteObservations,
		GeneratedAt: plan.AsOfTime,
	})
	if !errors.Is(err, historicalread.ErrRouteMetadataMismatch) {
		t.Fatalf("Build() error = %v, want ErrRouteMetadataMismatch", err)
	}
}

func TestBuildValidatesSnapshotWindowContainment(t *testing.T) {
	plan := routeTestPlan()
	record := historicalRouteRecord(t, "window", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-window",
		"UBBB",
		"UGTB",
		0.9,
		450,
		plan.Buckets[0].StartTime,
		plan.Buckets[0].EndTime.Add(-time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	))
	containingWindow := historicalcontract.TimeWindow{
		StartTime: plan.RequestedStartTime.Add(-2 * time.Hour),
		EndTime:   plan.RequestedEndTime,
		AsOfTime:  plan.AsOfTime,
	}
	request := Request{
		Snapshot: historicalread.Snapshot{
			Version:        historicalread.Version,
			IsolationLevel: historicalread.SnapshotIsolationRepeatableRead,
			Query:          historicalread.Query{Window: containingWindow},
			Routes:         []historicalread.RouteRecord{record},
		},
		Plan:        plan,
		MetricName:  historicalcontract.MetricNameRouteObservations,
		GeneratedAt: plan.AsOfTime,
	}
	if _, err := Build(request); err != nil {
		t.Fatalf("containing snapshot window rejected: %v", err)
	}

	request.Snapshot.Query.Window.EndTime = plan.Buckets[0].EndTime
	_, err := Build(request)
	if !errors.Is(err, ErrSnapshotWindowMismatch) {
		t.Fatalf("Build() error = %v, want ErrSnapshotWindowMismatch", err)
	}
}

func TestBuildRejectsIncompleteRoutePairCoverage(t *testing.T) {
	plan := routeTestPlan()
	record := historicalRouteRecord(t, "limited", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-limited",
		"UBBB",
		"UGTB",
		0.9,
		450,
		plan.Buckets[0].StartTime,
		plan.Buckets[0].EndTime.Add(-time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	))
	snapshot := historicalread.Snapshot{
		Version:           historicalread.Version,
		Routes:            []historicalread.RouteRecord{record},
		RouteMatchedCount: 2,
		RouteLimitReached: true,
	}

	_, err := Build(Request{
		Snapshot:            snapshot,
		Plan:                plan,
		OriginICAOCode:      "UBBB",
		DestinationICAOCode: "UGTB",
		MetricName:          historicalcontract.MetricNameRouteObservations,
		GeneratedAt:         plan.AsOfTime,
	})
	if !errors.Is(err, ErrRouteScopeCoverageUnavailable) {
		t.Fatalf("route-pair error = %v, want ErrRouteScopeCoverageUnavailable", err)
	}

	snapshot.RouteMatchedCount = 0
	_, err = Build(Request{
		Snapshot:    snapshot,
		Plan:        plan,
		MetricName:  historicalcontract.MetricNameRouteObservations,
		GeneratedAt: plan.AsOfTime,
	})
	if !errors.Is(err, ErrRouteMatchedCountRequired) {
		t.Fatalf("global error = %v, want ErrRouteMatchedCountRequired", err)
	}

	snapshot.RouteMatchedCount = int64(len(snapshot.Routes))
	_, err = Build(Request{
		Snapshot:    snapshot,
		Plan:        plan,
		MetricName:  historicalcontract.MetricNameRouteObservations,
		GeneratedAt: plan.AsOfTime,
	})
	if !errors.Is(err, ErrRouteMatchedCountInvalid) {
		t.Fatalf("global error = %v, want ErrRouteMatchedCountInvalid", err)
	}
}

func TestBuildUsesExactIncompleteGlobalCoverage(t *testing.T) {
	plan := routeTestPlan()
	record := historicalRouteRecord(t, "partial-coverage", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-partial-coverage",
		"UBBB",
		"UGTB",
		0.9,
		450,
		plan.Buckets[0].StartTime,
		plan.Buckets[0].EndTime.Add(-time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	))
	result, err := Build(Request{
		Snapshot: historicalread.Snapshot{
			Version:           historicalread.Version,
			Routes:            []historicalread.RouteRecord{record},
			RouteMatchedCount: 3,
			RouteLimitReached: true,
		},
		Plan:        plan,
		MetricName:  historicalcontract.MetricNameRouteObservations,
		GeneratedAt: plan.AsOfTime,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status != historicalcontract.SeriesStatusPartial ||
		len(result.Points) != 2 ||
		math.Abs(result.Points[0].CoverageRatio-(1.0/3.0)) > 1e-12 ||
		result.Points[1].Status != historicalcontract.BucketStatusUnavailable {
		t.Fatalf("unexpected partial coverage result: %#v", result)
	}
}

func TestBuildFailsClosedWithoutScopedSourceEvidence(t *testing.T) {
	plan := routeTestPlan()
	record := historicalRouteRecord(t, "other-pair", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-other-pair",
		"UGTB",
		"UBBB",
		0.9,
		450,
		plan.Buckets[0].StartTime,
		plan.Buckets[0].EndTime.Add(-time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	))
	_, err := Build(Request{
		Snapshot: historicalread.Snapshot{
			Version: historicalread.Version,
			Routes:  []historicalread.RouteRecord{record},
		},
		Plan:                plan,
		OriginICAOCode:      "UBBB",
		DestinationICAOCode: "UGTB",
		MetricName:          historicalcontract.MetricNameRouteObservations,
		GeneratedAt:         plan.AsOfTime,
	})
	if !errors.Is(err, ErrRouteSourceEvidenceUnavailable) {
		t.Fatalf("Build() error = %v, want ErrRouteSourceEvidenceUnavailable", err)
	}
}

func TestRouteFingerprintBindsStoredAtAndLimitsOnlyWhenRelevant(t *testing.T) {
	plan := routeTestPlan()
	record := historicalRouteRecord(t, "fingerprint", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-fingerprint",
		"UBBB",
		"UGTB",
		0.9,
		450,
		plan.Buckets[0].StartTime,
		plan.Buckets[0].EndTime.Add(-time.Minute),
		plan.AsOfTime.Add(-10*time.Minute),
		[]string{"opensky"},
	))
	items, err := decodeRouteEvidenceSet(
		[]historicalread.RouteRecord{record},
		plan.AsOfTime,
	)
	if err != nil {
		t.Fatalf("decodeRouteEvidenceSet() error = %v", err)
	}
	request := Request{
		Snapshot:   historicalread.Snapshot{Version: historicalread.Version},
		Plan:       plan,
		MetricName: historicalcontract.MetricNameRouteObservations,
	}
	base := routeFingerprint(request, items, "", "")

	changedRecord := record
	changedRecord.StoredAt = changedRecord.StoredAt.Add(time.Second)
	changedItems, decodeErr := decodeRouteEvidenceSet(
		[]historicalread.RouteRecord{changedRecord},
		plan.AsOfTime,
	)
	if decodeErr != nil {
		t.Fatalf("decode changed record: %v", decodeErr)
	}
	if routeFingerprint(request, changedItems, "", "") == base {
		t.Fatal("fingerprint did not change with StoredAt")
	}

	limited := request
	limited.Snapshot.RouteLimitReached = true
	limited.Snapshot.RouteMatchedCount = 2
	if routeFingerprint(limited, items, "", "") == base {
		t.Fatal("fingerprint did not change with RouteLimitReached")
	}

	byteLimited := request
	byteLimited.Snapshot.RouteByteLimitReached = true
	byteLimited.Snapshot.RouteMatchedCount = 2
	byteLimited.Snapshot.RoutePayloadBytes = 100
	byteLimited.Snapshot.RouteTotalPayloadBytes = 200
	if routeFingerprint(byteLimited, items, "", "") == base {
		t.Fatal("fingerprint did not change with RouteByteLimitReached")
	}

	changedResult := append([]routeEvidence(nil), items...)
	changedResult[0].result.Confidence.Score = 0.85
	if routeFingerprint(request, changedResult, "", "") == base {
		t.Fatal("fingerprint did not change with validated payload semantics")
	}

	matchedOnly := request
	matchedOnly.Snapshot.RouteMatchedCount = 999
	matchedOnly.Snapshot.RoutePayloadBytes = 999
	if routeFingerprint(matchedOnly, items, "", "") != base {
		t.Fatal("complete-read fingerprint changed with irrelevant global counters")
	}
}

func TestLatestRouteRecordsRejectsMissingIdentityAndUsesStoredAtTieBreak(t *testing.T) {
	plan := routeTestPlan()
	missing := historicalread.RouteRecord{
		ID:       "missing",
		AsOfTime: plan.AsOfTime,
	}
	if _, err := latestRouteRecords(
		[]historicalread.RouteRecord{missing},
		plan.AsOfTime,
	); !errors.Is(err, ErrRouteRecordIdentityRequired) {
		t.Fatalf("latestRouteRecords() error = %v", err)
	}

	older := historicalRouteRecord(t, "older", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-tie",
		"UBBB",
		"UGTB",
		0.8,
		450,
		plan.Buckets[0].StartTime,
		plan.Buckets[0].EndTime.Add(-time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	))
	older.StoredAt = plan.AsOfTime.Add(-time.Minute)
	newer := older
	newer.ID = "newer"
	newer.StoredAt = plan.AsOfTime
	selected, err := latestRouteRecords(
		[]historicalread.RouteRecord{older, newer},
		plan.AsOfTime,
	)
	if err != nil {
		t.Fatalf("latestRouteRecords() error = %v", err)
	}
	if len(selected) != 1 || selected[0].ID != "newer" {
		t.Fatalf("selected records = %#v", selected)
	}
}

func TestBucketBoundariesAndZeroDenominatorRemainExplicit(t *testing.T) {
	plan := routeTestPlan()
	record := historicalRouteRecord(t, "boundary", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-boundary",
		"UBBB",
		"UGTB",
		0.9,
		450,
		plan.Buckets[0].EndTime,
		plan.Buckets[1].EndTime.Add(-time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	))
	result, err := Build(Request{
		Snapshot: historicalread.Snapshot{
			Version: historicalread.Version,
			Routes:  []historicalread.RouteRecord{record},
		},
		Plan:        plan,
		MetricName:  historicalcontract.MetricNamePartialRouteRatio,
		GeneratedAt: plan.AsOfTime,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Points[0].Value != 0 ||
		result.Points[0].SampleCount != 0 ||
		result.Points[0].Status != historicalcontract.BucketStatusComplete {
		t.Fatalf("zero-denominator bucket = %#v", result.Points[0])
	}
	if result.Points[1].SampleCount != 1 {
		t.Fatalf("boundary bucket = %#v", result.Points[1])
	}
}

func TestActiveRoutePairsSampleCountTracksSupportingResults(t *testing.T) {
	plan := routeTestPlan()
	start := plan.Buckets[0].StartTime
	first := historicalRouteRecord(t, "pair-one", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-pair-one",
		"UBBB",
		"UGTB",
		0.9,
		450,
		start.Add(time.Minute),
		start.Add(20*time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	))
	second := historicalRouteRecord(t, "pair-two", routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-pair-two",
		"UBBB",
		"UGTB",
		0.8,
		450,
		start.Add(2*time.Minute),
		start.Add(21*time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	))
	result, err := Build(Request{
		Snapshot: historicalread.Snapshot{
			Version: historicalread.Version,
			Routes:  []historicalread.RouteRecord{first, second},
		},
		Plan:        plan,
		MetricName:  historicalcontract.MetricNameActiveRoutes,
		GeneratedAt: plan.AsOfTime,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Points[0].Value != 1 || result.Points[0].SampleCount != 2 {
		t.Fatalf("active route pair point = %#v", result.Points[0])
	}
}

func TestSameAirportDistanceIsRecomputedAsZero(t *testing.T) {
	plan := routeTestPlan()
	resultContract := routeTestResult(
		routecontract.RouteStatusComplete,
		"trajectory-same-airport",
		"UBBB",
		"UBBB",
		0.9,
		999,
		plan.Buckets[0].StartTime,
		plan.Buckets[0].EndTime.Add(-time.Minute),
		plan.AsOfTime,
		[]string{"opensky"},
	)
	resultContract.Destination.Airport.Latitude =
		resultContract.Origin.Airport.Latitude
	resultContract.Destination.Airport.Longitude =
		resultContract.Origin.Airport.Longitude
	resultContract.Summary.SameAirport = true
	record := historicalRouteRecord(t, "same-airport", resultContract)
	result, err := Build(Request{
		Snapshot: historicalread.Snapshot{
			Version: historicalread.Version,
			Routes:  []historicalread.RouteRecord{record},
		},
		Plan:        plan,
		MetricName:  historicalcontract.MetricNameGreatCircleDistanceKM,
		GeneratedAt: plan.AsOfTime,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Points[0].Value != 0 {
		t.Fatalf("same-airport distance = %f", result.Points[0].Value)
	}
}

func historicalRouteRecord(
	t *testing.T,
	id string,
	result routecontract.Result,
) historicalread.RouteRecord {
	t.Helper()
	report := routecontract.Validate(result)
	if report.Status != routecontract.ValidationStatusValid {
		t.Fatalf("test route contract is invalid: %#v", report)
	}
	return historicalRouteRecordWithReport(t, id, result, report.WarningCount)
}

func historicalRouteRecordUnchecked(
	t *testing.T,
	id string,
	result routecontract.Result,
) historicalread.RouteRecord {
	t.Helper()
	report := routecontract.Validate(result)
	return historicalRouteRecordWithReport(t, id, result, report.WarningCount)
}

func historicalRouteRecordWithReport(
	t *testing.T,
	id string,
	result routecontract.Result,
	warningCount int,
) historicalread.RouteRecord {
	t.Helper()
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal route result: %v", err)
	}
	sum := sha256.Sum256(payload)
	return historicalread.RouteRecord{
		ID:                     id,
		TrajectoryID:           result.TrajectoryID,
		EventStartTime:         result.Window.StartTime,
		EventEndTime:           result.Window.EndTime,
		AsOfTime:               result.Window.AsOfTime,
		InputFingerprint:       result.Provenance.InputFingerprint,
		Status:                 string(result.Status),
		ConfidenceLevel:        string(result.Confidence.Level),
		ValidationWarningCount: warningCount,
		StoredAt:               result.GeneratedAt,
		PayloadBytes:           int64(len(payload)),
		PayloadFingerprint:     "sha256:" + hex.EncodeToString(sum[:]),
		Result:                 result.Clone(),
		ResultAvailable:        true,
	}
}

func routeTestResult(
	status routecontract.RouteStatus,
	trajectoryID string,
	origin string,
	destination string,
	confidence float64,
	persistedDistance float64,
	startTime time.Time,
	endTime time.Time,
	asOfTime time.Time,
	sourceNames []string,
) routecontract.Result {
	fingerprint := "sha256:" + strings.Repeat("a", 64)
	sourceName := "route-source"
	if len(sourceNames) > 0 {
		sourceName = sourceNames[0]
	}
	result := routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        status,
		TrajectoryID:  trajectoryID,
		ICAO24:        "ABC123",
		Callsign:      "J2001",
		Window: routecontract.RouteWindow{
			StartTime: startTime.UTC(),
			EndTime:   endTime.UTC(),
			AsOfTime:  asOfTime.UTC(),
		},
		Summary: routecontract.RouteSummary{
			GreatCircleDistanceKM: persistedDistance,
			SameAirport:           origin != "" && origin == destination,
		},
		Provenance: routecontract.Provenance{
			ResolverVersion:     "route-resolver-test-v1",
			InputFingerprint:    fingerprint,
			TrajectoryUpdatedAt: endTime.UTC(),
			SourceNames:         append([]string(nil), sourceNames...),
		},
		GeneratedAt: asOfTime.UTC(),
		Limitations: []routecontract.Limitation{
			{
				Code:    "probable_route_only",
				Message: "Route evidence is inferred rather than filed flight-plan data.",
				Scope:   "route",
			},
		},
	}

	switch status {
	case routecontract.RouteStatusComplete:
		result.Origin = routeTestEndpoint(
			routecontract.EndpointRoleOrigin,
			origin,
			40.4675,
			50.0467,
			startTime,
			confidence,
			sourceName,
		)
		result.Destination = routeTestEndpoint(
			routecontract.EndpointRoleDestination,
			destination,
			41.6692,
			44.9547,
			endTime,
			confidence,
			sourceName,
		)
		result.Confidence = routeTestConfidence(
			confidence,
			2,
			"route_complete",
		)
	case routecontract.RouteStatusPartial:
		if origin != "" {
			result.Origin = routeTestEndpoint(
				routecontract.EndpointRoleOrigin,
				origin,
				40.4675,
				50.0467,
				startTime,
				confidence,
				sourceName,
			)
		} else if destination != "" {
			result.Destination = routeTestEndpoint(
				routecontract.EndpointRoleDestination,
				destination,
				41.6692,
				44.9547,
				endTime,
				confidence,
				sourceName,
			)
		}
		result.Confidence = routeTestConfidence(
			confidence,
			1,
			"route_partial",
		)
	case routecontract.RouteStatusUnavailable:
		result.Confidence = routecontract.Confidence{
			Level: routecontract.ConfidenceLevelNone,
			Reasons: []routecontract.ConfidenceReason{
				{
					Code:    "no_endpoint_evidence",
					Message: "No route endpoint evidence is available.",
				},
			},
		}
	}
	return result
}

func routeTestEndpoint(
	role routecontract.EndpointRole,
	icaoCode string,
	latitude float64,
	longitude float64,
	observedAt time.Time,
	score float64,
	sourceName string,
) *routecontract.EndpointInference {
	return &routecontract.EndpointInference{
		Role: role,
		Airport: routecontract.AirportReference{
			ICAOCode:           icaoCode,
			Name:               "Test Airport " + icaoCode,
			Latitude:           latitude,
			Longitude:          longitude,
			ElevationM:         10,
			ElevationAvailable: true,
		},
		DistanceKM: 2,
		Confidence: routeTestConfidence(
			score,
			1,
			string(role)+"_confidence",
		),
		Evidence: []routecontract.Evidence{
			{
				Type:          routecontract.EvidenceTypeTrajectoryEndpointProximity,
				SourceName:    sourceName,
				SourceVersion: "route-test-v1",
				Score:         score,
				Weight:        1,
				ObservedAt:    observedAt.UTC(),
				Summary:       "Validated endpoint evidence.",
				Attributes: []routecontract.EvidenceAttribute{
					{Key: "airport_icao", Value: icaoCode},
				},
			},
		},
	}
}

func routeTestConfidence(
	score float64,
	evidenceCount int,
	code string,
) routecontract.Confidence {
	return routecontract.Confidence{
		Score:         score,
		Level:         routecontract.ConfidenceLevelForScore(score),
		EvidenceCount: evidenceCount,
		Reasons: []routecontract.ConfidenceReason{
			{
				Code:         code,
				Message:      "Validated route confidence.",
				Contribution: score,
			},
		},
	}
}

func routeTestPlan() historicalwindow.Plan {
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
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
		Granularity:        historicalcontract.GranularityHour,
		EffectiveWindow:    &window,
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
