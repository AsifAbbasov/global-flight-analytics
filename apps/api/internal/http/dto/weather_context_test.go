package dto

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheruncertainty"
)

func TestToWeatherContextResponseMapsAllSectionsAndClonesPointers(t *testing.T) {
	t.Parallel()

	asOfTime := time.Date(2026, time.July, 16, 14, 0, 0, 0, time.UTC)
	temperature := -41.5
	windSpeed := 28.0
	conditionCode := 61
	altitude := 9000.0
	horizontalResolution := 9.0
	verticalRadius := 600.0
	encounterMean := -41.5
	direction := 250.0
	concentration := 0.9
	weatherSequence := 0
	weatherValidAt := asOfTime.Add(-10 * time.Minute)
	temporalDistance := 2 * time.Minute
	horizontalDistance := 3.5
	verticalDistance := 100.0

	weather := weathercontract.Result{
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
					Provider:                       "test-weather",
					Dataset:                        "flight-level",
					EvidenceKind:                   weathercontract.EvidenceKindAnalysis,
					HorizontalResolutionKilometers: &horizontalResolution,
					TemporalResolution:             time.Hour,
				},
				Features: weathercontract.FeatureVector{
					TemperatureCelsius:       &temperature,
					WindSpeedMetersPerSecond: &windSpeed,
					ConditionCode:            &conditionCode,
					ConditionCodeScheme:      "wmo",
				},
				ValidAt:     weatherValidAt,
				AvailableAt: weatherValidAt.Add(time.Minute),
				RetrievedAt: weatherValidAt.Add(2 * time.Minute),
			},
		},
		Confidence: weathercontract.Confidence{
			Score: 0.9,
			Level: weathercontract.ConfidenceLevelHigh,
			Reasons: []weathercontract.ConfidenceReason{
				{
					Code:         "trusted",
					Message:      "Trusted weather.",
					Contribution: 0.9,
				},
			},
		},
		Explanations: []weathercontract.Explanation{
			{Code: "context", Message: "Context only."},
		},
		ScopeGuard: weathercontract.ScopeGuardContextOnly,
		Provenance: weathercontract.Provenance{
			InputFingerprint:  testWeatherContextFingerprint("weather"),
			SourceNames:       []string{"test-weather"},
			LatestAvailableAt: weatherValidAt.Add(time.Minute),
		},
		GeneratedAt: asOfTime.Add(time.Minute),
	}

	trust := weathertrust.Result{
		Version:  weathertrust.Version,
		Decision: weathertrust.DecisionAllowed,
		Usable:   true,
		AsOfTime: asOfTime,
		Score:    0.9,
		Components: []weathertrust.Component{
			{Name: weathertrust.ComponentContractConfidence, Score: 0.9, Weight: 0.35},
			{Name: weathertrust.ComponentTemporalFreshness, Score: 0.9, Weight: 0.30},
			{Name: weathertrust.ComponentFeatureCompleteness, Score: 0.8, Weight: 0.20},
			{Name: weathertrust.ComponentVerticalApplicability, Score: 1, Weight: 0.15},
		},
		AllowedScopes: []weathertrust.UsageScope{
			weathertrust.UsageScopeProjectionUncertainty,
			weathertrust.UsageScopeTrajectoryContext,
		},
		Explanations: []weathertrust.Notice{
			{Code: "allowed", Message: "Weather is allowed."},
		},
		InputFingerprint: testWeatherContextFingerprint("trust"),
	}

	alignment := weatheralignment.Result{
		Version:       weatheralignment.Version,
		Status:        weatheralignment.StatusComplete,
		TrajectoryID:  "trajectory-1",
		AsOfTime:      asOfTime,
		TrustDecision: weathertrust.DecisionAllowed,
		TrustScore:    0.9,
		PointCount:    1,
		AlignedCount:  1,
		CoverageRatio: 1,
		Matches: []weatheralignment.Match{
			{
				TrajectoryPointSequence:      0,
				TrajectoryPointID:            "point-1",
				TrajectoryObservedAt:         asOfTime.Add(-9 * time.Minute),
				WeatherSampleSequence:        &weatherSequence,
				WeatherValidAt:               &weatherValidAt,
				Status:                       weatheralignment.MatchStatusAligned,
				AltitudeBasis:                weatheralignment.AltitudeBasisGeometric,
				AltitudeMeters:               &altitude,
				HorizontalDistanceKilometers: &horizontalDistance,
				TemporalDistance:             &temporalDistance,
				VerticalDistanceMeters:       &verticalDistance,
				Score:                        0.9,
				Components: []weatheralignment.Component{
					{Name: weatheralignment.ComponentHorizontal, Score: 0.9, Weight: 0.45},
					{Name: weatheralignment.ComponentTemporal, Score: 0.9, Weight: 0.35},
					{Name: weatheralignment.ComponentVertical, Score: 0.9, Weight: 0.20},
				},
			},
		},
		Explanations: []weatheralignment.Notice{
			{Code: "aligned", Message: "Weather aligned."},
		},
		InputFingerprint: testWeatherContextFingerprint("alignment"),
		GeneratedAt:      asOfTime.Add(2 * time.Minute),
	}

	startedAt := asOfTime.Add(-9 * time.Minute)
	endedAt := startedAt
	encounter := weatherencounter.Result{
		Version:                weatherencounter.Version,
		Status:                 weatherencounter.StatusComplete,
		TrajectoryID:           "trajectory-1",
		AsOfTime:               asOfTime,
		AlignmentStatus:        weatheralignment.StatusComplete,
		AlignmentCoverageRatio: 1,
		PointCount:             1,
		EncounterPointCount:    1,
		ProfileCoverageRatio:   1,
		EncounterStartedAt:     &startedAt,
		EncounterEndedAt:       &endedAt,
		TemperatureCelsius: weatherencounter.MetricSummary{
			PresentCount: 1, CoverageRatio: 1,
			Minimum: &encounterMean, Maximum: &encounterMean, Mean: &encounterMean,
		},
		WindDirectionDegrees: weatherencounter.CircularDirectionSummary{
			PresentCount: 1, CoverageRatio: 1,
			MeanDirectionDegrees: &direction, Concentration: &concentration,
		},
		Points: []weatherencounter.EncounterPoint{
			{
				TrajectoryPointSequence: 0,
				TrajectoryPointID:       "point-1",
				TrajectoryObservedAt:    startedAt,
				WeatherSampleSequence:   0,
				WeatherValidAt:          weatherValidAt,
				AlignmentScore:          0.9,
				FeatureCount:            3,
			},
		},
		Explanations: []weatherencounter.Notice{
			{Code: "profile", Message: "Profile available."},
		},
		InputFingerprint: testWeatherContextFingerprint("encounter"),
		GeneratedAt:      asOfTime.Add(3 * time.Minute),
	}

	projection := projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        projectioncontract.ResultStatusComplete,
		TrajectoryID:  "trajectory-1",
		Method: projectioncontract.Method{
			Name: "test", Version: "v1",
			DecisionClass: projectioncontract.DecisionClassProjectDerived,
		},
		Horizon: projectioncontract.Horizon{
			AsOfTime: asOfTime,
			EndTime:  asOfTime.Add(10 * time.Minute),
			Step:     10 * time.Minute,
		},
		Points: []projectioncontract.ProjectionPoint{
			{
				Sequence:     0,
				ForecastTime: asOfTime.Add(10 * time.Minute),
				Position: projectioncontract.Position{
					Latitude: 40.3, Longitude: 49.9, AltitudeM: &altitude,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 1500,
					VerticalRadiusM:   &verticalRadius,
				},
				Confidence: projectioncontract.Confidence{
					Score: 0.7, Level: projectioncontract.ConfidenceLevelMedium,
				},
			},
		},
		Confidence: projectioncontract.Confidence{
			Score: 0.7, Level: projectioncontract.ConfidenceLevelMedium,
		},
		Explanations: []projectioncontract.Explanation{
			{Code: "projection", Message: "Projection."},
		},
		ScopeGuard: projectioncontract.ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: testWeatherContextFingerprint("projection"),
		},
		GeneratedAt: asOfTime.Add(4 * time.Minute),
	}

	uncertainty := weatheruncertainty.Result{
		Version:           weatheruncertainty.Version,
		Status:            weatheruncertainty.StatusApplied,
		TrajectoryID:      "trajectory-1",
		AsOfTime:          asOfTime,
		SeverityScore:     0.5,
		WeatherMultiplier: 1.75,
		Components: []weatheruncertainty.Component{
			{Name: weatheruncertainty.ComponentWindSpeed, Score: 0.5, Weight: 0.30},
			{Name: weatheruncertainty.ComponentWindGust, Score: 0.5, Weight: 0.20},
			{Name: weatheruncertainty.ComponentPrecipitation, Score: 0.5, Weight: 0.15},
			{Name: weatheruncertainty.ComponentCloudCover, Score: 0.5, Weight: 0.10},
			{Name: weatheruncertainty.ComponentEvidenceQuality, Score: 0.5, Weight: 0.25},
		},
		PointAdjustments: []weatheruncertainty.PointAdjustment{
			{
				Sequence: 0, ForecastTime: asOfTime.Add(10 * time.Minute),
				HorizonProgress: 1, Multiplier: 1.75,
				OriginalHorizontalRadiusM: 1000,
				AdjustedHorizontalRadiusM: 1500,
				OriginalVerticalRadiusM:   &verticalRadius,
				AdjustedVerticalRadiusM:   &verticalRadius,
				OriginalConfidenceScore:   0.8,
				AdjustedConfidenceScore:   0.7,
			},
		},
		AdjustedProjection: projection,
		Explanations: []weatheruncertainty.Notice{
			{Code: "adjusted", Message: "Uncertainty adjusted."},
		},
		InputFingerprint: testWeatherContextFingerprint("uncertainty"),
		GeneratedAt:      asOfTime.Add(4 * time.Minute),
	}

	response := ToWeatherContextResponse(
		weather,
		trust,
		alignment,
		encounter,
		uncertainty,
		testWeatherContextFingerprint("aggregate"),
		asOfTime.Add(5*time.Minute),
	)

	if response.Version != WeatherContextResponseVersion ||
		response.Weather.Samples[0].Features.PresentCount != 3 ||
		response.Trust.Decision != "allowed" ||
		response.Alignment.Matches[0].TemporalDistanceSeconds == nil ||
		*response.Alignment.Matches[0].TemporalDistanceSeconds != 120 ||
		response.Encounter.WindDirectionDegrees.MeanDirectionDegrees == nil ||
		response.Uncertainty.WeatherMultiplier != 1.75 ||
		response.Uncertainty.AdjustedProjection.Points[0].Uncertainty.HorizontalRadiusM != 1500 {
		t.Fatalf("unexpected mapped Weather Context response: %#v", response)
	}

	*response.Weather.Samples[0].Features.TemperatureCelsius = 99
	*response.Alignment.Matches[0].AltitudeMeters = 99
	*response.Encounter.TemperatureCelsius.Mean = 99
	*response.Uncertainty.PointAdjustments[0].AdjustedVerticalRadiusM = 99

	if temperature == 99 ||
		altitude == 99 ||
		encounterMean == 99 ||
		verticalRadius == 99 {
		t.Fatal("Weather Context DTO shares pointer values with domain results")
	}
}

func testWeatherContextFingerprint(seed string) string {
	values := map[string]string{
		"weather":     "0123456789abcdef",
		"trust":       "abcdef0123456789",
		"alignment":   "0011223344556677",
		"encounter":   "7766554433221100",
		"projection":  "1234567890abcdef",
		"uncertainty": "fedcba0987654321",
		"aggregate":   "a1b2c3d4e5f60718",
	}
	value := values[seed]
	return "sha256:" + value + value + value + value
}
