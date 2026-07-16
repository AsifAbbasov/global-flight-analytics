package weatherencounter

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
)

func TestDefaultPolicyIsValid(t *testing.T) {
	t.Parallel()

	if err := DefaultPolicy().Validate(); err != nil {
		t.Fatalf("default policy is invalid: %v", err)
	}
}

func TestBuildCompleteWeatherEncounterProfile(t *testing.T) {
	t.Parallel()

	result, err := Build(validRequest())
	if err != nil {
		t.Fatalf("build weather encounter: %v", err)
	}

	if result.Status != StatusComplete ||
		result.EncounterPointCount != 2 ||
		result.UnprofiledPointCount != 0 ||
		result.ProfileCoverageRatio != 1 {
		t.Fatalf("unexpected complete profile %#v", result)
	}
	if result.TemperatureCelsius.Mean == nil ||
		*result.TemperatureCelsius.Mean != -41 {
		t.Fatalf(
			"unexpected temperature summary %#v",
			result.TemperatureCelsius,
		)
	}
	if result.WindSpeedMetersPerSecond.Mean == nil ||
		*result.WindSpeedMetersPerSecond.Mean != 35 {
		t.Fatalf(
			"unexpected wind-speed summary %#v",
			result.WindSpeedMetersPerSecond,
		)
	}
	if len(result.Conditions) != 2 ||
		result.DominantCondition == nil ||
		result.DominantCondition.Code != 3 {
		t.Fatalf(
			"unexpected conditions %#v dominant=%#v",
			result.Conditions,
			result.DominantCondition,
		)
	}
	if result.EncounterStartedAt == nil ||
		result.EncounterEndedAt == nil ||
		!result.EncounterStartedAt.Equal(
			result.Points[0].TrajectoryObservedAt,
		) ||
		!result.EncounterEndedAt.Equal(
			result.Points[1].TrajectoryObservedAt,
		) {
		t.Fatalf(
			"unexpected encounter time range: %#v",
			result,
		)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("complete profile is invalid: %v", err)
	}
}

func TestBuildCircularWindDirectionAcrossNorth(t *testing.T) {
	t.Parallel()

	request := validRequest()
	firstDirection := 350.0
	secondDirection := 10.0
	request.Weather.Samples[0].
		Features.WindDirectionDegrees = &firstDirection
	request.Weather.Samples[1].
		Features.WindDirectionDegrees = &secondDirection

	result, err := Build(request)
	if err != nil {
		t.Fatalf("build weather encounter: %v", err)
	}
	if result.WindDirectionDegrees.MeanDirectionDegrees == nil ||
		result.WindDirectionDegrees.Concentration == nil {
		t.Fatal("circular wind summary is missing")
	}

	mean := *result.WindDirectionDegrees.MeanDirectionDegrees
	if !(mean < 1 || mean > 359) {
		t.Fatalf(
			"expected circular mean near north, got %f",
			mean,
		)
	}
	if math.Abs(mean-180) < 1 {
		t.Fatalf(
			"wind direction used arithmetic mean: %f",
			mean,
		)
	}
	if *result.WindDirectionDegrees.Concentration < 0.98 {
		t.Fatalf(
			"unexpected circular concentration %#v",
			result.WindDirectionDegrees,
		)
	}
}

func TestBuildPartialAlignmentIsLimited(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.Alignment.Status = weatheralignment.StatusLimited
	request.Alignment.PointCount = 3
	request.Alignment.UnmatchedCount = 1
	request.Alignment.CoverageRatio = 2.0 / 3.0
	request.Alignment.Matches = append(
		request.Alignment.Matches,
		unmatchedAlignmentMatch(2),
	)
	request.Alignment.Limitations = []weatheralignment.Notice{
		{
			Code:    "partial",
			Message: "Alignment is partial.",
		},
	}

	result, err := Build(request)
	if err != nil {
		t.Fatalf(
			"build partial weather encounter: %v",
			err,
		)
	}
	if result.Status != StatusLimited ||
		result.EncounterPointCount != 2 ||
		result.UnprofiledPointCount != 1 ||
		!hasNotice(
			result.Limitations,
			"weather_alignment_not_complete",
		) {
		t.Fatalf("unexpected limited profile %#v", result)
	}
}

