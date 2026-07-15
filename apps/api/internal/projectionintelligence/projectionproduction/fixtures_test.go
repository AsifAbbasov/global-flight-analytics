package projectionproduction

import (
	"errors"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionarrival"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontinuation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionfreshness"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type fakeHorizonPlanner struct {
	plan projectionhorizon.Plan
	err  error
}

func (fake fakeHorizonPlanner) Build(
	projectionhorizon.Request,
) (projectionhorizon.Plan, error) {
	return fake.plan.Clone(), fake.err
}

type fakeKinematicProjector struct {
	result projectioncontract.Result
	err    error
	calls  int
}

func (fake *fakeKinematicProjector) Project(
	projectionbaseline.Request,
) (projectioncontract.Result, error) {
	fake.calls++
	return fake.result.Clone(), fake.err
}

type fakeHistoricalProjector struct {
	result projectioncontract.Result
	err    error
	calls  int
}

func (fake *fakeHistoricalProjector) Project(
	projectioncontinuation.Request,
) (projectioncontract.Result, error) {
	fake.calls++
	return fake.result.Clone(), fake.err
}

type fakeNeighborSelector struct {
	result projectionneighbors.Result
	err    error
	calls  int
}

func (fake *fakeNeighborSelector) Select(
	projectionneighbors.Request,
) (projectionneighbors.Result, error) {
	fake.calls++
	return fake.result.Clone(), fake.err
}

type fakePatternEvaluator struct {
	result projectionpatternconfidence.Result
	err    error
	calls  int
}

func (fake *fakePatternEvaluator) Evaluate(
	projectionneighbors.Result,
) (projectionpatternconfidence.Result, error) {
	fake.calls++
	return fake.result.Clone(), fake.err
}

type fakeFreshnessEvaluator struct {
	result projectionfreshness.Result
	err    error
	calls  int
}

func (fake *fakeFreshnessEvaluator) Evaluate(
	projectionneighbors.Result,
	projectionpatternconfidence.Result,
) (projectionfreshness.Result, error) {
	fake.calls++
	return fake.result.Clone(), fake.err
}

type fakeRouteFrequencyEvaluator struct {
	result projectionroutefrequency.Result
	err    error
	calls  int
}

func (fake *fakeRouteFrequencyEvaluator) Evaluate(
	routecontract.Result,
	projectionroutefrequency.HistorySummary,
) (projectionroutefrequency.Result, error) {
	fake.calls++
	return fake.result.Clone(), fake.err
}

type fakeArrivalEstimator struct {
	result projectioncontract.Result
	err    error
	calls  int
	attach bool
}

func (fake *fakeArrivalEstimator) Estimate(
	request projectionarrival.Request,
) (projectioncontract.Result, error) {
	fake.calls++
	if fake.err != nil {
		return projectioncontract.Result{}, fake.err
	}
	if fake.result.TrajectoryID != "" {
		return fake.result.Clone(), nil
	}

	result := request.Projection.Clone()
	if fake.attach {
		result.Arrival =
			&projectioncontract.ArrivalEstimate{
				AirportICAOCode: "LTBA",
				EarliestTime: request.Projection.
					Horizon.AsOfTime.Add(
					3 * time.Minute,
				),
				EstimatedTime: request.Projection.
					Horizon.AsOfTime.Add(
					4 * time.Minute,
				),
				LatestTime: request.Projection.
					Horizon.AsOfTime.Add(
					5 * time.Minute,
				),
				Confidence: projectioncontract.Confidence{
					Score: 0.7,
					Level: projectioncontract.
						ConfidenceLevelMedium,
					Reasons: []projectioncontract.
						ConfidenceReason{
						{
							Code:         "arrival",
							Message:      "Arrival confidence.",
							Contribution: 0.7,
						},
					},
				},
				Limitations: []projectioncontract.Limitation{
					{
						Code:    "arrival_radius",
						Message: "Arrival represents airport-radius entry.",
						Scope:   "arrival",
					},
				},
			}
	}

	return result, nil
}

type productionFixture struct {
	config  Config
	request Request

	kinematic  *fakeKinematicProjector
	historical *fakeHistoricalProjector
	selector   *fakeNeighborSelector
	pattern    *fakePatternEvaluator
	freshness  *fakeFreshnessEvaluator
	frequency  *fakeRouteFrequencyEvaluator
	arrival    *fakeArrivalEstimator
}

func newProductionFixture() productionFixture {
	asOfTime := productionTestAsOfTime()
	plan := projectionhorizon.Plan{
		Version:           projectionhorizon.Version,
		PolicyName:        "production-test-policy",
		AsOfTime:          asOfTime,
		EndTime:           asOfTime.Add(2 * time.Minute),
		Step:              time.Minute,
		RequestedDuration: 2 * time.Minute,
		EffectiveDuration: 2 * time.Minute,
		ForecastTimes: []time.Time{
			asOfTime.Add(time.Minute),
			asOfTime.Add(2 * time.Minute),
		},
	}

	kinematic := &fakeKinematicProjector{
		result: validProductionProjection(
			asOfTime,
			projectionbaseline.MethodName,
			projectionbaseline.Version,
			projectioncontract.DecisionClassPhysicsDerived,
			"b",
		),
	}
	historical := &fakeHistoricalProjector{
		result: validProductionProjection(
			asOfTime,
			projectioncontinuation.MethodName,
			projectioncontinuation.Version,
			projectioncontract.DecisionClassExperimental,
			"c",
		),
	}
	selector := &fakeNeighborSelector{
		result: validProductionSelection(asOfTime),
	}
	pattern := &fakePatternEvaluator{
		result: validProductionPattern(),
	}
	freshness := &fakeFreshnessEvaluator{
		result: validProductionFreshness(asOfTime),
	}
	frequency := &fakeRouteFrequencyEvaluator{
		result: validProductionFrequency(asOfTime),
	}
	arrival := &fakeArrivalEstimator{
		attach: true,
	}

	config := Config{
		HorizonPlanner: fakeHorizonPlanner{
			plan: plan,
		},
		KinematicProjector:          kinematic,
		HistoricalProjector:         historical,
		NeighborSelector:            selector,
		PatternConfidenceEvaluator:  pattern,
		FreshnessEvaluator:          freshness,
		RouteFrequencyEvaluator:     frequency,
		ArrivalEstimator:            arrival,
		FreshnessLimitedPolicy:      LimitedEvidenceReject,
		RouteFrequencyLimitedPolicy: LimitedEvidenceReject,
		DependencyFailurePolicy:     DependencyFailureFallbackToKinematic,
		ArrivalFailurePolicy:        ArrivalFailurePreserveProjection,
	}

	history := validProductionHistory(asOfTime)
	request := Request{
		CurrentTrajectory: validProductionTrajectory(asOfTime),
		HistoricalCandidates: []trajectory.FlightTrajectory{
			{ID: "historical-a"},
			{ID: "historical-b"},
		},
		Route:             validProductionRoute(asOfTime),
		RouteHistory:      &history,
		AsOfTime:          asOfTime,
		RequestedDuration: 2 * time.Minute,
		GeneratedAt:       asOfTime.Add(3 * time.Second),
	}

	return productionFixture{
		config:     config,
		request:    request,
		kinematic:  kinematic,
		historical: historical,
		selector:   selector,
		pattern:    pattern,
		freshness:  freshness,
		frequency:  frequency,
		arrival:    arrival,
	}
}

func productionTestConfig() Config {
	fixture := newProductionFixture()
	return fixture.config
}

func validProductionProjection(
	asOfTime time.Time,
	methodName string,
	methodVersion string,
	decisionClass projectioncontract.DecisionClass,
	fingerprintCharacter string,
) projectioncontract.Result {
	points := make(
		[]projectioncontract.ProjectionPoint,
		0,
		2,
	)
	for index := 0; index < 2; index++ {
		points = append(
			points,
			projectioncontract.ProjectionPoint{
				Sequence: index,
				ForecastTime: asOfTime.Add(
					time.Duration(index+1) * time.Minute,
				),
				Position: projectioncontract.Position{
					Latitude:  40 + float64(index)*0.01,
					Longitude: 50 + float64(index)*0.01,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 500 + float64(index)*100,
				},
				Confidence: projectioncontract.Confidence{
					Score: 0.8 - float64(index)*0.1,
					Level: projectioncontract.ConfidenceLevelMedium,
					Reasons: []projectioncontract.ConfidenceReason{
						{
							Code:         "point_confidence",
							Message:      "Point confidence.",
							Contribution: 0.8 - float64(index)*0.1,
						},
					},
				},
			},
		)
	}

	return projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        projectioncontract.ResultStatusComplete,
		TrajectoryID:  "trajectory-001",
		FlightID:      "flight-001",
		AircraftID:    "aircraft-001",
		ICAO24:        "4A1234",
		Callsign:      "AHY123",
		Method: projectioncontract.Method{
			Name:          methodName,
			Version:       methodVersion,
			DecisionClass: decisionClass,
		},
		Horizon: projectioncontract.Horizon{
			AsOfTime: asOfTime,
			EndTime:  asOfTime.Add(2 * time.Minute),
			Step:     time.Minute,
		},
		Points: points,
		Confidence: projectioncontract.Confidence{
			Score: 0.7,
			Level: projectioncontract.ConfidenceLevelMedium,
			Reasons: []projectioncontract.ConfidenceReason{
				{
					Code:         "result_confidence",
					Message:      "Result confidence.",
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
				Code:    methodName,
				Message: "Production composition test projection.",
			},
		},
		ScopeGuard: projectioncontract.ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: "sha256:" + strings.Repeat(
				fingerprintCharacter,
				64,
			),
			Inputs: []projectioncontract.InputReference{
				{
					Name:           "current_trajectory",
					Classification: projectioncontract.InputClassificationObserved,
					SourceName:     "test-source",
					ObservedAt:     asOfTime,
					RetrievedAt:    asOfTime.Add(time.Second),
				},
			},
			LatestInputObservedAt: asOfTime,
		},
		GeneratedAt: asOfTime.Add(3 * time.Second),
	}
}

