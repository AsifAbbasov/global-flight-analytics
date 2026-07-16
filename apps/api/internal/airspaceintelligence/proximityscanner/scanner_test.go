package proximityscanner

import (
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactionradius"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
)

func TestScanBuildsCandidatesAndInteractionGraph(t *testing.T) {
	request := scanRequest(t)
	result, err := Scan(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if result.Status != ResultStatusComplete {
		t.Fatalf("Status = %q, want %q", result.Status, ResultStatusComplete)
	}
	if result.Metrics.AircraftCount != 3 ||
		result.Metrics.PossiblePairCount != 3 ||
		result.Metrics.EvaluatedPairCount != 3 ||
		result.Metrics.CandidatePairCount != 1 ||
		result.Metrics.HorizontalRejectedPairCount != 2 {
		t.Fatalf("Metrics = %+v", result.Metrics)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("Candidates = %d, want 1", len(result.Candidates))
	}
	candidate := result.Candidates[0]
	if candidate.ID != "trajectory:a--trajectory:b" {
		t.Fatalf("Candidate.ID = %q", candidate.ID)
	}
	if candidate.Kind != interactiongraph.InteractionKindConverging {
		t.Fatalf("Candidate.Kind = %q, want converging", candidate.Kind)
	}
	if candidate.Status != CandidateStatusComplete || !candidate.VerticalFilteringApplied {
		t.Fatalf("Candidate status = %q, vertical=%v", candidate.Status, candidate.VerticalFilteringApplied)
	}
	if result.Graph.Metrics.NodeCount != 3 || result.Graph.Metrics.EdgeCount != 1 {
		t.Fatalf("Graph metrics = %+v", result.Graph.Metrics)
	}
	if result.Graph.Edges[0].ID != candidate.ID {
		t.Fatalf("graph edge ID = %q, candidate ID = %q", result.Graph.Edges[0].ID, candidate.ID)
	}
	if result.Provenance.InputFingerprint == "" {
		t.Fatal("InputFingerprint is empty")
	}
	if report := Validate(result, DefaultPolicy()); report.Status != ValidationStatusValid {
		t.Fatalf("Validate() = %+v", report)
	}
}

func TestScanWithholdsVerticalFilteringAndLimitsCandidate(t *testing.T) {
	request := scanRequest(t)
	request.Scene.Aircraft = request.Scene.Aircraft[:2]
	request.Scene.Aircraft[1].AltitudeMeters = nil
	request.Scene.Aircraft[1].AltitudeReference = interactiongraph.AltitudeReferenceUnknown
	request.Scene.Aircraft[1].RadiusDecision.VerticalFilteringPermitted = false
	request.Scene.Aircraft[1].RadiusDecision.VerticalRadiusMeters = interactionradius.DefaultPolicy().MaximumVerticalRadiusMeters
	request.Scene.Aircraft[1].RadiusDecision.Status = interactionradius.DecisionStatusLimited
	request.Scene.Status = localtrafficscene.ResultStatusComplete
	request.Scene.Metrics.IncludedAircraftCount = 2
	request.Scene.Metrics.AllowedAircraftCount = 1
	request.Scene.Metrics.LimitedAircraftCount = 1

	result, err := Scan(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("Candidates = %d, want 1", len(result.Candidates))
	}
	candidate := result.Candidates[0]
	if candidate.Status != CandidateStatusLimited || candidate.VerticalFilteringApplied {
		t.Fatalf("Candidate = %+v", candidate)
	}
	if candidate.VerticalSeparationMeters != nil || candidate.EffectiveVerticalRadiusMeters != nil {
		t.Fatal("withheld vertical values should be nil")
	}
	if result.Status != ResultStatusLimited {
		t.Fatalf("Status = %q, want limited", result.Status)
	}
}

func TestScanRejectsTemporalAndVerticalPairs(t *testing.T) {
	request := scanRequest(t)
	request.Scene.Aircraft = request.Scene.Aircraft[:2]
	request.Scene.Aircraft[0].ObservedAt = request.Scene.AsOfTime.Add(-50 * time.Second)
	request.Scene.Aircraft[1].ObservedAt = request.Scene.AsOfTime.Add(-10 * time.Second)

	result, err := Scan(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("Scan() temporal error = %v", err)
	}
	if result.Metrics.TemporalRejectedPairCount != 1 || len(result.Candidates) != 0 {
		t.Fatalf("temporal metrics = %+v", result.Metrics)
	}

	request = scanRequest(t)
	request.Scene.Aircraft = request.Scene.Aircraft[:2]
	altitude := 6000.0
	request.Scene.Aircraft[1].AltitudeMeters = &altitude
	result, err = Scan(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("Scan() vertical error = %v", err)
	}
	if result.Metrics.VerticalRejectedPairCount != 1 || len(result.Candidates) != 0 {
		t.Fatalf("vertical metrics = %+v", result.Metrics)
	}
}

func TestScanFingerprintIsDeterministicForReorderedScene(t *testing.T) {
	request := scanRequest(t)
	first, err := Scan(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	reordered := scanRequest(t)
	reordered.Scene.Aircraft[0], reordered.Scene.Aircraft[2] =
		reordered.Scene.Aircraft[2], reordered.Scene.Aircraft[0]
	second, err := Scan(reordered, DefaultPolicy())
	if err != nil {
		t.Fatalf("Scan(reordered) error = %v", err)
	}
	if first.Provenance.InputFingerprint != second.Provenance.InputFingerprint {
		t.Fatalf("fingerprints differ: %q != %q", first.Provenance.InputFingerprint, second.Provenance.InputFingerprint)
	}
}

func TestHorizontalDistanceKilometers(t *testing.T) {
	distance := horizontalDistanceKilometers(40.4093, 49.8671, 40.4093, 49.8790)
	if math.Abs(distance-1.008) > 0.03 {
		t.Fatalf("distance = %.6f km", distance)
	}
}

func scanRequest(t *testing.T) Request {
	t.Helper()
	asOfTime := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	altitudeA := 3000.0
	altitudeB := 3250.0
	altitudeC := 3100.0
	sceneRequest := localtrafficscene.Request{
		RegionCode: "AZ-BAKU",
		RegionBounds: localtrafficscene.Bounds{
			MinimumLatitude:  39,
			MaximumLatitude:  42,
			MinimumLongitude: 48,
			MaximumLongitude: 51,
		},
		AsOfTime:    asOfTime,
		GeneratedAt: asOfTime.Add(time.Second),
		Observations: []localtrafficscene.ObservationInput{
			{
				ID: "trajectory:a", TrajectoryID: "a", ICAO24: "A1B2C3", Callsign: "GFA101",
				Latitude: 40.4093, Longitude: 49.8671, AltitudeMeters: &altitudeA,
				AltitudeReference:       interactiongraph.AltitudeReferenceBarometric,
				VelocityMetersPerSecond: 120, HeadingDegrees: 90, VerticalRateMetersPerSecond: 0,
				ObservedAt: asOfTime.Add(-10 * time.Second), SourceName: "fixture", QualityScore: 0.90,
			},
			{
				ID: "trajectory:b", TrajectoryID: "b", ICAO24: "D4E5F6", Callsign: "GFA202",
				Latitude: 40.4093, Longitude: 49.8790, AltitudeMeters: &altitudeB,
				AltitudeReference:       interactiongraph.AltitudeReferenceBarometric,
				VelocityMetersPerSecond: 120, HeadingDegrees: 270, VerticalRateMetersPerSecond: 0,
				ObservedAt: asOfTime.Add(-12 * time.Second), SourceName: "fixture", QualityScore: 0.88,
			},
			{
				ID: "trajectory:c", TrajectoryID: "c", ICAO24: "112233", Callsign: "GFA303",
				Latitude: 41.0, Longitude: 50.8, AltitudeMeters: &altitudeC,
				AltitudeReference:       interactiongraph.AltitudeReferenceBarometric,
				VelocityMetersPerSecond: 100, HeadingDegrees: 45, VerticalRateMetersPerSecond: 0,
				ObservedAt: asOfTime.Add(-15 * time.Second), SourceName: "fixture", QualityScore: 0.92,
			},
		},
	}
	scene, err := localtrafficscene.Build(
		sceneRequest,
		localtrafficscene.DefaultPolicy(),
		interactionradius.DefaultPolicy(),
	)
	if err != nil {
		t.Fatalf("localtrafficscene.Build() error = %v", err)
	}
	return Request{Scene: scene, GeneratedAt: asOfTime.Add(2 * time.Second)}
}