func TestBuildUnavailableAlignment(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.Alignment.Status = weatheralignment.StatusUnavailable
	request.Alignment.AlignedCount = 0
	request.Alignment.UnmatchedCount = 2
	request.Alignment.CoverageRatio = 0
	request.Alignment.Matches = []weatheralignment.Match{
		unmatchedAlignmentMatch(0),
		unmatchedAlignmentMatch(1),
	}
	request.Alignment.Limitations = []weatheralignment.Notice{
		{
			Code:    "unavailable",
			Message: "Alignment is unavailable.",
		},
	}

	result, err := Build(request)
	if err != nil {
		t.Fatalf("build unavailable profile: %v", err)
	}
	if result.Status != StatusUnavailable ||
		result.EncounterPointCount != 0 ||
		result.UnprofiledPointCount != 2 ||
		result.EncounterStartedAt != nil ||
		result.EncounterEndedAt != nil {
		t.Fatalf(
			"unexpected unavailable profile %#v",
			result,
		)
	}
}

func TestBuildMissingCoreMetricIsLimited(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.Weather.Samples[0].
		Features.WindSpeedMetersPerSecond = nil
	request.Weather.Samples[1].
		Features.WindSpeedMetersPerSecond = nil

	result, err := Build(request)
	if err != nil {
		t.Fatalf("build weather encounter: %v", err)
	}
	if result.Status != StatusLimited ||
		result.WindSpeedMetersPerSecond.PresentCount != 0 ||
		!hasNotice(
			result.Limitations,
			"wind_speed_coverage_below_complete_threshold",
		) {
		t.Fatalf(
			"missing core metric did not limit profile %#v",
			result,
		)
	}
}

func TestBuildTrajectoryPointWeighting(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.Alignment.Matches[1].WeatherSampleSequence =
		intPointer(0)
	request.Alignment.Matches[1].WeatherValidAt =
		timePointer(request.Weather.Samples[0].ValidAt)

	result, err := Build(request)
	if err != nil {
		t.Fatalf("build weighted profile: %v", err)
	}
	expected := *request.Weather.Samples[0].
		Features.TemperatureCelsius
	if result.TemperatureCelsius.Mean == nil ||
		*result.TemperatureCelsius.Mean != expected {
		t.Fatalf(
			"sample was not weighted by encounter points: %#v",
			result.TemperatureCelsius,
		)
	}
	if len(result.Conditions) != 1 ||
		result.Conditions[0].Count != 2 ||
		result.Conditions[0].Share != 1 {
		t.Fatalf(
			"condition weighting is invalid %#v",
			result.Conditions,
		)
	}
}

func TestBuildRejectsInputMismatch(t *testing.T) {
	t.Parallel()

	missingSample := validRequest()
	missingSample.Alignment.Matches[0].
		WeatherSampleSequence = intPointer(99)

	_, err := Build(missingSample)
	if !errors.Is(err, ErrInputMismatch) {
		t.Fatalf(
			"expected missing-sample mismatch, got %v",
			err,
		)
	}

	validTimeMismatch := validRequest()
	validTimeMismatch.Alignment.Matches[0].
		WeatherValidAt = timePointer(
		validTimeMismatch.Weather.Samples[0].
			ValidAt.Add(time.Second),
	)

	_, err = Build(validTimeMismatch)
	if !errors.Is(err, ErrInputMismatch) {
		t.Fatalf(
			"expected valid-time mismatch, got %v",
			err,
		)
	}

	trajectoryMismatch := validRequest()
	trajectoryMismatch.Alignment.TrajectoryID = "other"

	_, err = Build(trajectoryMismatch)
	if !errors.Is(err, ErrInputMismatch) {
		t.Fatalf(
			"expected trajectory mismatch, got %v",
			err,
		)
	}
}

func TestBuildRejectsInvalidContractAndAlignment(t *testing.T) {
	t.Parallel()

	invalidWeather := validRequest()
	invalidWeather.Weather.ScopeGuard = "operational"
	_, err := Build(invalidWeather)
	if !errors.Is(err, ErrWeatherContractInvalid) {
		t.Fatalf(
			"expected weather contract error, got %v",
			err,
		)
	}

	invalidAlignment := validRequest()
	invalidAlignment.Alignment.InputFingerprint = "invalid"
	_, err = Build(invalidAlignment)
	if !errors.Is(err, ErrAlignmentInvalid) {
		t.Fatalf(
			"expected alignment error, got %v",
			err,
		)
	}
}

