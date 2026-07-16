package main

import (
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontinuation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestBuildVerificationSchedule(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		16,
		12,
		34,
		56,
		999,
		time.FixedZone(
			"UTC plus four",
			4*60*60,
		),
	)

	schedule, err :=
		buildVerificationSchedule(now)
	if err != nil {
		t.Fatalf(
			"buildVerificationSchedule() error = %v",
			err,
		)
	}

	expectedGeneratedAt := time.Date(
		2026,
		time.July,
		16,
		8,
		34,
		56,
		0,
		time.UTC,
	)
	if !schedule.GeneratedAt.Equal(
		expectedGeneratedAt,
	) {
		t.Fatalf(
			"GeneratedAt = %s, want %s",
			schedule.GeneratedAt,
			expectedGeneratedAt,
		)
	}
	if !schedule.AsOfTime.Equal(
		expectedGeneratedAt.Add(
			-time.Minute,
		),
	) {
		t.Fatalf(
			"AsOfTime = %s",
			schedule.AsOfTime,
		)
	}
}

func TestVerificationRoutesAreValid(
	t *testing.T,
) {
	schedule, err :=
		buildVerificationSchedule(
			time.Date(
				2026,
				time.July,
				16,
				12,
				0,
				0,
				0,
				time.UTC,
			),
		)
	if err != nil {
		t.Fatalf(
			"build schedule: %v",
			err,
		)
	}

	for _, flight := range verificationFlights {
		endTime := schedule.AsOfTime.Add(
			-time.Duration(
				flight.AgeDays,
			) * 24 * time.Hour,
		)
		startTime := endTime.Add(
			-time.Duration(
				flight.PointCount-1,
			) * time.Minute,
		)
		result := buildCompleteRoute(
			flight,
			startTime,
			endTime,
			schedule.GeneratedAt,
		)
		report := routecontract.Validate(
			result,
		)
		if report.Status !=
			routecontract.ValidationStatusValid {
			t.Fatalf(
				"route for %s is invalid: %#v",
				flight.TrajectoryID,
				report.Issues,
			)
		}
	}
}

func TestHistoricalCandidateGeometryProvidesContinuation(
	t *testing.T,
) {
	currentLatitude, currentLongitude :=
		trackCoordinate(
			currentPointCount-1,
			0,
			0,
		)
	candidateLatitude,
		candidateLongitude :=
		trackCoordinate(
			currentPointCount-1,
			verificationFlights[1].
				LatitudeShift,
			verificationFlights[1].
				LongitudeShift,
		)
	finalLatitude, finalLongitude :=
		trackCoordinate(
			candidatePointCount-1,
			0,
			0,
		)

	if absolute(
		currentLatitude-
			candidateLatitude,
	) > 0.01 ||
		absolute(
			currentLongitude-
				candidateLongitude,
		) > 0.01 {
		t.Fatalf(
			"candidate anchor is too far from current endpoint",
		)
	}
	if finalLatitude <= currentLatitude ||
		finalLongitude <= currentLongitude {
		t.Fatalf(
			"candidate track does not continue beyond the current endpoint",
		)
	}
	if absolute(finalLatitude-40.56) > 1e-9 ||
		absolute(finalLongitude-50.12) > 1e-9 {
		t.Fatalf(
			"unexpected synthetic destination: %.6f %.6f",
			finalLatitude,
			finalLongitude,
		)
	}
}

func TestExpectedFixtureCounts(
	t *testing.T,
) {
	counts := expectedFixtureCounts()

	if counts.Trajectories != 5 ||
		counts.RouteResults != 5 ||
		counts.FlightStates != 42 {
		t.Fatalf(
			"unexpected fixture counts: %#v",
			counts,
		)
	}
}

func TestProjectionRequestURLUsesExplicitInputs(
	t *testing.T,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		16,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	requestURL := projectionRequestURL(
		verificationFlights[0].
			TrajectoryID,
		asOfTime,
		verificationDuration,
	)

	for _, fragment := range []string{
		"/api/v1/trajectories/" +
			verificationFlights[0].
				TrajectoryID +
			"/projection-intelligence?",
		"as_of_time=2026-07-16T12%3A00%3A00Z",
		"duration_seconds=180",
	} {
		if !strings.Contains(
			requestURL,
			fragment,
		) {
			t.Fatalf(
				"request URL is missing %q: %s",
				fragment,
				requestURL,
			)
		}
	}
}

func TestValidateHistoricalPayloadAcceptsAuthorizedResult(
	t *testing.T,
) {
	schedule, err :=
		buildVerificationSchedule(
			time.Date(
				2026,
				time.July,
				16,
				12,
				0,
				0,
				0,
				time.UTC,
			),
		)
	if err != nil {
		t.Fatalf(
			"build schedule: %v",
			err,
		)
	}

	payload :=
		validHistoricalPayload(
			schedule,
		)
	if err := validateHistoricalPayload(
		payload,
		schedule,
	); err != nil {
		t.Fatalf(
			"validateHistoricalPayload() error = %v",
			err,
		)
	}
}

