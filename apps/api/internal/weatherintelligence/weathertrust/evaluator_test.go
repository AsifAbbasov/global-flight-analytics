package weathertrust

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
)

func TestDefaultPolicyIsValid(t *testing.T) {
	t.Parallel()
	if err := DefaultPolicy().Validate(); err != nil {
		t.Fatalf("default policy is invalid: %v", err)
	}
}

func TestEvaluateSurfaceOnlyWeatherIsLimited(t *testing.T) {
	t.Parallel()
	result, err := Evaluate(validSurfaceResult(), DefaultPolicy())
	if err != nil {
		t.Fatalf("evaluate weather trust: %v", err)
	}
	if result.Decision != DecisionLimited || !result.Usable {
		t.Fatalf("expected usable limited result, got %#v", result)
	}
	if len(result.AllowedScopes) != 1 || result.AllowedScopes[0] != UsageScopeSurfaceContext {
		t.Fatalf("unexpected allowed scopes %#v", result.AllowedScopes)
	}
	if !hasNotice(result.Limitations, "surface_weather_not_flight_level") {
		t.Fatalf("surface limitation is missing: %#v", result.Limitations)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("limited result is invalid: %v", err)
	}
}

func TestEvaluateFreshFlightLevelWeatherIsAllowed(t *testing.T) {
	t.Parallel()
	result, err := Evaluate(validFlightLevelResult(), DefaultPolicy())
	if err != nil {
		t.Fatalf("evaluate weather trust: %v", err)
	}
	if result.Decision != DecisionAllowed || !result.Usable {
		t.Fatalf("expected allowed result, got %#v", result)
	}
	for _, scope := range []UsageScope{
		UsageScopeProjectionUncertainty,
		UsageScopeTrajectoryContext,
	} {
		if !hasScope(result.AllowedScopes, scope) {
			t.Fatalf("expected scope %q in %#v", scope, result.AllowedScopes)
		}
	}
}

func TestEvaluateUnavailableWeatherIsBlocked(t *testing.T) {
	t.Parallel()
	result, err := Evaluate(validUnavailableResult(), DefaultPolicy())
	if err != nil {
		t.Fatalf("evaluate weather trust: %v", err)
	}
	if result.Decision != DecisionBlocked || result.Usable || len(result.AllowedScopes) != 0 {
		t.Fatalf("expected blocked unavailable result, got %#v", result)
	}
}

func TestEvaluateInvalidContractIsBlocked(t *testing.T) {
	t.Parallel()
	input := validSurfaceResult()
	input.ScopeGuard = "operational"
	result, err := Evaluate(input, DefaultPolicy())
	if err != nil {
		t.Fatalf("evaluate weather trust: %v", err)
	}
	if result.Decision != DecisionBlocked {
		t.Fatalf("expected invalid contract to be blocked, got %#v", result)
	}
	if !hasNotice(result.Limitations, "contract_scope_guard_invalid") {
		t.Fatalf("contract issue is missing: %#v", result.Limitations)
	}
}

func TestEvaluateStaleAnalysisIsBlocked(t *testing.T) {
	t.Parallel()
	input := validFlightLevelResult()
	input.Samples[0].ValidAt = input.AsOfTime.Add(-3 * time.Hour)
	result, err := Evaluate(input, DefaultPolicy())
	if err != nil {
		t.Fatalf("evaluate weather trust: %v", err)
	}
	if result.Decision != DecisionBlocked ||
		!hasNotice(result.Limitations, "weather_temporal_boundary_exceeded") {
		t.Fatalf("expected stale analysis block, got %#v", result)
	}
}

func TestEvaluateLowConfidenceIsBlocked(t *testing.T) {
	t.Parallel()
	input := validFlightLevelResult()
	input.Confidence.Score = 0.20
	input.Confidence.Level = weathercontract.ConfidenceLevelLow
	result, err := Evaluate(input, DefaultPolicy())
	if err != nil {
		t.Fatalf("evaluate weather trust: %v", err)
	}
	if result.Decision != DecisionBlocked ||
		!hasNotice(result.Limitations, "weather_confidence_below_usable_minimum") {
		t.Fatalf("expected low confidence block, got %#v", result)
	}
}