func TestBuildFingerprintIsDeterministic(t *testing.T) {
	t.Parallel()

	request := validRequest()
	first, err := Build(request)
	if err != nil {
		t.Fatalf("build first profile: %v", err)
	}
	second, err := Build(request)
	if err != nil {
		t.Fatalf("build second profile: %v", err)
	}
	if first.InputFingerprint != second.InputFingerprint {
		t.Fatal(
			"same encounter inputs produced different fingerprints",
		)
	}

	changed := request
	changed.Policy.MinimumCompleteProfileCoverage = 0.90
	changedResult, err := Build(changed)
	if err != nil {
		t.Fatalf("build changed profile: %v", err)
	}
	if changedResult.InputFingerprint ==
		first.InputFingerprint {
		t.Fatal(
			"changed policy did not change fingerprint",
		)
	}
}

func TestResultCloneIsDeep(t *testing.T) {
	t.Parallel()

	result, err := Build(validRequest())
	if err != nil {
		t.Fatalf("build profile: %v", err)
	}

	cloned := result.Clone()
	*cloned.TemperatureCelsius.Mean = 99
	*cloned.WindDirectionDegrees.MeanDirectionDegrees = 99
	cloned.Conditions[0].Code = 99
	cloned.DominantCondition.Code = 99
	cloned.Points[0].FeatureCount = 99
	cloned.Explanations[0].Code = "changed"

	if *result.TemperatureCelsius.Mean == 99 ||
		*result.WindDirectionDegrees.
			MeanDirectionDegrees == 99 ||
		result.Conditions[0].Code == 99 ||
		result.DominantCondition.Code == 99 ||
		result.Points[0].FeatureCount == 99 ||
		result.Explanations[0].Code == "changed" {
		t.Fatal(
			"weather encounter clone shares backing data",
		)
	}
}

func validRequest() Request {
	asOfTime := time.Date(
		2026,
		time.July,
		16,
		16,
		0,
		0,
		0,
		time.UTC,
	)

	first := sample(
		0,
		asOfTime.Add(-20*time.Minute),
		-42,
		34,
		350,
		3,
	)
	second := sample(
		1,
		asOfTime.Add(-10*time.Minute),
		-40,
		36,
		10,
		61,
	)

	weather := weathercontract.Result{
		SchemaVersion: weathercontract.SchemaVersionV1,
		Status:        weathercontract.ResultStatusComplete,
		TrajectoryID:  "trajectory-1",
		AsOfTime:      asOfTime,
		Samples:       []weathercontract.Sample{first, second},
		Confidence: weathercontract.Confidence{
			Score: 0.90,
			Level: weathercontract.ConfidenceLevelHigh,
			Reasons: []weathercontract.ConfidenceReason{
				{
					Code:         "trusted",
					Message:      "Trusted weather evidence is available.",
					Contribution: 0.90,
				},
			},
		},
		Explanations: []weathercontract.Explanation{
			{
				Code:    "context_only",
				Message: "Weather is contextual evidence.",
			},
		},
		ScopeGuard: weathercontract.ScopeGuardContextOnly,
		Provenance: weathercontract.Provenance{
			InputFingerprint:  testFingerprint("weather"),
			SourceNames:       []string{"test_weather"},
			LatestAvailableAt: second.AvailableAt,
		},
		GeneratedAt: asOfTime.Add(time.Minute),
	}

	alignment := weatheralignment.Result{
		Version:       weatheralignment.Version,
		Status:        weatheralignment.StatusComplete,
		TrajectoryID:  "trajectory-1",
		AsOfTime:      asOfTime,
		TrustDecision: weathertrust.DecisionAllowed,
		TrustScore:    0.90,
		PointCount:    2,
		AlignedCount:  2,
		CoverageRatio: 1,
		Matches: []weatheralignment.Match{
			alignedMatch(
				0,
				"point-1",
				asOfTime.Add(-19*time.Minute),
				first,
			),
			alignedMatch(
				1,
				"point-2",
				asOfTime.Add(-9*time.Minute),
				second,
			),
		},
		Explanations: []weatheralignment.Notice{
			{
				Code:    "aligned",
				Message: "Weather was aligned.",
			},
		},
		InputFingerprint: testFingerprint("alignment"),
		GeneratedAt:      asOfTime.Add(2 * time.Minute),
	}

	return Request{
		Weather:     weather,
		Alignment:   alignment,
		Policy:      DefaultPolicy(),
		GeneratedAt: asOfTime.Add(3 * time.Minute),
	}
}

