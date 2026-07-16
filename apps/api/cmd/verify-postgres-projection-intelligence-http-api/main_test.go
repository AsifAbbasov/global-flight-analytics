package main

import (
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
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
		789,
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

	expectedGeneratedAt :=
		time.Date(
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
	if len(schedule.PointTimes) !=
		verificationPointCount {
		t.Fatalf(
			"PointTimes length = %d, want %d",
			len(schedule.PointTimes),
			verificationPointCount,
		)
	}
	for index := 1; index <
		len(schedule.PointTimes); index++ {
		previousIndex := index - 1
		if schedule.PointTimes[index].
			Sub(
				schedule.PointTimes[previousIndex],
			) != time.Minute {
			t.Fatalf(
				"point interval at index %d is not one minute",
				index,
			)
		}
	}
	lastPointIndex :=
		len(schedule.PointTimes) - 1
	if !schedule.PointTimes[lastPointIndex].
		Equal(schedule.AsOfTime) {
		t.Fatal(
			"last point does not equal the analytical time",
		)
	}
}

func TestBuildVerificationScheduleRejectsMissingClock(
	t *testing.T,
) {
	_, err :=
		buildVerificationSchedule(
			time.Time{},
		)
	if err == nil {
		t.Fatal(
			"expected missing verification clock to be rejected",
		)
	}
}

func TestProjectionRequestURLUsesExplicitQueryInputs(
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

	requestURL :=
		projectionRequestURL(
			verificationTrajectoryID,
			asOfTime,
			verificationDuration,
		)

	for _, fragment := range []string{
		"/api/v1/trajectories/" +
			verificationTrajectoryID +
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

func TestValidateSuccessPayloadAcceptsKinematicRuntimeContract(
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
		validRuntimeSuccessPayload(
			schedule,
		)

	if err := validateSuccessPayload(
		payload,
		schedule,
	); err != nil {
		t.Fatalf(
			"validateSuccessPayload() error = %v",
			err,
		)
	}
}

func TestValidateSuccessPayloadRejectsMissingFallbackNotice(
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
		validRuntimeSuccessPayload(
			schedule,
		)
	payload.Data.Notices =
		[]dto.ProjectionIntelligenceNotice{}

	err = validateSuccessPayload(
		payload,
		schedule,
	)
	if err == nil ||
		!strings.Contains(
			err.Error(),
			"fallback notice",
		) {
		t.Fatalf(
			"error = %v, want missing fallback notice",
			err,
		)
	}
}

func TestHasNotice(
	t *testing.T,
) {
	items := []dto.ProjectionIntelligenceNotice{
		{
			Code: "first",
		},
		{
			Code: "historical_neighbors_unavailable",
		},
	}

	if !hasNotice(
		items,
		"historical_neighbors_unavailable",
	) {
		t.Fatal(
			"expected notice to be found",
		)
	}
	if hasNotice(
		items,
		"missing",
	) {
		t.Fatal(
			"unexpected notice was found",
		)
	}
}

func validRuntimeSuccessPayload(
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
					time.Duration(
						index+1,
					) * 30 * time.Second,
				),
				Position: dto.ProjectionIntelligencePosition{
					Latitude: 40.50 +
						float64(index)*0.01,
					Longitude: 50.10 +
						float64(index)*0.01,
				},
				Uncertainty: dto.ProjectionIntelligenceUncertainty{
					HorizontalRadiusM: 500 +
						float64(index)*100,
				},
				Confidence: dto.ProjectionIntelligenceConfidence{
					Score:   0.80,
					Level:   "high",
					Reasons: []dto.ProjectionIntelligenceConfidenceReason{},
				},
			},
		)
	}

	return response.SuccessResponse[dto.ProjectionIntelligenceResponse]{
		Success: true,
		Data: dto.ProjectionIntelligenceResponse{
			Version: projectionproduction.Version,
			Strategy: string(
				projectionproduction.
					StrategyKinematic,
			),
			FallbackReason: "historical_neighbors_unavailable",
			ArrivalStatus: string(
				projectionproduction.
					ArrivalStatusWithheld,
			),
			Projection: dto.ProjectionIntelligenceProjection{
				SchemaVersion: "projection-intelligence-v1",
				Status:        "complete",
				TrajectoryID:  verificationTrajectoryID,
				ICAO24:        verificationICAO24,
				Callsign:      verificationCallsign,
				Method: dto.ProjectionIntelligenceMethod{
					Name:          projectionbaseline.MethodName,
					Version:       projectionbaseline.Version,
					DecisionClass: "physics_derived",
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
				Confidence: dto.ProjectionIntelligenceConfidence{
					Score:   0.80,
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
					Version:             "projection-historical-neighbor-selection-v1",
					Status:              "unavailable",
					CurrentTrajectoryID: verificationTrajectoryID,
					AsOfTime:            schedule.AsOfTime,
					RequiredContinuationSeconds: int64(
						verificationDuration /
							time.Second,
					),
					Neighbors:   []dto.ProjectionIntelligenceNeighbor{},
					Limitations: []dto.ProjectionIntelligenceNotice{},
					InputFingerprint: "sha256:" +
						strings.Repeat(
							"b",
							64,
						),
				},
			},
			Notices: []dto.ProjectionIntelligenceNotice{
				{
					Code:    "historical_neighbors_unavailable",
					Message: "Kinematic fallback selected.",
				},
			},
			InputFingerprint: "sha256:" +
				strings.Repeat(
					"c",
					64,
				),
			GeneratedAt: schedule.GeneratedAt,
		},
	}
}
