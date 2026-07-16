package localtrafficscene

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactionradius"
)

func TestBuildCompleteLocalTrafficScene(t *testing.T) {
	request := validRequest()
	request.Observations = []ObservationInput{
		validObservation("trajectory:a", "A1B2C3", 40.40, 49.80, 10*time.Second),
		validObservation("trajectory:b", "D4E5F6", 40.45, 49.85, 15*time.Second),
		groundObservation("trajectory:ground"),
		outsideObservation("trajectory:outside"),
	}

	result, err := Build(request, DefaultPolicy(), interactionradius.DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status != ResultStatusComplete {
		t.Fatalf("Status = %q, want %q", result.Status, ResultStatusComplete)
	}
	if result.Metrics.IncludedAircraftCount != 2 ||
		result.Metrics.GroundExcludedCount != 1 ||
		result.Metrics.OutsideRegionExcludedCount != 1 {
		t.Fatalf("Metrics = %+v", result.Metrics)
	}
	if len(result.GraphNodeInputs()) != 2 {
		t.Fatalf("GraphNodeInputs length = %d", len(result.GraphNodeInputs()))
	}
	if result.ScopeGuard != ScopeGuardResearchOnly {
		t.Fatalf("ScopeGuard = %q", result.ScopeGuard)
	}
}

func TestBuildSelectsLatestDuplicateAndExcludesFutureEvidence(t *testing.T) {
	request := validRequest()
	older := validObservation("trajectory:a", "A1B2C3", 40.40, 49.80, 40*time.Second)
	newer := validObservation("trajectory:a", "A1B2C3", 40.41, 49.81, 10*time.Second)
	future := validObservation("trajectory:b", "D4E5F6", 40.45, 49.85, -10*time.Second)
	request.Observations = []ObservationInput{older, future, newer}

	result, err := Build(request, DefaultPolicy(), interactionradius.DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(result.Aircraft) != 1 || result.Aircraft[0].Latitude != newer.Latitude {
		t.Fatalf("Aircraft = %+v", result.Aircraft)
	}
	if result.Metrics.DuplicateExcludedCount != 1 ||
		result.Metrics.FutureEvidenceExcludedCount != 1 ||
		result.Status != ResultStatusLimited {
		t.Fatalf("Metrics = %+v, Status = %q", result.Metrics, result.Status)
	}
}

func TestBuildIncludesLimitedUnknownAltitudeDecision(t *testing.T) {
	request := validRequest()
	observation := validObservation("trajectory:a", "A1B2C3", 40.40, 49.80, 10*time.Second)
	observation.AltitudeMeters = nil
	observation.AltitudeReference = interactiongraph.AltitudeReferenceUnknown
	request.Observations = []ObservationInput{observation}

	result, err := Build(request, DefaultPolicy(), interactionradius.DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(result.Aircraft) != 1 ||
		result.Aircraft[0].RadiusDecision.Status != interactionradius.DecisionStatusLimited ||
		result.Metrics.LimitedAircraftCount != 1 {
		t.Fatalf("Result = %+v", result)
	}
}

func TestBuildExcludesRadiusPolicyBlockedObservation(t *testing.T) {
	request := validRequest()
	stale := validObservation("trajectory:a", "A1B2C3", 40.40, 49.80, 2*time.Minute)
	request.Observations = []ObservationInput{stale}

	result, err := Build(request, DefaultPolicy(), interactionradius.DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status != ResultStatusUnavailable ||
		result.Metrics.RadiusPolicyBlockedCount != 1 ||
		len(result.Aircraft) != 0 {
		t.Fatalf("Result = %+v", result)
	}
}

func TestBuildIsDeterministicAcrossInputOrder(t *testing.T) {
	request := validRequest()
	one := validObservation("trajectory:a", "A1B2C3", 40.40, 49.80, 10*time.Second)
	two := validObservation("trajectory:b", "D4E5F6", 40.45, 49.85, 15*time.Second)
	request.Observations = []ObservationInput{one, two}
	first, err := Build(request, DefaultPolicy(), interactionradius.DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	request.Observations = []ObservationInput{two, one}
	second, err := Build(request, DefaultPolicy(), interactionradius.DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if first.Provenance.InputFingerprint != second.Provenance.InputFingerprint {
		t.Fatalf("fingerprints differ: %s != %s", first.Provenance.InputFingerprint, second.Provenance.InputFingerprint)
	}
	if first.Aircraft[0].NodeID != "trajectory:a" || second.Aircraft[0].NodeID != "trajectory:a" {
		t.Fatalf("aircraft order is not deterministic")
	}
}

func TestBuildRejectsInvalidBounds(t *testing.T) {
	request := validRequest()
	request.RegionBounds.MinimumLatitude = request.RegionBounds.MaximumLatitude
	_, err := Build(request, DefaultPolicy(), interactionradius.DefaultPolicy())
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("Build() error = %v, want ErrInvalidRequest", err)
	}
}

func TestResultCloneIsDeep(t *testing.T) {
	request := validRequest()
	request.Observations = []ObservationInput{
		validObservation("trajectory:a", "A1B2C3", 40.40, 49.80, 10*time.Second),
	}
	result, err := Build(request, DefaultPolicy(), interactionradius.DefaultPolicy())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	cloned := result.Clone()
	*cloned.Aircraft[0].AltitudeMeters = 999
	cloned.Aircraft[0].RadiusDecision.Limitations[0].Code = "changed"
	if *result.Aircraft[0].AltitudeMeters == 999 ||
		result.Aircraft[0].RadiusDecision.Limitations[0].Code == "changed" {
		t.Fatal("Clone() did not deep-copy nested values")
	}
}

func validRequest() Request {
	asOf := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	return Request{
		RegionCode: "az-test",
		RegionBounds: Bounds{
			MinimumLatitude:  39,
			MaximumLatitude:  42,
			MinimumLongitude: 47,
			MaximumLongitude: 51,
		},
		AsOfTime:    asOf,
		GeneratedAt: asOf.Add(5 * time.Second),
	}
}

func validObservation(
	id string,
	icao24 string,
	latitude float64,
	longitude float64,
	age time.Duration,
) ObservationInput {
	asOf := validRequest().AsOfTime
	altitude := 10000.0
	return ObservationInput{
		ID:                          id,
		TrajectoryID:                id[len("trajectory:"):],
		ICAO24:                      icao24,
		Callsign:                    "test123",
		Latitude:                    latitude,
		Longitude:                   longitude,
		AltitudeMeters:              &altitude,
		AltitudeReference:           interactiongraph.AltitudeReferenceBarometric,
		VelocityMetersPerSecond:     220,
		HeadingDegrees:              90,
		VerticalRateMetersPerSecond: 0,
		ObservedAt:                  asOf.Add(-age),
		SourceName:                  "fixture",
		QualityScore:                0.90,
	}
}

func groundObservation(id string) ObservationInput {
	observation := validObservation(id, "112233", 40.42, 49.82, 10*time.Second)
	observation.OnGround = true
	return observation
}

func outsideObservation(id string) ObservationInput {
	return validObservation(id, "445566", 45, 55, 10*time.Second)
}