func validProductionSelection(
	asOfTime time.Time,
) projectionneighbors.Result {
	neighbors := []projectionneighbors.Neighbor{
		validProductionNeighbor(
			"historical-a",
			0.9,
			24*time.Hour,
			asOfTime,
			"d",
		),
		validProductionNeighbor(
			"historical-b",
			0.8,
			48*time.Hour,
			asOfTime,
			"e",
		),
	}

	return projectionneighbors.Result{
		Version:                      projectionneighbors.Version,
		Status:                       projectionneighbors.StatusComplete,
		CurrentTrajectoryID:          "trajectory-001",
		AsOfTime:                     asOfTime,
		RequiredContinuationDuration: 2 * time.Minute,
		InputCandidateCount:          2,
		CheckedCandidateCount:        2,
		QualifiedCandidateCount:      2,
		RejectedCandidateCount:       0,
		SelectionLimit:               2,
		Neighbors:                    neighbors,
		InputFingerprint:             "sha256:" + strings.Repeat("f", 64),
	}
}

func validProductionNeighbor(
	id string,
	score float64,
	age time.Duration,
	asOfTime time.Time,
	fingerprintCharacter string,
) projectionneighbors.Neighbor {
	endTime := asOfTime.Add(-age)
	anchorTime := endTime.Add(-3 * time.Minute)

	return projectionneighbors.Neighbor{
		TrajectoryID:    id,
		SimilarityScore: score,
		SimilarityLevel: historicalsimilarity.LevelForScore(score),
		SimilarityInputFingerprint: "sha256:" + strings.Repeat(
			fingerprintCharacter,
			64,
		),
		AnchorPointIndex:       4,
		AnchorObservedAt:       anchorTime,
		AnchorDistanceKM:       1,
		CandidateStartTime:     endTime.Add(-10 * time.Minute),
		CandidateEndTime:       endTime,
		CandidateAge:           age,
		PrefixPointCount:       5,
		ContinuationPointCount: 3,
		ContinuationEndTime:    anchorTime.Add(3 * time.Minute),
	}
}