func TestEvaluateInsufficientFeaturesIsBlocked(t *testing.T) {
	t.Parallel()
	input := validFlightLevelResult()
	temperature := 15.0
	input.Samples[0].Features = weathercontract.FeatureVector{
		TemperatureCelsius: &temperature,
	}
	result, err := Evaluate(input, DefaultPolicy())
	if err != nil {
		t.Fatalf("evaluate weather trust: %v", err)
	}
	if result.Decision != DecisionBlocked ||
		!hasNotice(result.Limitations, "weather_features_insufficient") {
		t.Fatalf("expected feature block, got %#v", result)
	}
}

func TestEvaluateForecastLeadBoundary(t *testing.T) {
	t.Parallel()
	policy := DefaultPolicy()
	input := validFlightLevelResult()
	input.Samples[0].Source.EvidenceKind = weathercontract.EvidenceKindForecast
	input.Samples[0].AvailableAt = input.AsOfTime.Add(-time.Hour)
	input.Samples[0].RetrievedAt = input.AsOfTime.Add(-30 * time.Minute)
	input.Provenance.LatestAvailableAt = input.Samples[0].AvailableAt

	input.Samples[0].ValidAt = input.AsOfTime.Add(2 * time.Hour)
	supported, err := Evaluate(input, policy)
	if err != nil {
		t.Fatalf("evaluate supported forecast: %v", err)
	}
	if supported.Decision == DecisionBlocked {
		t.Fatalf("supported forecast lead was blocked: %#v", supported)
	}

	input.Samples[0].ValidAt = input.AsOfTime.Add(7 * time.Hour)
	blocked, err := Evaluate(input, policy)
	if err != nil {
		t.Fatalf("evaluate excessive forecast: %v", err)
	}
	if blocked.Decision != DecisionBlocked {
		t.Fatalf("excessive forecast lead was not blocked: %#v", blocked)
	}
}

func TestEvaluateFingerprintIsDeterministic(t *testing.T) {
	t.Parallel()
	input := validSurfaceResult()
	policy := DefaultPolicy()
	first, err := Evaluate(input, policy)
	if err != nil {
		t.Fatalf("evaluate first result: %v", err)
	}
	second, err := Evaluate(input, policy)
	if err != nil {
		t.Fatalf("evaluate second result: %v", err)
	}
	if first.InputFingerprint != second.InputFingerprint {
		t.Fatal("same trust input produced different fingerprints")
	}
	changedPolicy := policy
	changedPolicy.MinimumAllowedScore = 0.80
	changed, err := Evaluate(input, changedPolicy)
	if err != nil {
		t.Fatalf("evaluate changed policy: %v", err)
	}
	if changed.InputFingerprint == first.InputFingerprint {
		t.Fatal("changed policy did not change trust fingerprint")
	}
}

func TestEvaluateRejectsInvalidPolicy(t *testing.T) {
	t.Parallel()
	policy := DefaultPolicy()
	policy.Weights.ContractConfidence = 1
	_, err := Evaluate(validSurfaceResult(), policy)
	if !errors.Is(err, ErrInvalidPolicy) {
		t.Fatalf("expected invalid policy error, got %v", err)
	}
}

func TestResultCloneIsDeep(t *testing.T) {
	t.Parallel()
	result, err := Evaluate(validSurfaceResult(), DefaultPolicy())
	if err != nil {
		t.Fatalf("evaluate weather trust: %v", err)
	}
	cloned := result.Clone()
	cloned.Components[0].Score = 0
	cloned.AllowedScopes[0] = UsageScopeTrajectoryContext
	cloned.Limitations[0].Code = "changed"
	cloned.Explanations[0].Code = "changed"
	if result.Components[0].Score == 0 ||
		result.AllowedScopes[0] == UsageScopeTrajectoryContext ||
		result.Limitations[0].Code == "changed" ||
		result.Explanations[0].Code == "changed" {
		t.Fatal("weather trust clone shares backing arrays")
	}
}

func validSurfaceResult() weathercontract.Result {
	result := validFlightLevelResult()
	result.Status = weathercontract.ResultStatusLimited
	result.Confidence.Score = 0.55
	result.Confidence.Level = weathercontract.ConfidenceLevelMedium
	result.Samples[0].Position.AltitudeMeters = nil
	result.Samples[0].Position.VerticalReference = weathercontract.VerticalReferenceSurface
	result.Limitations = []weathercontract.Limitation{
		{
			Code:    "surface_weather_not_flight_level",
			Message: "Surface weather is not flight-level weather.",
			Scope:   "vertical_alignment",
		},
	}
	return result
}

