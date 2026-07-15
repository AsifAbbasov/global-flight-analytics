package projectionarrival

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestEstimateAttachesArrivalWithinProjectionHorizon(
	t *testing.T,
) {
	estimator := newArrivalEstimator(t)
	request := arrivalTestRequest(
		0.03,
		[]float64{
			0.01,
			0.025,
			0.04,
		},
	)

	result, err := estimator.Estimate(request)
	if err != nil {
		t.Fatalf(
			"Estimate() error = %v",
			err,
		)
	}

	if result.Arrival == nil {
		t.Fatal(
			"arrival estimate is nil",
		)
	}
	if result.Status !=
		projectioncontract.
			ResultStatusComplete {
		t.Fatalf(
			"status = %q, want complete",
			result.Status,
		)
	}
	if result.Arrival.AirportICAOCode !=
		"BBBB" {
		t.Fatalf(
			"airport = %q, want BBBB",
			result.Arrival.
				AirportICAOCode,
		)
	}

	lowerBound :=
		request.Projection.
			Horizon.AsOfTime.Add(
			time.Minute,
		)
	upperBound :=
		request.Projection.
			Horizon.AsOfTime.Add(
			2 * time.Minute,
		)
	if !result.Arrival.
		EstimatedTime.After(
		lowerBound,
	) ||
		!result.Arrival.
			EstimatedTime.Before(
			upperBound,
		) {
		t.Fatalf(
			"estimated time = %s, want between %s and %s",
			result.Arrival.EstimatedTime,
			lowerBound,
			upperBound,
		)
	}
	if hasLimitation(
		result.Arrival.Limitations,
		"arrival_extrapolated_beyond_projection_horizon",
	) {
		t.Fatalf(
			"within-horizon arrival was marked extrapolated: %#v",
			result.Arrival.Limitations,
		)
	}
	if !hasExplanation(
		result.Explanations,
		MethodName,
	) {
		t.Fatalf(
			"arrival explanation missing: %#v",
			result.Explanations,
		)
	}

	report := projectioncontract.Validate(
		result,
	)
	if report.Status !=
		projectioncontract.
			ValidationStatusValid {
		t.Fatalf(
			"generated projection contract invalid: %#v",
			report.Issues,
		)
	}
}

func TestEstimateBuildsBoundedExtrapolatedArrival(
	t *testing.T,
) {
	estimator := newArrivalEstimator(t)
	request := arrivalTestRequest(
		0.10,
		[]float64{
			0.01,
			0.02,
			0.03,
		},
	)

	first, err := estimator.Estimate(request)
	if err != nil {
		t.Fatalf(
			"Estimate() error = %v",
			err,
		)
	}
	second, err := estimator.Estimate(request)
	if err != nil {
		t.Fatalf(
			"second Estimate() error = %v",
			err,
		)
	}

	if first.Arrival == nil {
		t.Fatal(
			"arrival estimate is nil",
		)
	}
	if first.Status !=
		projectioncontract.
			ResultStatusLimited {
		t.Fatalf(
			"status = %q, want limited",
			first.Status,
		)
	}
	if !first.Arrival.
		EstimatedTime.After(
		first.Horizon.EndTime,
	) {
		t.Fatalf(
			"estimated time = %s, horizon end = %s",
			first.Arrival.EstimatedTime,
			first.Horizon.EndTime,
		)
	}
	if !hasLimitation(
		first.Arrival.Limitations,
		"arrival_extrapolated_beyond_projection_horizon",
	) {
		t.Fatalf(
			"extrapolation limitation missing: %#v",
			first.Arrival.Limitations,
		)
	}
	if first.Provenance.InputFingerprint !=
		second.Provenance.InputFingerprint {
		t.Fatal(
			"deterministic input produced different fingerprints",
		)
	}
	if !first.Arrival.EarliestTime.Before(
		first.Arrival.EstimatedTime,
	) ||
		!first.Arrival.LatestTime.After(
			first.Arrival.EstimatedTime,
		) {
		t.Fatalf(
			"arrival interval is not ordered: %#v",
			first.Arrival,
		)
	}
}

