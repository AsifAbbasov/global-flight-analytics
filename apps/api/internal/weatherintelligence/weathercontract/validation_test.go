package weathercontract

import (
	"math"
	"testing"
	"time"
)

func TestValidateCompleteWeatherFeatureResult(
	t *testing.T,
) {
	t.Parallel()

	result := validCompleteResult()
	report := Validate(result)

	if report.Status != ValidationStatusValid {
		t.Fatalf(
			"expected valid report, got %#v",
			report.Issues,
		)
	}
	if len(report.Issues) != 0 {
		t.Fatalf(
			"expected no validation issues, got %#v",
			report.Issues,
		)
	}
}

func TestValidateForecastAfterAsOfWithoutFutureLeak(
	t *testing.T,
) {
	t.Parallel()

	result := validCompleteResult()
	result.Samples[0].Source.EvidenceKind =
		EvidenceKindForecast
	result.Samples[0].ValidAt =
		result.AsOfTime.Add(30 * time.Minute)
	result.Samples[0].AvailableAt =
		result.AsOfTime.Add(-10 * time.Minute)
	result.Samples[0].RetrievedAt =
		result.AsOfTime.Add(-5 * time.Minute)
	result.Provenance.LatestAvailableAt =
		result.Samples[0].AvailableAt

	report := Validate(result)
	if report.Status != ValidationStatusValid {
		t.Fatalf(
			"forecast available before as-of should be valid, got %#v",
			report.Issues,
		)
	}
}

func TestValidateRejectsFutureEvidenceAvailability(
	t *testing.T,
) {
	t.Parallel()

	result := validCompleteResult()
	result.Samples[0].AvailableAt =
		result.AsOfTime.Add(time.Minute)
	result.Samples[0].RetrievedAt =
		result.AsOfTime.Add(2 * time.Minute)
	result.Provenance.LatestAvailableAt =
		result.Samples[0].AvailableAt

	report := Validate(result)
	if !report.HasCode(
		IssueFutureEvidenceAvailability,
	) {
		t.Fatalf(
			"expected future evidence availability issue, got %#v",
			report.Issues,
		)
	}
}

func TestValidateRejectsFutureObservation(
	t *testing.T,
) {
	t.Parallel()

	result := validCompleteResult()
	result.Samples[0].Source.EvidenceKind =
		EvidenceKindObservation
	result.Samples[0].ValidAt =
		result.AsOfTime.Add(time.Minute)

	report := Validate(result)
	if !report.HasCode(
		IssueFutureNonForecastEvidence,
	) {
		t.Fatalf(
			"expected future non-forecast issue, got %#v",
			report.Issues,
		)
	}
}