func validProductionPattern() projectionpatternconfidence.Result {
	return projectionpatternconfidence.Result{
		Version:                 projectionpatternconfidence.Version,
		Status:                  projectionpatternconfidence.StatusComplete,
		Usable:                  true,
		NeighborCount:           2,
		TargetNeighborCount:     2,
		MeanSimilarityScore:     0.85,
		MeanCandidateAgeSeconds: 36 * 60 * 60,
		MeanAnchorDistanceKM:    1,
		Score:                   0.8,
		Level:                   projectioncontract.ConfidenceLevelHigh,
		Components: []projectionpatternconfidence.Component{
			{Name: projectionpatternconfidence.ComponentSimilarity, Score: 0.85, Weight: 0.25},
			{Name: projectionpatternconfidence.ComponentSupport, Score: 1, Weight: 0.25},
			{Name: projectionpatternconfidence.ComponentFreshness, Score: 0.8, Weight: 0.25},
			{Name: projectionpatternconfidence.ComponentAnchorProximity, Score: 0.9, Weight: 0.25},
		},
		SelectedTrajectoryIDs: []string{"historical-a", "historical-b"},
		InputFingerprint:      "sha256:" + strings.Repeat("1", 64),
	}
}

func validProductionFreshness(
	asOfTime time.Time,
) projectionfreshness.Result {
	return projectionfreshness.Result{
		Version:             projectionfreshness.Version,
		Decision:            projectionfreshness.DecisionAllowed,
		Usable:              true,
		AsOfTime:            asOfTime,
		NeighborCount:       2,
		RecentNeighborCount: 2,
		NewestNeighborAge:   24 * time.Hour,
		MeanNeighborAge:     36 * time.Hour,
		OldestNeighborAge:   48 * time.Hour,
		Score:               0.8,
		Components: []projectionfreshness.Component{
			{Name: projectionfreshness.ComponentNewestAge, Score: 0.9, Weight: 0.25},
			{Name: projectionfreshness.ComponentMeanAge, Score: 0.8, Weight: 0.25},
			{Name: projectionfreshness.ComponentOldestAge, Score: 0.7, Weight: 0.25},
			{Name: projectionfreshness.ComponentRecentSupport, Score: 1, Weight: 0.25},
		},
		SelectedTrajectoryIDs: []string{"historical-a", "historical-b"},
		InputFingerprint:      "sha256:" + strings.Repeat("2", 64),
	}
}

