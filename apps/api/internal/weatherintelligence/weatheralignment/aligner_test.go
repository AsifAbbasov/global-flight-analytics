package weatheralignment

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
)

func TestDefaultPolicyIsValid(t *testing.T) {
	t.Parallel()
	if err := DefaultPolicy().Validate(); err != nil {
		t.Fatalf("default policy is invalid: %v", err)
	}
}

func TestAlignFreshFlightLevelWeather(t *testing.T) {
	t.Parallel()
	request := validFlightLevelRequest()
	result, err := Align(request)
	if err != nil {
		t.Fatalf("align weather: %v", err)
	}
	if result.Status != StatusComplete || result.AlignedCount != 2 || result.UnmatchedCount != 0 || result.CoverageRatio != 1 {
		t.Fatalf("unexpected complete result %#v", result)
	}
	for _, match := range result.Matches {
		if match.Status != MatchStatusAligned || match.WeatherSampleSequence == nil || match.Score < request.Policy.MinimumMatchScore {
			t.Fatalf("unexpected aligned match %#v", match)
		}
	}
	if result.Matches[0].AltitudeBasis != AltitudeBasisGeometric {
		t.Fatalf("expected geometric altitude preference, got %q", result.Matches[0].AltitudeBasis)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("complete result is invalid: %v", err)
	}
}

func TestAlignUsesBarometricAltitudeFallback(t *testing.T) {
	t.Parallel()
	request := validFlightLevelRequest()
	request.Trajectory.Points[0].GeometricAltitudeStatus = flightstate.AltitudeStatusUnavailable
	request.Trajectory.Points[0].GeometricAltitudeM = 0
	result, err := Align(request)
	if err != nil {
		t.Fatalf("align weather: %v", err)
	}
	if result.Matches[0].AltitudeBasis != AltitudeBasisBarometric {
		t.Fatalf("expected barometric fallback, got %q", result.Matches[0].AltitudeBasis)
	}
}

func TestAlignSurfaceWeatherToGroundOnly(t *testing.T) {
	t.Parallel()
	request := validSurfaceRequest()
	result, err := Align(request)
	if err != nil {
		t.Fatalf("align surface weather: %v", err)
	}
	if result.Status != StatusLimited || result.AlignedCount != 1 || result.UnmatchedCount != 1 {
		t.Fatalf("expected one ground match and one airborne rejection, got %#v", result)
	}
	if result.Matches[0].Status != MatchStatusAligned || result.Matches[0].AltitudeBasis != AltitudeBasisGround {
		t.Fatalf("ground point was not aligned: %#v", result.Matches[0])
	}
	if result.Matches[1].Status != MatchStatusUnmatched || !hasNotice(result.Matches[1].Limitations, "weather_usage_scope_not_allowed") {
		t.Fatalf("airborne point was not restricted: %#v", result.Matches[1])
	}
}

func TestAlignBlockedTrustProducesUnavailableResult(t *testing.T) {
	t.Parallel()
	request := validFlightLevelRequest()
	request.Trust.Decision = weathertrust.DecisionBlocked
	request.Trust.Usable = false
	request.Trust.AllowedScopes = nil
	request.Trust.Limitations = []weathertrust.Notice{{Code: "blocked", Message: "Weather is blocked."}}
	result, err := Align(request)
	if err != nil {
		t.Fatalf("align blocked trust: %v", err)
	}
	if result.Status != StatusUnavailable || result.AlignedCount != 0 || result.UnmatchedCount != 2 {
		t.Fatalf("unexpected blocked result %#v", result)
	}
}

func TestAlignRejectsFutureTrajectoryPoint(t *testing.T) {
	t.Parallel()
	request := validFlightLevelRequest()
	request.Trajectory.Points[0].ObservedAt = request.Weather.AsOfTime.Add(time.Second)
	_, err := Align(request)
	if !errors.Is(err, ErrTrajectoryInvalid) {
		t.Fatalf("expected future trajectory rejection, got %v", err)
	}
}

func TestAlignRejectsInvalidWeatherContract(t *testing.T) {
	t.Parallel()
	request := validFlightLevelRequest()
	request.Weather.ScopeGuard = "operational"
	_, err := Align(request)
	if !errors.Is(err, ErrWeatherContractInvalid) {
		t.Fatalf("expected weather contract error, got %v", err)
	}
}

func TestAlignRejectsInvalidTrustResult(t *testing.T) {
	t.Parallel()
	request := validFlightLevelRequest()
	request.Trust.InputFingerprint = "invalid"
	_, err := Align(request)
	if !errors.Is(err, ErrTrustResultInvalid) {
		t.Fatalf("expected trust result error, got %v", err)
	}
}

func TestAlignHorizontalBoundary(t *testing.T) {
	t.Parallel()
	request := validFlightLevelRequest()
	request.Weather.Samples[0].Position.Latitude = 50
	result, err := Align(request)
	if err != nil {
		t.Fatalf("align distant weather: %v", err)
	}
	if result.Status != StatusUnavailable || !hasNotice(result.Matches[0].Limitations, "weather_horizontal_boundary_exceeded") {
		t.Fatalf("horizontal boundary was not enforced: %#v", result)
	}
}