func validFlightLevelResult() weathercontract.Result {
	asOfTime := time.Date(2026, time.July, 16, 12, 0, 0, 0, time.UTC)
	validAt := asOfTime.Add(-10 * time.Minute)
	availableAt := asOfTime.Add(-8 * time.Minute)
	retrievedAt := asOfTime.Add(-5 * time.Minute)

	altitude := 9000.0
	temperature := -42.0
	humidity := 45.0
	cloudCover := 30.0
	pressure := 300.0
	windSpeed := 35.0
	windDirection := 250.0
	windGust := 42.0
	conditionCode := 3

	return weathercontract.Result{
		SchemaVersion: weathercontract.SchemaVersionV1,
		Status:        weathercontract.ResultStatusComplete,
		TrajectoryID:  "trajectory-1",
		AsOfTime:      asOfTime,
		Samples: []weathercontract.Sample{
			{
				Sequence: 0,
				Position: weathercontract.Position{
					Latitude:          40.2,
					Longitude:         49.8,
					AltitudeMeters:    &altitude,
					VerticalReference: weathercontract.VerticalReferenceMeanSeaLevel,
				},
				Source: weathercontract.Source{
					Provider:     "test_weather",
					Dataset:      "flight_level_analysis",
					EvidenceKind: weathercontract.EvidenceKindAnalysis,
				},
				Features: weathercontract.FeatureVector{
					TemperatureCelsius:       &temperature,
					RelativeHumidityPercent:  &humidity,
					CloudCoverPercent:        &cloudCover,
					SurfacePressureHPA:       &pressure,
					WindSpeedMetersPerSecond: &windSpeed,
					WindDirectionDegrees:     &windDirection,
					WindGustsMetersPerSecond: &windGust,
					ConditionCode:            &conditionCode,
					ConditionCodeScheme:      "test_scheme",
				},
				ValidAt:     validAt,
				AvailableAt: availableAt,
				RetrievedAt: retrievedAt,
			},
		},
		Confidence: weathercontract.Confidence{
			Score: 0.90,
			Level: weathercontract.ConfidenceLevelHigh,
			Reasons: []weathercontract.ConfidenceReason{
				{
					Code:         "trusted_analysis",
					Message:      "Fresh flight-level weather analysis is available.",
					Contribution: 0.90,
				},
			},
		},
		Explanations: []weathercontract.Explanation{
			{
				Code:    "weather_context_only",
				Message: "Weather is contextual evidence only.",
			},
		},
		ScopeGuard: weathercontract.ScopeGuardContextOnly,
		Provenance: weathercontract.Provenance{
			InputFingerprint:  testFingerprint(),
			SourceNames:       []string{"test_weather"},
			LatestAvailableAt: availableAt,
		},
		GeneratedAt: asOfTime.Add(time.Minute),
	}
}

func validUnavailableResult() weathercontract.Result {
	asOfTime := time.Date(2026, time.July, 16, 12, 0, 0, 0, time.UTC)
	return weathercontract.Result{
		SchemaVersion: weathercontract.SchemaVersionV1,
		Status:        weathercontract.ResultStatusUnavailable,
		TrajectoryID:  "trajectory-1",
		AsOfTime:      asOfTime,
		Confidence: weathercontract.Confidence{
			Score: 0,
			Level: weathercontract.ConfidenceLevelNone,
		},
		Limitations: []weathercontract.Limitation{
			{
				Code:    "weather_unavailable",
				Message: "Weather is unavailable.",
				Scope:   "weather_context",
			},
		},
		Explanations: []weathercontract.Explanation{
			{
				Code:    "weather_withheld",
				Message: "Weather context is withheld.",
			},
		},
		ScopeGuard: weathercontract.ScopeGuardContextOnly,
		Provenance: weathercontract.Provenance{
			InputFingerprint: testFingerprint(),
		},
		GeneratedAt: asOfTime,
	}
}

func hasNotice(notices []Notice, code string) bool {
	for _, notice := range notices {
		if notice.Code == code {
			return true
		}
	}
	return false
}

func hasScope(scopes []UsageScope, target UsageScope) bool {
	for _, scope := range scopes {
		if scope == target {
			return true
		}
	}
	return false
}

func testFingerprint() string {
	return "sha256:" +
		"0123456789abcdef" +
		"0123456789abcdef" +
		"0123456789abcdef" +
		"0123456789abcdef"
}