func TestValidateHistoricalPayloadRejectsFallback(
	t *testing.T,
) {
	schedule, err :=
		buildVerificationSchedule(
			time.Date(
				2026,
				time.July,
				16,
				12,
				0,
				0,
				0,
				time.UTC,
			),
		)
	if err != nil {
		t.Fatalf(
			"build schedule: %v",
			err,
		)
	}

	payload :=
		validHistoricalPayload(
			schedule,
		)
	payload.Data.FallbackReason =
		"historical_projection_failed"

	err = validateHistoricalPayload(
		payload,
		schedule,
	)
	if err == nil ||
		!strings.Contains(
			err.Error(),
			"fallback reason",
		) {
		t.Fatalf(
			"error = %v, want fallback rejection",
			err,
		)
	}
}

func validHistoricalPayload(
	schedule verificationSchedule,
) response.SuccessResponse[dto.ProjectionIntelligenceResponse] {
	points := make(
		[]dto.ProjectionIntelligencePoint,
		0,
		6,
	)
	for index := 0; index < 6; index++ {
		points = append(
			points,
			dto.ProjectionIntelligencePoint{
				Sequence: index,
				ForecastTime: schedule.AsOfTime.Add(
					time.Duration(index+1) *
						30 * time.Second,
				),
				Position: dto.ProjectionIntelligencePosition{
					Latitude: 40.50 +
						float64(index)*0.012,
					Longitude: 50.00 +
						float64(index)*0.024,
				},
				Uncertainty: dto.ProjectionIntelligenceUncertainty{
					HorizontalRadiusM: 750 +
						float64(index)*50,
				},
				Confidence: dto.ProjectionIntelligenceConfidence{
					Score:   0.85,
					Level:   "high",
					Reasons: []dto.ProjectionIntelligenceConfidenceReason{},
				},
			},
		)
	}

	neighbors := make(
		[]dto.ProjectionIntelligenceNeighbor,
		0,
		4,
	)
	for index := 1; index < 5; index++ {
		neighbors = append(
			neighbors,
			dto.ProjectionIntelligenceNeighbor{
				TrajectoryID: verificationFlights[index].
					TrajectoryID,
				SimilarityScore:  0.99,
				SimilarityLevel:  "high",
				AnchorPointIndex: currentPointCount - 1,
				AnchorObservedAt: schedule.AsOfTime.Add(
					-time.Duration(index) *
						24 * time.Hour,
				),
				AnchorDistanceKM: 0.1,
				CandidateStartTime: schedule.AsOfTime.Add(
					-time.Duration(index)*
						24*time.Hour -
						8*time.Minute,
				),
				CandidateEndTime: schedule.AsOfTime.Add(
					-time.Duration(index) *
						24 * time.Hour,
				),
				CandidateAgeSeconds: int64(
					time.Duration(index) *
						24 * time.Hour /
						time.Second,
				),
				PrefixPointCount:       currentPointCount,
				ContinuationPointCount: 3,
				ContinuationEndTime: schedule.AsOfTime.Add(
					-time.Duration(index)*
						24*time.Hour +
						3*time.Minute,
				),
			},
		)
	}

	return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{
		Success: true,
		Data: dto.ProjectionIntelligenceResponse{
			Version: projectionproduction.Version,
			Strategy: string(
				projectionproduction.
					StrategyHistoricalNeighbor,
			),
			ArrivalStatus: string(
				projectionproduction.
					ArrivalStatusAttached,
			),
			Projection: dto.ProjectionIntelligenceProjection{
				SchemaVersion: "projection-intelligence-v1",
				Status:        "complete",
				TrajectoryID: verificationFlights[0].
					TrajectoryID,
				Method: dto.ProjectionIntelligenceMethod{
					Name:          projectioncontinuation.MethodName,
					Version:       projectioncontinuation.Version,
					DecisionClass: "experimental",
				},
				Horizon: dto.ProjectionIntelligenceHorizon{
					AsOfTime: schedule.AsOfTime,
					EndTime: schedule.AsOfTime.Add(
						verificationDuration,
					),
					StepSeconds: 30,
					DurationSeconds: int64(
						verificationDuration /
							time.Second,
					),
				},
				Points: points,
				Arrival: &dto.ProjectionIntelligenceArrivalEstimate{
					AirportICAOCode: "ZBBB",
					EarliestTime: schedule.AsOfTime.Add(
						2 * time.Minute,
					),
					EstimatedTime: schedule.AsOfTime.Add(
						3 * time.Minute,
					),
					LatestTime: schedule.AsOfTime.Add(
						4 * time.Minute,
					),
					Confidence: dto.ProjectionIntelligenceConfidence{
						Score:   0.8,
						Level:   "high",
						Reasons: []dto.ProjectionIntelligenceConfidenceReason{},
					},
					Limitations: []dto.ProjectionIntelligenceLimitation{},
				},
				Confidence: dto.ProjectionIntelligenceConfidence{
					Score:   0.85,
					Level:   "high",
					Reasons: []dto.ProjectionIntelligenceConfidenceReason{},
				},
				Limitations:  []dto.ProjectionIntelligenceLimitation{},
				Explanations: []dto.ProjectionIntelligenceExplanation{},
				ScopeGuard:   "research_only_not_for_operational_use",
				Provenance: dto.ProjectionIntelligenceProvenance{
					InputFingerprint: "sha256:" +
						strings.Repeat(
							"a",
							64,
						),
					Inputs:                []dto.ProjectionIntelligenceInputReference{},
					LatestInputObservedAt: schedule.AsOfTime,
				},
				GeneratedAt: schedule.GeneratedAt,
			},
			Evidence: dto.ProjectionIntelligenceEvidence{
				NeighborSelection: &dto.ProjectionIntelligenceNeighborSelection{
					Version: "projection-historical-neighbor-selection-v1",
					Status:  "complete",
					CurrentTrajectoryID: verificationFlights[0].
						TrajectoryID,
					AsOfTime:                    schedule.AsOfTime,
					RequiredContinuationSeconds: 180,
					InputCandidateCount:         4,
					CheckedCandidateCount:       4,
					QualifiedCandidateCount:     4,
					SelectionLimit:              5,
					Neighbors:                   neighbors,
					Limitations:                 []dto.ProjectionIntelligenceNotice{},
					InputFingerprint: "sha256:" +
						strings.Repeat(
							"b",
							64,
						),
				},
				PatternConfidence: &dto.ProjectionIntelligencePatternConfidence{
					Version:                 "projection-pattern-confidence-v1",
					Status:                  "complete",
					Usable:                  true,
					NeighborCount:           4,
					TargetNeighborCount:     5,
					MeanSimilarityScore:     0.99,
					MeanCandidateAgeSeconds: 216000,
					MeanAnchorDistanceKM:    0.1,
					Score:                   0.9,
					Level:                   "high",
					Components:              []dto.ProjectionIntelligenceScoreComponent{},
					SelectedTrajectoryIDs: []string{
						verificationFlights[1].TrajectoryID,
						verificationFlights[2].TrajectoryID,
						verificationFlights[3].TrajectoryID,
						verificationFlights[4].TrajectoryID,
					},
					Limitations: []dto.ProjectionIntelligenceNotice{},
					InputFingerprint: "sha256:" +
						strings.Repeat(
							"c",
							64,
						),
				},
				Freshness: &dto.ProjectionIntelligenceFreshness{
					Version:                  "projection-pattern-freshness-guard-v1",
					Decision:                 "allowed",
					Usable:                   true,
					AsOfTime:                 schedule.AsOfTime,
					NeighborCount:            4,
					RecentNeighborCount:      4,
					NewestNeighborAgeSeconds: 86400,
					MeanNeighborAgeSeconds:   216000,
					OldestNeighborAgeSeconds: 345600,
					Score:                    0.9,
					Components:               []dto.ProjectionIntelligenceScoreComponent{},
					SelectedTrajectoryIDs: []string{
						verificationFlights[1].TrajectoryID,
						verificationFlights[2].TrajectoryID,
						verificationFlights[3].TrajectoryID,
						verificationFlights[4].TrajectoryID,
					},
					Limitations: []dto.ProjectionIntelligenceNotice{},
					InputFingerprint: "sha256:" +
						strings.Repeat(
							"d",
							64,
						),
				},
				RouteFrequency: &dto.ProjectionIntelligenceRouteFrequency{
					Version:                     "projection-low-frequency-route-guard-v1",
					Decision:                    "allowed",
					Usable:                      true,
					RouteKey:                    "ZAAA>ZBBB",
					AsOfTime:                    schedule.AsOfTime,
					ObservationCount:            5,
					DistinctFlightCount:         5,
					DistinctDayCount:            5,
					RecentObservationCount:      5,
					LatestObservationAgeSeconds: 0,
					RouteConfidenceScore:        0.95,
					Score:                       0.8,
					Components:                  []dto.ProjectionIntelligenceScoreComponent{},
					Limitations:                 []dto.ProjectionIntelligenceNotice{},
					HistoryInputFingerprint: "sha256:" +
						strings.Repeat(
							"e",
							64,
						),
					InputFingerprint: "sha256:" +
						strings.Repeat(
							"f",
							64,
						),
				},
			},
			Notices: []dto.ProjectionIntelligenceNotice{
				{
					Code:    "historical_neighbor_continuation_authorized",
					Message: "Historical continuation authorized.",
				},
			},
			InputFingerprint: "sha256:" +
				strings.Repeat(
					"1",
					64,
				),
			GeneratedAt: schedule.GeneratedAt,
		},
	}
}

func absolute(
	value float64,
) float64 {
	if value < 0 {
		return -value
	}

	return value
}