func TestAlignTemporalBoundary(t *testing.T) {
	t.Parallel()
	request := validFlightLevelRequest()
	request.Weather.Samples[0].ValidAt = request.Trajectory.Points[0].ObservedAt.Add(-3 * time.Hour)
	result, err := Align(request)
	if err != nil {
		t.Fatalf("align old weather: %v", err)
	}
	if result.Status != StatusUnavailable || !hasNotice(result.Matches[0].Limitations, "weather_temporal_boundary_exceeded") {
		t.Fatalf("temporal boundary was not enforced: %#v", result)
	}
}

func TestAlignVerticalBoundary(t *testing.T) {
	t.Parallel()
	request := validFlightLevelRequest()
	altitude := 2000.0
	request.Weather.Samples[0].Position.AltitudeMeters = &altitude
	result, err := Align(request)
	if err != nil {
		t.Fatalf("align vertical mismatch: %v", err)
	}
	if result.Status != StatusUnavailable || !hasNotice(result.Matches[0].Limitations, "weather_vertical_boundary_exceeded") {
		t.Fatalf("vertical boundary was not enforced: %#v", result)
	}
}

func TestAlignChoosesHighestScoringSample(t *testing.T) {
	t.Parallel()
	request := validFlightLevelRequest()
	second := cloneWeatherSample(request.Weather.Samples[0])
	second.Sequence = 1
	second.Position.Latitude += 0.50
	request.Weather.Samples = append(request.Weather.Samples, second)
	result, err := Align(request)
	if err != nil {
		t.Fatalf("align multiple samples: %v", err)
	}
	for _, match := range result.Matches {
		if match.WeatherSampleSequence == nil || *match.WeatherSampleSequence != 0 {
			t.Fatalf("highest-scoring sample was not selected: %#v", match)
		}
	}
}

func TestAlignFingerprintIsDeterministic(t *testing.T) {
	t.Parallel()
	request := validFlightLevelRequest()
	first, err := Align(request)
	if err != nil {
		t.Fatalf("align first result: %v", err)
	}
	second, err := Align(request)
	if err != nil {
		t.Fatalf("align second result: %v", err)
	}
	if first.InputFingerprint != second.InputFingerprint {
		t.Fatal("same alignment input produced different fingerprints")
	}
	changed := request
	changed.Policy.MaximumTemporalDistance = 2 * time.Hour
	changedResult, err := Align(changed)
	if err != nil {
		t.Fatalf("align changed policy: %v", err)
	}
	if changedResult.InputFingerprint == first.InputFingerprint {
		t.Fatal("changed alignment policy did not change fingerprint")
	}
}

func TestResultCloneIsDeep(t *testing.T) {
	t.Parallel()
	result, err := Align(validFlightLevelRequest())
	if err != nil {
		t.Fatalf("align weather: %v", err)
	}
	cloned := result.Clone()
	*cloned.Matches[0].WeatherSampleSequence = 99
	*cloned.Matches[0].HorizontalDistanceKilometers = 99
	cloned.Matches[0].Components[0].Score = 0
	cloned.Explanations[0].Code = "changed"
	if *result.Matches[0].WeatherSampleSequence == 99 ||
		*result.Matches[0].HorizontalDistanceKilometers == 99 ||
		result.Matches[0].Components[0].Score == 0 ||
		result.Explanations[0].Code == "changed" {
		t.Fatal("weather alignment clone shares backing data")
	}
}

