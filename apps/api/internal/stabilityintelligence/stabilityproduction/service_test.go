package stabilityproduction

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionbaseline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/scopeenforcement"
)

const testTrajectoryID = "2e0dc3a0-4c5e-4bda-a5ad-5a14de916a41"

type testProjectionReader struct {
	generatedAt time.Time
}

func (
	reader testProjectionReader,
) ReadProjection(
	_ context.Context,
	request ProjectionRequest,
) (projectionproduction.Result, error) {
	return testProductionProjection(
		request.TrajectoryID,
		request.AsOfTime,
		request.RequestedDuration,
		reader.generatedAt,
	), nil
}

func TestServiceComposesProductionStabilityIntelligence(
	t *testing.T,
) {
	generatedAt := time.Date(
		2035,
		time.January,
		15,
		12,
		10,
		0,
		0,
		time.UTC,
	)
	service, err := New(
		Config{
			ProjectionReader: testProjectionReader{
				generatedAt: generatedAt,
			},
			Now: func() time.Time {
				return generatedAt.Add(
					time.Second,
				)
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	firstAsOf := generatedAt.Add(
		-10 * time.Minute,
	)
	request := Request{
		TrajectoryID: testTrajectoryID,
		AsOfTimes: []time.Time{
			firstAsOf,
			firstAsOf.Add(30 * time.Second),
			firstAsOf.Add(time.Minute),
		},
		RequestedDuration: 5 * time.Minute,
	}

	result, err := service.Get(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := result.Validate(); err != nil {
		t.Fatalf("result validation: %v", err)
	}
	if len(result.ForecastVersions) != 3 ||
		len(result.Transitions) != 2 ||
		result.ForecastAnalysis.Metrics.
			VersionCount != 3 {
		t.Fatalf(
			"unexpected production counts: %#v",
			result,
		)
	}
	if result.PropagatedConfidence.Score <= 0 ||
		result.FailureExplanation.PrimaryCode == "" ||
		result.UnknownIntervention.Decision == "" ||
		result.ScopeEnforcement.Decision ==
			scopeenforcement.DecisionBlocked {
		t.Fatalf(
			"production safety composition is incomplete: %#v",
			result,
		)
	}

	replayed, err := service.Get(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.InputFingerprint !=
		result.InputFingerprint {
		t.Fatalf(
			"deterministic replay fingerprint mismatch: %s != %s",
			replayed.InputFingerprint,
			result.InputFingerprint,
		)
	}
}

func TestServiceRejectsInvalidAsOfSequence(
	t *testing.T,
) {
	now := time.Date(
		2035,
		time.January,
		15,
		12,
		10,
		0,
		0,
		time.UTC,
	)
	service, err := New(
		Config{
			ProjectionReader: testProjectionReader{
				generatedAt: now,
			},
			Now: func() time.Time {
				return now
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Get(
		context.Background(),
		Request{
			TrajectoryID: testTrajectoryID,
			AsOfTimes: []time.Time{
				now.Add(-time.Minute),
				now.Add(-time.Minute),
			},
			RequestedDuration: 5 * time.Minute,
		},
	)
	if err == nil {
		t.Fatal(
			"duplicate as-of time was accepted",
		)
	}
}

func testProductionProjection(
	trajectoryID string,
	asOfTime time.Time,
	duration time.Duration,
	generatedAt time.Time,
) projectionproduction.Result {
	step := 30 * time.Second
	pointCount := int(duration / step)
	altitude := 9000.0
	vertical := 250.0
	points := make(
		[]projectioncontract.ProjectionPoint,
		0,
		pointCount,
	)
	anchor := time.Date(
		2035,
		time.January,
		15,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	for index := 0; index < pointCount; index++ {
		forecastTime := asOfTime.Add(
			time.Duration(index+1) * step,
		)
		progressMinutes :=
			forecastTime.Sub(anchor).Minutes()
		points = append(
			points,
			projectioncontract.ProjectionPoint{
				Sequence:     index,
				ForecastTime: forecastTime,
				Position: projectioncontract.Position{
					Latitude:  40.40,
					Longitude: 49.80 + progressMinutes*0.001,
					AltitudeM: float64Pointer(altitude),
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 1000 +
						float64(index)*100,
					VerticalRadiusM: float64Pointer(
						vertical,
					),
				},
				Confidence: projectioncontract.Confidence{
					Score: 0.80 -
						float64(index)*0.01,
					Level: projectioncontract.
						ConfidenceLevelMedium,
					Reasons: []projectioncontract.
						ConfidenceReason{
						{
							Code:         "bounded_horizon",
							Message:      "Bounded short-horizon confidence.",
							Contribution: 0.80,
						},
					},
				},
			},
		)
	}

	inputFingerprint :=
		testFingerprint(
			asOfTime.Format(
				time.RFC3339Nano,
			),
		)
	projection := projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        projectioncontract.ResultStatusComplete,
		TrajectoryID:  trajectoryID,
		ICAO24:        "A1B2C3",
		Callsign:      "GFA1204",
		Method: projectioncontract.Method{
			Name:    projectionbaseline.MethodName,
			Version: projectionbaseline.Version,
			DecisionClass: projectioncontract.
				DecisionClassPhysicsDerived,
		},
		Horizon: projectioncontract.Horizon{
			AsOfTime: asOfTime,
			EndTime:  asOfTime.Add(duration),
			Step:     step,
		},
		Points: points,
		Confidence: projectioncontract.Confidence{
			Score: 0.78,
			Level: projectioncontract.
				ConfidenceLevelMedium,
			Reasons: []projectioncontract.
				ConfidenceReason{
				{
					Code:         "projection_method",
					Message:      "Kinematic baseline confidence.",
					Contribution: 0.78,
				},
			},
		},
		Limitations: []projectioncontract.Limitation{
			{
				Code:    "research_only",
				Message: "Research only.",
				Scope:   "operational_use",
			},
			{
				Code:    "short_horizon",
				Message: "Short horizon only.",
				Scope:   "horizon",
			},
		},
		Explanations: []projectioncontract.Explanation{
			{
				Code:    "method",
				Message: "Kinematic baseline.",
			},
		},
		ScopeGuard: projectioncontract.ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: inputFingerprint,
			Inputs: []projectioncontract.InputReference{
				{
					Name: "current_trajectory",
					Classification: projectioncontract.
						InputClassificationObserved,
					ObservedAt: asOfTime.Add(
						-5 * time.Second,
					),
				},
			},
			LatestInputObservedAt: asOfTime.Add(
				-5 * time.Second,
			),
		},
		GeneratedAt: generatedAt,
	}

	return projectionproduction.Result{
		Version:        projectionproduction.Version,
		Strategy:       projectionproduction.StrategyKinematic,
		FallbackReason: "historical_neighbors_unavailable",
		ArrivalStatus:  projectionproduction.ArrivalStatusWithheld,
		Projection:     projection,
		Notices: []projectionproduction.Notice{
			{
				Code:    "kinematic_fallback",
				Message: "Historical evidence is unavailable.",
			},
		},
		InputFingerprint: testFingerprint(
			"production:" +
				asOfTime.Format(
					time.RFC3339Nano,
				),
		),
		GeneratedAt: generatedAt,
	}
}

func float64Pointer(
	value float64,
) *float64 {
	return &value
}

func testFingerprint(
	value string,
) string {
	digest := sha256.Sum256(
		[]byte(value),
	)
	return "sha256:" +
		hex.EncodeToString(digest[:])
}