func TestEstimateExcludesFutureCurrentTrajectoryPoints(
	t *testing.T,
) {
	estimator := newArrivalEstimator(t)
	request := arrivalTestRequest(
		0.10,
		[]float64{
			0.01,
			0.02,
			0.03,
		},
	)

	withoutFuture, err :=
		estimator.Estimate(request)
	if err != nil {
		t.Fatalf(
			"Estimate() error = %v",
			err,
		)
	}

	futurePoint :=
		request.CurrentTrajectory.Points[len(request.CurrentTrajectory.Points)-1]
	futurePoint.ID = "future-point"
	futurePoint.ObservedAt =
		request.Projection.
			Horizon.AsOfTime.Add(
			time.Minute,
		)
	futurePoint.Longitude = 100
	request.CurrentTrajectory.Points =
		append(
			request.CurrentTrajectory.Points,
			futurePoint,
		)

	withFuture, err :=
		estimator.Estimate(request)
	if err != nil {
		t.Fatalf(
			"Estimate() with future point error = %v",
			err,
		)
	}

	if withoutFuture.Provenance.
		InputFingerprint !=
		withFuture.Provenance.
			InputFingerprint ||
		!withoutFuture.Arrival.
			EstimatedTime.Equal(
			withFuture.Arrival.
				EstimatedTime,
		) {
		t.Fatal(
			"future current-trajectory point changed the as-of arrival estimate",
		)
	}
}

func TestEstimateWithholdsLowConfidenceDestination(
	t *testing.T,
) {
	estimator := newArrivalEstimator(t)
	request := arrivalTestRequest(
		0.10,
		[]float64{
			0.01,
			0.02,
			0.03,
		},
	)
	request.Route.Destination.
		Confidence.Score = 0.4
	request.Route.Destination.
		Confidence.Level =
		routecontract.ConfidenceLevelLow

	result, err := estimator.Estimate(request)
	if err != nil {
		t.Fatalf(
			"Estimate() error = %v",
			err,
		)
	}

	if result.Arrival != nil ||
		result.Status !=
			projectioncontract.
				ResultStatusLimited ||
		!hasLimitation(
			result.Limitations,
			"estimated_arrival_unavailable_reason",
		) {
		t.Fatalf(
			"unexpected withheld result: %#v",
			result,
		)
	}

	report := projectioncontract.Validate(
		result,
	)
	if report.Status !=
		projectioncontract.
			ValidationStatusValid {
		t.Fatalf(
			"withheld result invalid: %#v",
			report.Issues,
		)
	}
}

func TestEstimateWithholdsArrivalBeyondMaximumDuration(
	t *testing.T,
) {
	config := validArrivalConfig()
	config.MaximumEstimatedArrivalDuration =
		time.Minute
	estimator, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	result, err := estimator.Estimate(
		arrivalTestRequest(
			1,
			[]float64{
				0.01,
				0.02,
				0.03,
			},
		),
	)
	if err != nil {
		t.Fatalf(
			"Estimate() error = %v",
			err,
		)
	}

	if result.Arrival != nil ||
		!hasReasonText(
			result.Limitations,
			"arrival_speed_or_duration_unavailable",
		) {
		t.Fatalf(
			"unbounded arrival was not withheld: %#v",
			result,
		)
	}
}

func TestEstimateRejectsFutureRouteEvidenceAndTrajectoryMismatch(
	t *testing.T,
) {
	estimator := newArrivalEstimator(t)

	futureRouteRequest :=
		arrivalTestRequest(
			0.10,
			[]float64{
				0.01,
				0.02,
				0.03,
			},
		)
	futureAsOf :=
		futureRouteRequest.Projection.
			Horizon.AsOfTime.Add(
			time.Minute,
		)
	futureRouteRequest.Route.Window.
		AsOfTime = futureAsOf
	futureRouteRequest.Route.Window.
		EndTime =
		futureRouteRequest.Projection.
			Horizon.AsOfTime
	futureRouteRequest.Route.Origin.
		Evidence[0].ObservedAt =
		futureAsOf
	futureRouteRequest.Route.Destination.
		Evidence[0].ObservedAt =
		futureAsOf
	futureRouteRequest.Route.Provenance.
		TrajectoryUpdatedAt =
		futureAsOf
	futureRouteRequest.Route.GeneratedAt =
		futureAsOf.Add(time.Second)
	futureRouteRequest.GeneratedAt =
		futureAsOf.Add(2 * time.Second)

	_, err := estimator.Estimate(
		futureRouteRequest,
	)
	if !errors.Is(
		err,
		ErrFutureRouteEvidence,
	) {
		t.Fatalf(
			"future route error = %v",
			err,
		)
	}

	mismatchRequest :=
		arrivalTestRequest(
			0.10,
			[]float64{
				0.01,
				0.02,
				0.03,
			},
		)
	mismatchRequest.Route.TrajectoryID =
		"other-trajectory"

	_, err = estimator.Estimate(
		mismatchRequest,
	)
	if !errors.Is(
		err,
		ErrTrajectoryMismatch,
	) {
		t.Fatalf(
			"trajectory mismatch error = %v",
			err,
		)
	}
}

