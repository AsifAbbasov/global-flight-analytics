package separationrisk

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactionradius"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/proximityscanner"
)

func TestEvaluateClassifiesHighConvergingRisk(t *testing.T) {
	request := riskRequest(t)
	result, err := Evaluate(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Status != ResultStatusComplete {
		t.Fatalf("Status = %q, want complete", result.Status)
	}
	if len(result.Assessments) != 1 {
		t.Fatalf("Assessments = %d, want 1", len(result.Assessments))
	}
	assessment := result.Assessments[0]
	if assessment.Level != RiskLevelHigh || assessment.Status != AssessmentStatusComplete {
		t.Fatalf("Assessment = %+v", assessment)
	}
	if assessment.RiskScore == nil || *assessment.RiskScore < DefaultPolicy().HighRiskMinimumScore {
		t.Fatalf("RiskScore = %v", assessment.RiskScore)
	}
	if result.Metrics.HighCount != 1 || result.Metrics.HighestDeterminateRiskLevel != RiskLevelHigh {
		t.Fatalf("Metrics = %+v", result.Metrics)
	}
	if report := Validate(result, DefaultPolicy()); report.Status != ValidationStatusValid {
		t.Fatalf("Validate() = %+v", report)
	}
}

func TestEvaluateWithholdsRiskWithoutVerticalEvidence(t *testing.T) {
	request := riskRequest(t)
	candidate := request.Scan.Candidates[0]
	candidate.Status = proximityscanner.CandidateStatusLimited
	candidate.VerticalFilteringApplied = false
	candidate.VerticalSeparationMeters = nil
	candidate.EffectiveVerticalRadiusMeters = nil
	request.Scan.Candidates[0] = candidate
	request.Scan.Status = proximityscanner.ResultStatusLimited

	result, err := Evaluate(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assessment := result.Assessments[0]
	if assessment.Status != AssessmentStatusLimited || assessment.Level != RiskLevelIndeterminate {
		t.Fatalf("Assessment = %+v", assessment)
	}
	if assessment.RiskScore != nil || assessment.VerticalRadiusRatio != nil {
		t.Fatal("indeterminate assessment published determinate values")
	}
	if result.Status != ResultStatusLimited || result.Metrics.IndeterminateCount != 1 {
		t.Fatalf("Result = %+v", result)
	}
}

func TestEvaluateClassifiesDivergingAsContextual(t *testing.T) {
	request := riskRequest(t)
	candidate := request.Scan.Candidates[0]
	candidate.Kind = interactiongraph.InteractionKindDiverging
	candidate.ClosingRateMetersPerSecond = -25
	request.Scan.Candidates[0] = candidate

	result, err := Evaluate(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Assessments[0].Level != RiskLevelContextual {
		t.Fatalf("Level = %q, want contextual", result.Assessments[0].Level)
	}
}

func TestEvaluateFingerprintIsDeterministic(t *testing.T) {
	request := riskRequest(t)
	first, err := Evaluate(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	second, err := Evaluate(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("Evaluate() second error = %v", err)
	}
	if first.Provenance.InputFingerprint != second.Provenance.InputFingerprint {
		t.Fatalf("fingerprints differ: %q != %q", first.Provenance.InputFingerprint, second.Provenance.InputFingerprint)
	}
}

func riskRequest(t *testing.T) Request {
	t.Helper()
	asOfTime := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	altitudeA := 3000.0
	altitudeB := 3200.0
	scene, err := localtrafficscene.Build(
		localtrafficscene.Request{
			RegionCode: "AZ-BAKU",
			RegionBounds: localtrafficscene.Bounds{
				MinimumLatitude: 39, MaximumLatitude: 42,
				MinimumLongitude: 48, MaximumLongitude: 51,
			},
			AsOfTime: asOfTime, GeneratedAt: asOfTime.Add(time.Second),
			Observations: []localtrafficscene.ObservationInput{
				{ID: "trajectory:a", TrajectoryID: "a", ICAO24: "A1B2C3", Callsign: "GFA101", Latitude: 40.4093, Longitude: 49.8671, AltitudeMeters: &altitudeA, AltitudeReference: interactiongraph.AltitudeReferenceBarometric, VelocityMetersPerSecond: 120, HeadingDegrees: 90, ObservedAt: asOfTime.Add(-10 * time.Second), SourceName: "fixture", QualityScore: 0.95},
				{ID: "trajectory:b", TrajectoryID: "b", ICAO24: "D4E5F6", Callsign: "GFA202", Latitude: 40.4093, Longitude: 49.8790, AltitudeMeters: &altitudeB, AltitudeReference: interactiongraph.AltitudeReferenceBarometric, VelocityMetersPerSecond: 120, HeadingDegrees: 270, ObservedAt: asOfTime.Add(-12 * time.Second), SourceName: "fixture", QualityScore: 0.93},
			},
		},
		localtrafficscene.DefaultPolicy(),
		interactionradius.DefaultPolicy(),
	)
	if err != nil {
		t.Fatalf("localtrafficscene.Build() error = %v", err)
	}
	scan, err := proximityscanner.Scan(
		proximityscanner.Request{Scene: scene, GeneratedAt: asOfTime.Add(2 * time.Second)},
		proximityscanner.DefaultPolicy(),
	)
	if err != nil {
		t.Fatalf("proximityscanner.Scan() error = %v", err)
	}
	return Request{Scan: scan, GeneratedAt: asOfTime.Add(3 * time.Second)}
}