func validProductionFrequency(
	asOfTime time.Time,
) projectionroutefrequency.Result {
	return projectionroutefrequency.Result{
		Version:                projectionroutefrequency.Version,
		Decision:               projectionroutefrequency.DecisionAllowed,
		Usable:                 true,
		RouteKey:               "UBBB>LTBA",
		AsOfTime:               asOfTime,
		ObservationCount:       12,
		DistinctFlightCount:    10,
		DistinctDayCount:       7,
		RecentObservationCount: 5,
		LatestObservationAge:   24 * time.Hour,
		RouteConfidenceScore:   0.9,
		Score:                  0.85,
		Components: []projectionroutefrequency.Component{
			{Name: projectionroutefrequency.ComponentObservationCount, Score: 1, Weight: 0.2},
			{Name: projectionroutefrequency.ComponentDistinctDays, Score: 1, Weight: 0.2},
			{Name: projectionroutefrequency.ComponentRecentObservations, Score: 1, Weight: 0.2},
			{Name: projectionroutefrequency.ComponentLatestObservation, Score: 0.8, Weight: 0.2},
			{Name: projectionroutefrequency.ComponentRouteConfidence, Score: 0.9, Weight: 0.2},
		},
		HistoryInputFingerprint: "sha256:" + strings.Repeat("3", 64),
		InputFingerprint:        "sha256:" + strings.Repeat("4", 64),
	}
}

func validProductionHistory(
	asOfTime time.Time,
) projectionroutefrequency.HistorySummary {
	return projectionroutefrequency.HistorySummary{
		RouteKey:               "UBBB>LTBA",
		WindowStart:            asOfTime.Add(-60 * 24 * time.Hour),
		WindowEnd:              asOfTime,
		AsOfTime:               asOfTime,
		ObservationCount:       12,
		DistinctFlightCount:    10,
		DistinctDayCount:       7,
		RecentObservationCount: 5,
		LastObservedAt:         asOfTime.Add(-24 * time.Hour),
		SourceNames:            []string{"historical-route-store"},
		InputFingerprint:       "sha256:" + strings.Repeat("5", 64),
	}
}