func newArrivalEstimator(
	t *testing.T,
) *Estimator {
	t.Helper()

	estimator, err := New(
		validArrivalConfig(),
	)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	return estimator
}

func arrivalTestRequest(
	destinationLongitude float64,
	projectedLongitudes []float64,
) Request {
	asOfTime := arrivalTestAsOfTime()
	projection :=
		validProjectionResult(
			asOfTime,
			projectedLongitudes,
		)
	route := validRouteResult(
		asOfTime,
		destinationLongitude,
	)

	return Request{
		Projection: projection,
		Route:      route,
		CurrentTrajectory: validCurrentTrajectory(
			asOfTime,
		),
		GeneratedAt: asOfTime.Add(
			3 * time.Second,
		),
	}
}

func validProjectionResult(
	asOfTime time.Time,
	longitudes []float64,
) projectioncontract.Result {
	points := make(
		[]projectioncontract.ProjectionPoint,
		0,
		len(longitudes),
	)
	for index, longitude := range longitudes {
		points = append(
			points,
			projectioncontract.ProjectionPoint{
				Sequence: index,
				ForecastTime: asOfTime.Add(
					time.Duration(
						index+1,
					) * time.Minute,
				),
				Position: projectioncontract.Position{
					Latitude:  0,
					Longitude: longitude,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 200 +
						float64(index)*100,
				},
				Confidence: projectioncontract.Confidence{
					Score: 0.8 -
						float64(index)*0.05,
					Level: projectioncontract.
						ConfidenceLevelHigh,
					Reasons: []projectioncontract.ConfidenceReason{
						{
							Code:    "projection_point",
							Message: "Projection point confidence.",
							Contribution: 0.8 -
								float64(index)*0.05,
						},
					},
				},
			},
		)
	}

	return projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status: projectioncontract.
			ResultStatusComplete,
		TrajectoryID: "trajectory-001",
		FlightID:     "flight-001",
		AircraftID:   "aircraft-001",
		ICAO24:       "4A1234",
		Callsign:     "AHY123",
		Method: projectioncontract.Method{
			Name:    "test_projection",
			Version: "test-projection-v1",
			DecisionClass: projectioncontract.
				DecisionClassExperimental,
		},
		Horizon: projectioncontract.Horizon{
			AsOfTime: asOfTime,
			EndTime: asOfTime.Add(
				time.Duration(
					len(longitudes),
				) * time.Minute,
			),
			Step: time.Minute,
		},
		Points: points,
		Confidence: projectioncontract.Confidence{
			Score: 0.7,
			Level: projectioncontract.
				ConfidenceLevelMedium,
			Reasons: []projectioncontract.ConfidenceReason{
				{
					Code:         "projection_result",
					Message:      "Projection result confidence.",
					Contribution: 0.7,
				},
			},
		},
		Limitations: []projectioncontract.Limitation{
			{
				Code:    "research_only",
				Message: "Projection is a research estimate.",
				Scope:   "result",
			},
		},
		Explanations: []projectioncontract.Explanation{
			{
				Code:    "test_projection",
				Message: "Test position projection.",
			},
		},
		ScopeGuard: projectioncontract.
			ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: "sha256:" +
				strings.Repeat(
					"a",
					64,
				),
			Inputs: []projectioncontract.InputReference{
				{
					Name: "current_position",
					Classification: projectioncontract.
						InputClassificationObserved,
					SourceName: "test-source",
					ObservedAt: asOfTime,
					RetrievedAt: asOfTime.Add(
						time.Second,
					),
				},
			},
			LatestInputObservedAt: asOfTime,
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}
}

