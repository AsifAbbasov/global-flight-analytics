package forecastanalysis

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/forecaststability"
)

func TestAnalyzeForecastHistoryStableAndDeterministic(t *testing.T) {
	versions := testVersions(
		t,
		[]float64{0, 0.002, 0.004, 0.006},
	)
	evaluatedAt := versions[len(versions)-1].CreatedAt.Add(time.Second)

	left, err := AnalyzeForecastHistory(
		Request{
			Versions:    versions,
			EvaluatedAt: evaluatedAt,
		},
		DefaultPolicy(),
		forecaststability.DefaultStabilityPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}

	reversed := append(
		[]forecaststability.ForecastVersionRecord(nil),
		versions...,
	)
	for leftIndex, rightIndex := 0, len(reversed)-1; leftIndex < rightIndex; leftIndex, rightIndex = leftIndex+1, rightIndex-1 {
		reversed[leftIndex], reversed[rightIndex] =
			reversed[rightIndex], reversed[leftIndex]
	}

	right, err := AnalyzeForecastHistory(
		Request{
			Versions:    reversed,
			EvaluatedAt: evaluatedAt,
		},
		DefaultPolicy(),
		forecaststability.DefaultStabilityPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if left.Health != HealthStable ||
		left.Metrics.StableTransitionShare != 1 {
		t.Fatalf("result = %#v", left)
	}
	if left.Provenance.InputFingerprint !=
		right.Provenance.InputFingerprint {
		t.Fatal("input order changed fingerprint")
	}
	if err := ValidateResult(left, DefaultPolicy()); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzeForecastHistoryRejectsBrokenChain(t *testing.T) {
	versions := testVersions(t, []float64{0, 0.002})
	versions[1].ParentVersionID = "broken"
	_, err := AnalyzeForecastHistory(
		Request{
			Versions: versions,
			EvaluatedAt: versions[1].CreatedAt.Add(
				time.Second,
			),
		},
		DefaultPolicy(),
		forecaststability.DefaultStabilityPolicy(),
	)
	if err == nil {
		t.Fatal("broken chain was accepted")
	}
}

func TestValidateResultRejectsTamperedFingerprint(t *testing.T) {
	versions := testVersions(t, []float64{0, 0.002})
	result, err := AnalyzeForecastHistory(
		Request{
			Versions:    versions,
			EvaluatedAt: versions[1].CreatedAt.Add(time.Second),
		},
		DefaultPolicy(),
		forecaststability.DefaultStabilityPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}
	result.Provenance.InputFingerprint = testHash("tampered")
	if err := ValidateResult(result, DefaultPolicy()); err == nil {
		t.Fatal("tampered result was accepted")
	}
}

func testVersions(
	t *testing.T,
	shifts []float64,
) []forecaststability.ForecastVersionRecord {
	t.Helper()
	result := make(
		[]forecaststability.ForecastVersionRecord,
		0,
		len(shifts),
	)
	var previous *forecaststability.ForecastVersionRecord

	for index, shift := range shifts {
		projection := testProjection()
		for pointIndex := range projection.Points {
			projection.Points[pointIndex].Position.Longitude += shift
		}
		projection.GeneratedAt = projection.GeneratedAt.Add(
			time.Duration(index) * time.Minute,
		)
		projection.Provenance.InputFingerprint = testHash(
			"input-" + time.Duration(index).String(),
		)

		registered, err := forecaststability.RegisterVersion(
			forecaststability.ForecastVersionRequest{
				Projection:            projection,
				PolicyVersion:         "projection-policy-v1",
				ImplementationVersion: "build-v1",
				Previous:              previous,
				RegisteredAt: projection.GeneratedAt.Add(
					time.Second,
				),
			},
			forecaststability.DefaultVersionPolicy(),
		)
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, registered.Record)
		previous = &result[len(result)-1]
	}
	return result
}

func testProjection() projectioncontract.Result {
	asOfTime := time.Date(
		2035,
		time.January,
		15,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	altitude := 9000.0
	verticalUncertainty := 200.0
	points := make([]projectioncontract.ProjectionPoint, 0, 4)
	for index := 0; index < 4; index++ {
		points = append(
			points,
			projectioncontract.ProjectionPoint{
				Sequence: index,
				ForecastTime: asOfTime.Add(
					time.Duration(index+1) * 30 * time.Second,
				),
				Position: projectioncontract.Position{
					Latitude:  40.4,
					Longitude: 49.8 + float64(index)*0.02,
					AltitudeM: &altitude,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 1000 + float64(index)*100,
					VerticalRadiusM:   &verticalUncertainty,
				},
				Confidence: projectioncontract.Confidence{
					Score: 0.8,
					Level: projectioncontract.
						ConfidenceLevelMedium,
					Reasons: []projectioncontract.ConfidenceReason{
						{
							Code:         "bounded",
							Message:      "Bounded horizon.",
							Contribution: 0.8,
						},
					},
				},
			},
		)
	}

	return projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        projectioncontract.ResultStatusComplete,
		TrajectoryID:  "trajectory-stage-12-analysis",
		Method: projectioncontract.Method{
			Name:          "short_horizon_kinematic_baseline",
			Version:       "v1",
			DecisionClass: projectioncontract.DecisionClassPhysicsDerived,
		},
		Horizon: projectioncontract.Horizon{
			AsOfTime: asOfTime,
			EndTime:  asOfTime.Add(2 * time.Minute),
			Step:     30 * time.Second,
		},
		Points: points,
		Confidence: projectioncontract.Confidence{
			Score: 0.78,
			Level: projectioncontract.
				ConfidenceLevelMedium,
			Reasons: []projectioncontract.ConfidenceReason{
				{
					Code:         "method",
					Message:      "Projection method confidence.",
					Contribution: 0.78,
				},
			},
		},
		Limitations: []projectioncontract.Limitation{
			{
				Code:    "research",
				Message: "Research only.",
				Scope:   "use",
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
			InputFingerprint: testHash("input"),
			Inputs: []projectioncontract.InputReference{
				{
					Name: "trajectory",
					Classification: projectioncontract.
						InputClassificationObserved,
					ObservedAt: asOfTime.Add(-time.Second),
				},
			},
			LatestInputObservedAt: asOfTime.Add(-time.Second),
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}
}

func testHash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(digest[:])
}