func TestValidateUnavailableWeatherFeatureResult(
	t *testing.T,
) {
	t.Parallel()

	asOfTime := time.Date(
		2026,
		time.July,
		16,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	result := Result{
		SchemaVersion: SchemaVersionV1,
		Status:        ResultStatusUnavailable,
		TrajectoryID:  "trajectory-1",
		AsOfTime:      asOfTime,
		Confidence: Confidence{
			Score: 0,
			Level: ConfidenceLevelNone,
		},
		Limitations: []Limitation{
			{
				Code:    "weather_source_unavailable",
				Message: "No weather source was available.",
				Scope:   "weather_context",
			},
		},
		Explanations: []Explanation{
			{
				Code:    "weather_context_withheld",
				Message: "Weather context was withheld.",
			},
		},
		ScopeGuard: ScopeGuardContextOnly,
		Provenance: Provenance{
			InputFingerprint: testFingerprint(),
		},
		GeneratedAt: asOfTime,
	}

	report := Validate(result)
	if report.Status != ValidationStatusValid {
		t.Fatalf(
			"expected valid unavailable result, got %#v",
			report.Issues,
		)
	}
}

func TestValidateLimitedResultRequiresLimitation(
	t *testing.T,
) {
	t.Parallel()

	result := validCompleteResult()
	result.Status = ResultStatusLimited
	result.Limitations = nil

	report := Validate(result)
	if !report.HasCode(
		IssueLimitedContractInvalid,
	) {
		t.Fatalf(
			"expected limited contract issue, got %#v",
			report.Issues,
		)
	}
}

func TestValidateRejectsInvalidIdentityAndScope(
	t *testing.T,
) {
	t.Parallel()

	result := validCompleteResult()
	result.SchemaVersion = "wrong"
	result.Status = "wrong"
	result.TrajectoryID = ""
	result.AsOfTime = time.Time{}
	result.GeneratedAt = time.Time{}
	result.ScopeGuard = "operational"

	report := Validate(result)
	for _, code := range []string{
		IssueSchemaVersionInvalid,
		IssueStatusInvalid,
		IssueTrajectoryIDRequired,
		IssueAsOfTimeRequired,
		IssueGeneratedAtInvalid,
		IssueScopeGuardInvalid,
	} {
		if !report.HasCode(code) {
			t.Fatalf(
				"expected issue %q, got %#v",
				code,
				report.Issues,
			)
		}
	}
}

func TestValidateRejectsInvalidSampleStructure(
	t *testing.T,
) {
	t.Parallel()

	result := validCompleteResult()
	result.Samples = append(
		result.Samples,
		cloneSample(result.Samples[0]),
	)
	result.Samples[0].Sequence = 2
	result.Samples[0].Position.Latitude = 91
	result.Samples[0].Position.VerticalReference =
		"unsupported"
	result.Samples[0].Source.Provider = ""
	result.Samples[0].Source.Dataset = ""
	result.Samples[0].Source.EvidenceKind =
		"unsupported"
	negativeResolution := -1.0
	result.Samples[0].Source.
		HorizontalResolutionKilometers =
		&negativeResolution
	result.Samples[0].Source.TemporalResolution =
		-time.Minute
	result.Samples[0].Features =
		FeatureVector{}
	result.Samples[1].Sequence = 1
	result.Samples[1].ValidAt =
		result.Samples[0].
			ValidAt.Add(-time.Minute)
	result.Provenance.LatestAvailableAt =
		result.Samples[1].AvailableAt

	report := Validate(result)
	for _, code := range []string{
		IssueSampleSequenceInvalid,
		IssueSamplePositionInvalid,
		IssueVerticalReferenceInvalid,
		IssueSourceInvalid,
		IssueEvidenceKindInvalid,
		IssueSourceResolutionInvalid,
		IssueFeatureVectorEmpty,
		IssueSampleOrderInvalid,
	} {
		if !report.HasCode(code) {
			t.Fatalf(
				"expected issue %q, got %#v",
				code,
				report.Issues,
			)
		}
	}
}

func TestValidateRejectsInvalidFeatureValues(
	t *testing.T,
) {
	t.Parallel()

	result := validCompleteResult()
	temperature := math.Inf(1)
	humidity := 101.0
	precipitation := -1.0
	cloudCover := 101.0
	pressure := 99.0
	windSpeed := -1.0
	windDirection := 360.0
	windGust := 251.0

	result.Samples[0].Features = FeatureVector{
		TemperatureCelsius:       &temperature,
		RelativeHumidityPercent:  &humidity,
		PrecipitationMillimeters: &precipitation,
		CloudCoverPercent:        &cloudCover,
		SurfacePressureHPA:       &pressure,
		WindSpeedMetersPerSecond: &windSpeed,
		WindDirectionDegrees:     &windDirection,
		WindGustsMetersPerSecond: &windGust,
	}

	report := Validate(result)
	if !report.HasCode(
		IssueFeatureValueInvalid,
	) {
		t.Fatalf(
			"expected invalid feature issue, got %#v",
			report.Issues,
		)
	}
}

func TestValidateRejectsInvalidConfidence(
	t *testing.T,
) {
	t.Parallel()

	result := validCompleteResult()
	result.Confidence.Score = 0.90
	result.Confidence.Level =
		ConfidenceLevelLow
	result.Confidence.Reasons[0].
		Contribution = 2

	report := Validate(result)
	if !report.HasCode(
		IssueConfidenceInvalid,
	) {
		t.Fatalf(
			"expected confidence issue, got %#v",
			report.Issues,
		)
	}
	if !report.HasCode(
		IssueConfidenceReasonInvalid,
	) {
		t.Fatalf(
			"expected confidence reason issue, got %#v",
			report.Issues,
		)
	}
}

func TestValidateRejectsProvenanceMismatch(
	t *testing.T,
) {
	t.Parallel()

	result := validCompleteResult()
	result.Provenance.InputFingerprint = "invalid"
	result.Provenance.SourceNames =
		[]string{"other_provider"}
	result.Provenance.LatestAvailableAt =
		result.AsOfTime.Add(-2 * time.Hour)

	report := Validate(result)
	for _, code := range []string{
		IssueFingerprintInvalid,
		IssueProvenanceSourceMismatch,
		IssueLatestAvailableAtMismatch,
	} {
		if !report.HasCode(code) {
			t.Fatalf(
				"expected issue %q, got %#v",
				code,
				report.Issues,
			)
		}
	}
}

func TestResultCloneIsDeep(
	t *testing.T,
) {
	t.Parallel()

	result := validCompleteResult()
	cloned := result.Clone()

	*cloned.Samples[0].
		Position.AltitudeMeters = 1
	*cloned.Samples[0].
		Features.TemperatureCelsius = 99
	cloned.Confidence.Reasons[0].Code =
		"changed"
	cloned.Provenance.SourceNames[0] =
		"changed"

	if *result.Samples[0].
		Position.AltitudeMeters == 1 {
		t.Fatal(
			"position altitude pointer was not cloned",
		)
	}
	if *result.Samples[0].
		Features.TemperatureCelsius == 99 {
		t.Fatal(
			"feature pointer was not cloned",
		)
	}
	if result.Confidence.Reasons[0].Code ==
		"changed" {
		t.Fatal(
			"confidence reasons were not cloned",
		)
	}
	if result.Provenance.SourceNames[0] ==
		"changed" {
		t.Fatal(
			"provenance sources were not cloned",
		)
	}
}

func TestPresentCount(
	t *testing.T,
) {
	t.Parallel()

	temperature := 10.0
	windSpeed := 20.0

	features := FeatureVector{
		TemperatureCelsius:       &temperature,
		WindSpeedMetersPerSecond: &windSpeed,
	}
	if got := features.PresentCount(); got != 2 {
		t.Fatalf(
			"expected two present features, got %d",
			got,
		)
	}
}

func validCompleteResult() Result {
	asOfTime := time.Date(
		2026,
		time.July,
		16,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	validAt := asOfTime.Add(-5 * time.Minute)
	availableAt := asOfTime.Add(-4 * time.Minute)
	retrievedAt := asOfTime.Add(-3 * time.Minute)

	altitude := 120.0
	resolution := 11.0
	temperature := 23.5
	humidity := 54.0
	precipitation := 0.0
	rain := 0.0
	cloudCover := 18.0
	pressure := 1008.2
	windSpeed := 7.2
	windDirection := 245.0
	windGust := 10.1

	return Result{
		SchemaVersion: SchemaVersionV1,
		Status:        ResultStatusComplete,
		TrajectoryID:  "trajectory-1",
		AsOfTime:      asOfTime,
		Samples: []Sample{
			{
				Sequence: 0,
				Position: Position{
					Latitude:          40.4675,
					Longitude:         50.0467,
					AltitudeMeters:    &altitude,
					VerticalReference: VerticalReferenceMeanSeaLevel,
				},
				Source: Source{
					Provider:                       "open_meteo",
					Dataset:                        "current_weather",
					EvidenceKind:                   EvidenceKindAnalysis,
					HorizontalResolutionKilometers: &resolution,
					TemporalResolution:             time.Hour,
				},
				Features: FeatureVector{
					TemperatureCelsius:       &temperature,
					RelativeHumidityPercent:  &humidity,
					PrecipitationMillimeters: &precipitation,
					RainMillimeters:          &rain,
					CloudCoverPercent:        &cloudCover,
					SurfacePressureHPA:       &pressure,
					WindSpeedMetersPerSecond: &windSpeed,
					WindDirectionDegrees:     &windDirection,
					WindGustsMetersPerSecond: &windGust,
				},
				ValidAt:     validAt,
				AvailableAt: availableAt,
				RetrievedAt: retrievedAt,
			},
		},
		Confidence: Confidence{
			Score: 0.82,
			Level: ConfidenceLevelHigh,
			Reasons: []ConfidenceReason{
				{
					Code:         "provider_sample_available",
					Message:      "A provider weather sample is available.",
					Contribution: 0.82,
				},
			},
		},
		Explanations: []Explanation{
			{
				Code:    "weather_context_only",
				Message: "Weather is context and not proof of maneuver cause.",
			},
		},
		ScopeGuard: ScopeGuardContextOnly,
		Provenance: Provenance{
			InputFingerprint:  testFingerprint(),
			SourceNames:       []string{"open_meteo"},
			LatestAvailableAt: availableAt,
		},
		GeneratedAt: asOfTime.Add(time.Minute),
	}
}

func testFingerprint() string {
	return "sha256:" +
		"0123456789abcdef" +
		"0123456789abcdef" +
		"0123456789abcdef" +
		"0123456789abcdef"
}