func validRouteResult(
	asOfTime time.Time,
	destinationLongitude float64,
) routecontract.Result {
	originEvidence :=
		validRouteEvidence(
			asOfTime,
			"origin",
		)
	destinationEvidence :=
		validRouteEvidence(
			asOfTime,
			"destination",
		)

	return routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status: routecontract.
			RouteStatusComplete,
		TrajectoryID: "trajectory-001",
		FlightID:     "flight-001",
		AircraftID:   "aircraft-001",
		ICAO24:       "4A1234",
		Callsign:     "AHY123",
		Window: routecontract.RouteWindow{
			StartTime: asOfTime.Add(
				-10 * time.Minute,
			),
			EndTime:  asOfTime,
			AsOfTime: asOfTime,
		},
		Origin: &routecontract.EndpointInference{
			Role: routecontract.
				EndpointRoleOrigin,
			Airport: routecontract.AirportReference{
				ICAOCode:   "AAAA",
				IATACode:   "AAA",
				Name:       "Origin Airport",
				Latitude:   0,
				Longitude:  -1,
				ElevationM: 0,
			},
			DistanceKM: 1,
			Confidence: validRouteConfidence(
				0.9,
				1,
				"origin_confidence",
			),
			Evidence: []routecontract.Evidence{
				originEvidence,
			},
		},
		Destination: &routecontract.EndpointInference{
			Role: routecontract.
				EndpointRoleDestination,
			Airport: routecontract.AirportReference{
				ICAOCode:   "BBBB",
				IATACode:   "BBB",
				Name:       "Destination Airport",
				Latitude:   0,
				Longitude:  destinationLongitude,
				ElevationM: 0,
			},
			DistanceKM: 1,
			Confidence: validRouteConfidence(
				0.9,
				1,
				"destination_confidence",
			),
			Evidence: []routecontract.Evidence{
				destinationEvidence,
			},
		},
		Summary: routecontract.RouteSummary{
			GreatCircleDistanceKM: 100,
			SameAirport:           false,
		},
		Confidence: validRouteConfidence(
			0.9,
			2,
			"route_confidence",
		),
		Provenance: routecontract.Provenance{
			ResolverVersion: "route-resolver-test-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat(
					"b",
					64,
				),
			TrajectoryUpdatedAt: asOfTime,
			SourceNames: []string{
				"test-source",
			},
		},
		GeneratedAt: asOfTime.Add(
			2 * time.Second,
		),
	}
}

func validRouteEvidence(
	asOfTime time.Time,
	summary string,
) routecontract.Evidence {
	return routecontract.Evidence{
		Type: routecontract.
			EvidenceTypeTrajectoryEndpointProximity,
		SourceName:    "test-source",
		SourceVersion: "test-source-v1",
		Score:         0.9,
		Weight:        1,
		ObservedAt:    asOfTime,
		Summary:       summary,
	}
}

func validRouteConfidence(
	score float64,
	evidenceCount int,
	code string,
) routecontract.Confidence {
	return routecontract.Confidence{
		Score: score,
		Level: routecontract.
			ConfidenceLevelForScore(
				score,
			),
		EvidenceCount: evidenceCount,
		Reasons: []routecontract.ConfidenceReason{
			{
				Code:         code,
				Message:      "Route confidence reason.",
				Contribution: score,
			},
		},
	}
}

func validCurrentTrajectory(
	asOfTime time.Time,
) trajectory.FlightTrajectory {
	return trajectory.FlightTrajectory{
		ID:         "trajectory-001",
		FlightID:   "flight-001",
		AircraftID: "aircraft-001",
		ICAO24:     "4A1234",
		Callsign:   "AHY123",
		StartTime: asOfTime.Add(
			-4 * time.Minute,
		),
		EndTime:      asOfTime,
		PointCount:   2,
		QualityScore: 0.9,
		SourceName:   "test-source",
		Points: []trajectory.TrackPoint4D{
			{
				ID:        "current-previous",
				Latitude:  0,
				Longitude: -0.01,
				ObservedAt: asOfTime.Add(
					-time.Minute,
				),
				SourceName: "test-source",
			},
			{
				ID:         "current-endpoint",
				Latitude:   0,
				Longitude:  0,
				ObservedAt: asOfTime,
				SourceName: "test-source",
			},
		},
		UpdatedAt: asOfTime,
	}
}

func arrivalTestAsOfTime() time.Time {
	return time.Date(
		2026,
		time.July,
		16,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}

func hasLimitation(
	items []projectioncontract.Limitation,
	code string,
) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}

	return false
}

func hasReasonText(
	items []projectioncontract.Limitation,
	reason string,
) bool {
	for _, item := range items {
		if strings.Contains(
			item.Message,
			reason,
		) {
			return true
		}
	}

	return false
}

func hasExplanation(
	items []projectioncontract.Explanation,
	code string,
) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}

	return false
}

func TestArrivalConfidenceIsFinite(
	t *testing.T,
) {
	estimator := newArrivalEstimator(t)
	request := arrivalTestRequest(
		0.10,
		[]float64{
			0.01,
			0.02,
			0.03,
		},
	)

	result, err := estimator.Estimate(request)
	if err != nil {
		t.Fatalf(
			"Estimate() error = %v",
			err,
		)
	}
	if result.Arrival == nil ||
		math.IsNaN(
			result.Arrival.
				Confidence.Score,
		) ||
		math.IsInf(
			result.Arrival.
				Confidence.Score,
			0,
		) {
		t.Fatalf(
			"arrival confidence is invalid: %#v",
			result.Arrival,
		)
	}
}