func sample(
	sequence int,
	validAt time.Time,
	temperatureValue float64,
	windSpeedValue float64,
	windDirectionValue float64,
	conditionCodeValue int,
) weathercontract.Sample {
	altitude := 9000.0
	humidity := 45.0
	precipitation := 0.4
	rain := 0.2
	cloudCover := 65.0
	pressure := 300.0
	windGust := windSpeedValue + 7

	return weathercontract.Sample{
		Sequence: sequence,
		Position: weathercontract.Position{
			Latitude:          40.2,
			Longitude:         49.8,
			AltitudeMeters:    &altitude,
			VerticalReference: weathercontract.VerticalReferenceMeanSeaLevel,
		},
		Source: weathercontract.Source{
			Provider:     "test_weather",
			Dataset:      "flight_level",
			EvidenceKind: weathercontract.EvidenceKindAnalysis,
		},
		Features: weathercontract.FeatureVector{
			TemperatureCelsius:       floatPointer(temperatureValue),
			RelativeHumidityPercent:  &humidity,
			PrecipitationMillimeters: &precipitation,
			RainMillimeters:          &rain,
			CloudCoverPercent:        &cloudCover,
			SurfacePressureHPA:       &pressure,
			WindSpeedMetersPerSecond: floatPointer(windSpeedValue),
			WindDirectionDegrees:     floatPointer(windDirectionValue),
			WindGustsMetersPerSecond: &windGust,
			ConditionCode:            intPointer(conditionCodeValue),
			ConditionCodeScheme:      "test_scheme",
		},
		ValidAt:     validAt,
		AvailableAt: validAt.Add(time.Minute),
		RetrievedAt: validAt.Add(2 * time.Minute),
	}
}

func alignedMatch(
	sequence int,
	pointID string,
	observedAt time.Time,
	sample weathercontract.Sample,
) weatheralignment.Match {
	weatherSequence := sample.Sequence
	weatherValidAt := sample.ValidAt
	altitude := 9000.0
	horizontal := 1.0
	temporal := time.Minute
	vertical := 100.0

	return weatheralignment.Match{
		TrajectoryPointSequence:      sequence,
		TrajectoryPointID:            pointID,
		TrajectoryObservedAt:         observedAt,
		WeatherSampleSequence:        &weatherSequence,
		WeatherValidAt:               &weatherValidAt,
		Status:                       weatheralignment.MatchStatusAligned,
		AltitudeBasis:                weatheralignment.AltitudeBasisGeometric,
		AltitudeMeters:               &altitude,
		HorizontalDistanceKilometers: &horizontal,
		TemporalDistance:             &temporal,
		VerticalDistanceMeters:       &vertical,
		Score:                        0.90,
		Components: []weatheralignment.Component{
			{
				Name:   weatheralignment.ComponentHorizontal,
				Score:  0.90,
				Weight: 0.45,
			},
			{
				Name:   weatheralignment.ComponentTemporal,
				Score:  0.90,
				Weight: 0.35,
			},
			{
				Name:   weatheralignment.ComponentVertical,
				Score:  0.90,
				Weight: 0.20,
			},
		},
	}
}

func unmatchedAlignmentMatch(
	sequence int,
) weatheralignment.Match {
	return weatheralignment.Match{
		TrajectoryPointSequence: sequence,
		TrajectoryPointID:       "point-unmatched",
		TrajectoryObservedAt: time.Date(
			2026,
			time.July,
			16,
			15,
			55,
			0,
			0,
			time.UTC,
		),
		Status:        weatheralignment.MatchStatusUnmatched,
		AltitudeBasis: weatheralignment.AltitudeBasisUnavailable,
		Score:         0,
		Components: []weatheralignment.Component{
			{
				Name:   weatheralignment.ComponentHorizontal,
				Score:  0,
				Weight: 0.45,
			},
			{
				Name:   weatheralignment.ComponentTemporal,
				Score:  0,
				Weight: 0.35,
			},
			{
				Name:   weatheralignment.ComponentVertical,
				Score:  0,
				Weight: 0.20,
			},
		},
		Limitations: []weatheralignment.Notice{
			{
				Code:    "unmatched",
				Message: "Point is unmatched.",
			},
		},
	}
}

func hasNotice(
	notices []Notice,
	code string,
) bool {
	for _, notice := range notices {
		if notice.Code == code {
			return true
		}
	}
	return false
}

func floatPointer(value float64) *float64 {
	return &value
}

func intPointer(value int) *int {
	return &value
}

func timePointer(value time.Time) *time.Time {
	copied := value.UTC()
	return &copied
}

func testFingerprint(seed string) string {
	if seed == "alignment" {
		return "sha256:" +
			"abcdef0123456789" +
			"abcdef0123456789" +
			"abcdef0123456789" +
			"abcdef0123456789"
	}
	return "sha256:" +
		"0123456789abcdef" +
		"0123456789abcdef" +
		"0123456789abcdef" +
		"0123456789abcdef"
}