func validProductionTrajectory(
	asOfTime time.Time,
) trajectory.FlightTrajectory {
	return trajectory.FlightTrajectory{
		ID:           "trajectory-001",
		FlightID:     "flight-001",
		AircraftID:   "aircraft-001",
		ICAO24:       "4A1234",
		Callsign:     "AHY123",
		StartTime:    asOfTime.Add(-5 * time.Minute),
		EndTime:      asOfTime,
		PointCount:   2,
		QualityScore: 0.9,
		SourceName:   "test-source",
		Points: []trajectory.TrackPoint4D{
			{
				ID:         "point-1",
				Latitude:   40,
				Longitude:  50,
				ObservedAt: asOfTime.Add(-time.Minute),
				SourceName: "test-source",
			},
			{
				ID:         "point-2",
				Latitude:   40.01,
				Longitude:  50.01,
				ObservedAt: asOfTime,
				SourceName: "test-source",
			},
		},
		UpdatedAt: asOfTime,
	}
}

func validProductionRoute(
	asOfTime time.Time,
) routecontract.Result {
	originEvidence := validProductionRouteEvidence(
		asOfTime,
		"origin",
	)
	destinationEvidence := validProductionRouteEvidence(
		asOfTime,
		"destination",
	)

	return routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        routecontract.RouteStatusComplete,
		TrajectoryID:  "trajectory-001",
		FlightID:      "flight-001",
		AircraftID:    "aircraft-001",
		ICAO24:        "4A1234",
		Callsign:      "AHY123",
		Window: routecontract.RouteWindow{
			StartTime: asOfTime.Add(-30 * time.Minute),
			EndTime:   asOfTime,
			AsOfTime:  asOfTime,
		},
		Origin: &routecontract.EndpointInference{
			Role: routecontract.EndpointRoleOrigin,
			Airport: routecontract.AirportReference{
				ICAOCode:  "UBBB",
				Name:      "Heydar Aliyev International Airport",
				Latitude:  40.4675,
				Longitude: 50.0467,
			},
			DistanceKM: 5,
			Confidence: validProductionRouteConfidence(
				0.9,
				1,
				"origin",
			),
			Evidence: []routecontract.Evidence{originEvidence},
		},
		Destination: &routecontract.EndpointInference{
			Role: routecontract.EndpointRoleDestination,
			Airport: routecontract.AirportReference{
				ICAOCode:  "LTBA",
				Name:      "Istanbul Airport",
				Latitude:  40.9769,
				Longitude: 28.8146,
			},
			DistanceKM: 6,
			Confidence: validProductionRouteConfidence(
				0.9,
				1,
				"destination",
			),
			Evidence: []routecontract.Evidence{destinationEvidence},
		},
		Summary: routecontract.RouteSummary{
			GreatCircleDistanceKM: 1760,
			SameAirport:           false,
		},
		Confidence: validProductionRouteConfidence(
			0.9,
			2,
			"route",
		),
		Provenance: routecontract.Provenance{
			ResolverVersion:     "route-resolver-test-v1",
			InputFingerprint:    "sha256:" + strings.Repeat("6", 64),
			TrajectoryUpdatedAt: asOfTime,
			SourceNames:         []string{"test-source"},
		},
		GeneratedAt: asOfTime.Add(2 * time.Second),
	}
}

func validProductionRouteConfidence(
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
				Message:      "Route confidence reason.",
				Contribution: score,
			},
		},
	}
}

func validProductionRouteEvidence(
	asOfTime time.Time,
	summary string,
) routecontract.Evidence {
	return routecontract.Evidence{
		Type:          routecontract.EvidenceTypeTrajectoryEndpointProximity,
		SourceName:    "test-source",
		SourceVersion: "test-source-v1",
		Score:         0.9,
		Weight:        1,
		ObservedAt:    asOfTime,
		Summary:       summary,
	}
}

func productionTestAsOfTime() time.Time {
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

var errProductionDependency = errors.New(
	"production dependency failure",
)