func validFlightLevelRequest() Request {
	asOfTime := time.Date(2026, time.July, 16, 14, 0, 0, 0, time.UTC)
	weatherAltitude := 9100.0
	temperature := -43.0
	humidity := 45.0
	cloudCover := 20.0
	pressure := 300.0
	windSpeed := 35.0
	windDirection := 250.0
	windGust := 42.0
	conditionCode := 3

	weather := weathercontract.Result{
		SchemaVersion: weathercontract.SchemaVersionV1,
		Status:        weathercontract.ResultStatusComplete,
		TrajectoryID:  "trajectory-1",
		AsOfTime:      asOfTime,
		Samples: []weathercontract.Sample{
			{
				Sequence: 0,
				Position: weathercontract.Position{
					Latitude:          40.20,
					Longitude:         49.80,
					AltitudeMeters:    &weatherAltitude,
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
				ValidAt:     asOfTime.Add(-10 * time.Minute),
				AvailableAt: asOfTime.Add(-8 * time.Minute),
				RetrievedAt: asOfTime.Add(-5 * time.Minute),
			},
		},
		Confidence: weathercontract.Confidence{
			Score: 0.90,
			Level: weathercontract.ConfidenceLevelHigh,
			Reasons: []weathercontract.ConfidenceReason{
				{Code: "trusted_analysis", Message: "Fresh flight-level weather is available.", Contribution: 0.90},
			},
		},
		Explanations: []weathercontract.Explanation{
			{Code: "weather_context_only", Message: "Weather is contextual evidence."},
		},
		ScopeGuard: weathercontract.ScopeGuardContextOnly,
		Provenance: weathercontract.Provenance{
			InputFingerprint:  testFingerprint("weather"),
			SourceNames:       []string{"test_weather"},
			LatestAvailableAt: asOfTime.Add(-8 * time.Minute),
		},
		GeneratedAt: asOfTime.Add(time.Minute),
	}

	trust := validTrust(
		asOfTime,
		weathertrust.DecisionAllowed,
		[]weathertrust.UsageScope{
			weathertrust.UsageScopeProjectionUncertainty,
			weathertrust.UsageScopeTrajectoryContext,
		},
	)

	return Request{
		Trajectory: trajectory.FlightTrajectory{
			ID: "trajectory-1",
			Points: []trajectory.TrackPoint4D{
				{
					ID:                       "point-1",
					Latitude:                 40.21,
					Longitude:                49.81,
					GeometricAltitudeM:       9000,
					GeometricAltitudeStatus:  flightstate.AltitudeStatusObserved,
					BarometricAltitudeM:      8900,
					BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
					ObservedAt:               asOfTime.Add(-9 * time.Minute),
				},
				{
					ID:                       "point-2",
					Latitude:                 40.25,
					Longitude:                49.85,
					GeometricAltitudeM:       9200,
					GeometricAltitudeStatus:  flightstate.AltitudeStatusObserved,
					BarometricAltitudeM:      9100,
					BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
					ObservedAt:               asOfTime.Add(-7 * time.Minute),
				},
			},
		},
		Weather:     weather,
		Trust:       trust,
		Policy:      DefaultPolicy(),
		GeneratedAt: asOfTime.Add(2 * time.Minute),
	}
}

func validSurfaceRequest() Request {
	request := validFlightLevelRequest()
	asOfTime := request.Weather.AsOfTime
	request.Trajectory.Points[0].OnGround = true
	request.Trajectory.Points[0].GeometricAltitudeStatus = flightstate.AltitudeStatusGround
	request.Trajectory.Points[0].BarometricAltitudeStatus = flightstate.AltitudeStatusGround
	request.Trajectory.Points[0].GeometricAltitudeM = 0
	request.Trajectory.Points[0].BarometricAltitudeM = 0

	request.Weather.Status = weathercontract.ResultStatusLimited
	request.Weather.Confidence.Score = 0.55
	request.Weather.Confidence.Level = weathercontract.ConfidenceLevelMedium
	request.Weather.Samples[0].Position.AltitudeMeters = nil
	request.Weather.Samples[0].Position.VerticalReference = weathercontract.VerticalReferenceSurface
	request.Weather.Limitations = []weathercontract.Limitation{
		{Code: "surface_weather_not_flight_level", Message: "Surface weather is not flight-level weather.", Scope: "vertical_alignment"},
	}
	request.Trust = validTrust(
		asOfTime,
		weathertrust.DecisionLimited,
		[]weathertrust.UsageScope{weathertrust.UsageScopeSurfaceContext},
	)
	request.Trust.Limitations = []weathertrust.Notice{
		{Code: "surface_weather_not_flight_level", Message: "Surface weather is limited to surface context."},
	}
	return request
}

func validTrust(asOfTime time.Time, decision weathertrust.Decision, scopes []weathertrust.UsageScope) weathertrust.Result {
	limitations := []weathertrust.Notice(nil)
	if decision == weathertrust.DecisionLimited {
		limitations = []weathertrust.Notice{{Code: "limited", Message: "Weather trust is limited."}}
	}
	return weathertrust.Result{
		Version:  weathertrust.Version,
		Decision: decision,
		Usable:   decision != weathertrust.DecisionBlocked,
		AsOfTime: asOfTime,
		Score:    0.90,
		Components: []weathertrust.Component{
			{Name: weathertrust.ComponentContractConfidence, Score: 0.90, Weight: 0.35},
			{Name: weathertrust.ComponentTemporalFreshness, Score: 0.90, Weight: 0.30},
			{Name: weathertrust.ComponentFeatureCompleteness, Score: 1, Weight: 0.20},
			{Name: weathertrust.ComponentVerticalApplicability, Score: 1, Weight: 0.15},
		},
		AllowedScopes: scopes,
		Limitations:   limitations,
		Explanations: []weathertrust.Notice{
			{Code: "weather_context_only", Message: "Weather is contextual evidence."},
		},
		InputFingerprint: testFingerprint("trust"),
	}
}

func cloneWeatherSample(sample weathercontract.Sample) weathercontract.Sample {
	cloned := sample
	if sample.Position.AltitudeMeters != nil {
		altitude := *sample.Position.AltitudeMeters
		cloned.Position.AltitudeMeters = &altitude
	}
	return cloned
}

func hasNotice(notices []Notice, code string) bool {
	for _, notice := range notices {
		if notice.Code == code {
			return true
		}
	}
	return false
}

func testFingerprint(seed string) string {
	padding := "0123456789abcdef" + "0123456789abcdef" + "0123456789abcdef" + "0123456789abcdef"
	if seed == "trust" {
		padding = "abcdef0123456789" + "abcdef0123456789" + "abcdef0123456789" + "abcdef0123456789"
	}
	return "sha256:" + padding
}
