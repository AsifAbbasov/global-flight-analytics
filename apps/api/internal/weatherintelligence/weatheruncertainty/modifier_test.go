package weatheruncertainty

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatherencounter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
)

func TestDefaultPolicyIsValid(t *testing.T) {
	t.Parallel()

	if err := DefaultPolicy().Validate(); err != nil {
		t.Fatalf("default policy is invalid: %v", err)
	}
}

func TestApplyPreservesCoordinatesAndNeverReducesUncertainty(t *testing.T) {
	t.Parallel()

	request := validRequest()
	result, err := Apply(request)
	if err != nil {
		t.Fatalf("apply weather uncertainty: %v", err)
	}

	if result.Status != StatusApplied ||
		len(result.PointAdjustments) != len(request.Projection.Points) {
		t.Fatalf("unexpected applied result %#v", result)
	}

	for index, point := range result.AdjustedProjection.Points {
		original := request.Projection.Points[index]

		if point.Position.Latitude != original.Position.Latitude ||
			point.Position.Longitude != original.Position.Longitude ||
			!equalOptionalFloat(point.Position.AltitudeM, original.Position.AltitudeM) {
			t.Fatalf("projection coordinates changed at point %d", index)
		}

		if point.Uncertainty.HorizontalRadiusM < original.Uncertainty.HorizontalRadiusM {
			t.Fatalf("horizontal uncertainty was reduced at point %d", index)
		}

		if point.Uncertainty.VerticalRadiusM == nil ||
			original.Uncertainty.VerticalRadiusM == nil ||
			*point.Uncertainty.VerticalRadiusM < *original.Uncertainty.VerticalRadiusM {
			t.Fatalf("vertical uncertainty was reduced at point %d", index)
		}

		if point.Confidence.Score > original.Confidence.Score {
			t.Fatalf("confidence increased at point %d", index)
		}
	}

	if result.PointAdjustments[1].Multiplier <= result.PointAdjustments[0].Multiplier {
		t.Fatal("farther-horizon point did not receive a stronger weather effect")
	}

	if result.AdjustedProjection.Provenance.InputFingerprint != result.InputFingerprint {
		t.Fatal("adjusted projection fingerprint was not replaced")
	}

	if err := result.Validate(); err != nil {
		t.Fatalf("applied result is invalid: %v", err)
	}
}

func TestApplySevereWeatherIncreasesMultiplier(t *testing.T) {
	t.Parallel()

	calm := validRequest()
	calm.Encounter.WindSpeedMetersPerSecond = metricSummary(8, 9, 8.5)
	calm.Encounter.WindGustsMetersPerSecond = metricSummary(10, 12, 11)
	calm.Encounter.PrecipitationMillimeters = metricSummary(0, 0.1, 0.05)
	calm.Encounter.RainMillimeters = metricSummary(0, 0.1, 0.05)
	calm.Encounter.CloudCoverPercent = metricSummary(10, 25, 18)

	calmResult, err := Apply(calm)
	if err != nil {
		t.Fatalf("apply calm weather: %v", err)
	}

	severe := validRequest()
	severe.Encounter.WindSpeedMetersPerSecond = metricSummary(30, 40, 35)
	severe.Encounter.WindGustsMetersPerSecond = metricSummary(45, 60, 52)
	severe.Encounter.PrecipitationMillimeters = metricSummary(3, 8, 5)
	severe.Encounter.RainMillimeters = metricSummary(2, 7, 4)
	severe.Encounter.CloudCoverPercent = metricSummary(85, 100, 95)

	severeResult, err := Apply(severe)
	if err != nil {
		t.Fatalf("apply severe weather: %v", err)
	}

	if severeResult.WeatherMultiplier <= calmResult.WeatherMultiplier {
		t.Fatalf(
			"severe multiplier %f did not exceed calm multiplier %f",
			severeResult.WeatherMultiplier,
			calmResult.WeatherMultiplier,
		)
	}

	if severeResult.WeatherMultiplier > severe.Policy.MaximumUncertaintyMultiplier {
		t.Fatalf(
			"weather multiplier exceeded policy maximum: %f",
			severeResult.WeatherMultiplier,
		)
	}
}

func TestApplyBlockedTrustIsWithheld(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.Trust.Decision = weathertrust.DecisionBlocked
	request.Trust.Usable = false
	request.Trust.AllowedScopes = nil
	request.Trust.Limitations = []weathertrust.Notice{
		{
			Code:    "blocked",
			Message: "Weather is blocked.",
		},
	}

	result, err := Apply(request)
	if err != nil {
		t.Fatalf("apply blocked trust: %v", err)
	}

	if result.Status != StatusWithheld ||
		result.WeatherMultiplier != 1 ||
		len(result.PointAdjustments) != 0 ||
		!projectionsEqual(result.AdjustedProjection, request.Projection) {
		t.Fatalf("blocked trust did not preserve projection: %#v", result)
	}
}

func TestApplySurfaceOnlyScopeIsWithheld(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.Trust.Decision = weathertrust.DecisionLimited
	request.Trust.AllowedScopes = []weathertrust.UsageScope{
		weathertrust.UsageScopeSurfaceContext,
	}
	request.Trust.Limitations = []weathertrust.Notice{
		{
			Code:    "surface_only",
			Message: "Weather is surface-only.",
		},
	}

	result, err := Apply(request)
	if err != nil {
		t.Fatalf("apply surface-only trust: %v", err)
	}

	if result.Status != StatusWithheld ||
		!hasNotice(result.Limitations, "projection_uncertainty_scope_withheld") {
		t.Fatalf("surface-only weather was not withheld: %#v", result)
	}
}

func TestApplyUnavailableEncounterIsWithheld(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.Encounter.Status = weatherencounter.StatusUnavailable
	request.Encounter.EncounterPointCount = 0
	request.Encounter.UnprofiledPointCount = request.Encounter.PointCount
	request.Encounter.ProfileCoverageRatio = 0
	request.Encounter.EncounterStartedAt = nil
	request.Encounter.EncounterEndedAt = nil
	request.Encounter.Points = nil
	request.Encounter.TemperatureCelsius = weatherencounter.MetricSummary{}
	request.Encounter.RelativeHumidityPercent = weatherencounter.MetricSummary{}
	request.Encounter.PrecipitationMillimeters = weatherencounter.MetricSummary{}
	request.Encounter.RainMillimeters = weatherencounter.MetricSummary{}
	request.Encounter.CloudCoverPercent = weatherencounter.MetricSummary{}
	request.Encounter.SurfacePressureHPA = weatherencounter.MetricSummary{}
	request.Encounter.WindSpeedMetersPerSecond = weatherencounter.MetricSummary{}
	request.Encounter.WindDirectionDegrees = weatherencounter.CircularDirectionSummary{}
	request.Encounter.WindGustsMetersPerSecond = weatherencounter.MetricSummary{}
	request.Encounter.Conditions = nil
	request.Encounter.DominantCondition = nil
	request.Encounter.Limitations = []weatherencounter.Notice{
		{
			Code:    "unavailable",
			Message: "Encounter is unavailable.",
		},
	}

	result, err := Apply(request)
	if err != nil {
		t.Fatalf("apply unavailable encounter: %v", err)
	}

	if result.Status != StatusWithheld ||
		!hasNotice(result.Limitations, "weather_encounter_unavailable") {
		t.Fatalf("unavailable encounter was not withheld: %#v", result)
	}
}

func TestApplyLimitedEncounterIsAppliedLimited(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.Encounter.Status = weatherencounter.StatusLimited
	request.Encounter.Limitations = []weatherencounter.Notice{
		{
			Code:    "limited",
			Message: "Encounter coverage is limited.",
		},
	}

	result, err := Apply(request)
	if err != nil {
		t.Fatalf("apply limited encounter: %v", err)
	}

	if result.Status != StatusAppliedLimited ||
		!hasNotice(result.Limitations, "weather_evidence_limited") {
		t.Fatalf("limited evidence status is invalid: %#v", result)
	}
}

func TestApplyWidensArrivalInterval(t *testing.T) {
	t.Parallel()

	request := validRequest()
	result, err := Apply(request)
	if err != nil {
		t.Fatalf("apply weather uncertainty: %v", err)
	}

	if result.ArrivalAdjustment == nil || result.AdjustedProjection.Arrival == nil {
		t.Fatal("arrival adjustment is missing")
	}

	original := request.Projection.Arrival
	adjusted := result.AdjustedProjection.Arrival

	if adjusted.EarliestTime.After(original.EarliestTime) ||
		adjusted.LatestTime.Before(original.LatestTime) ||
		!adjusted.EstimatedTime.Equal(original.EstimatedTime) ||
		adjusted.Confidence.Score > original.Confidence.Score {
		t.Fatalf(
			"arrival interval was not conservatively widened: original=%#v adjusted=%#v",
			original,
			adjusted,
		)
	}
}

func TestApplyUnavailableProjection(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.Projection = unavailableProjection(request.Projection.Horizon.AsOfTime)

	result, err := Apply(request)
	if err != nil {
		t.Fatalf("apply unavailable projection: %v", err)
	}

	if result.Status != StatusUnavailable ||
		result.WeatherMultiplier != 1 ||
		len(result.PointAdjustments) != 0 {
		t.Fatalf("unexpected unavailable result %#v", result)
	}
}

func TestApplyRejectsMismatchedAndInvalidInputs(t *testing.T) {
	t.Parallel()

	mismatched := validRequest()
	mismatched.Encounter.TrajectoryID = "other"
	_, err := Apply(mismatched)
	if !errors.Is(err, ErrInputMismatch) {
		t.Fatalf("expected input mismatch, got %v", err)
	}

	invalidProjection := validRequest()
	invalidProjection.Projection.Points[0].Uncertainty.HorizontalRadiusM = 0
	_, err = Apply(invalidProjection)
	if !errors.Is(err, ErrProjectionInvalid) {
		t.Fatalf("expected invalid projection, got %v", err)
	}

	invalidTrust := validRequest()
	invalidTrust.Trust.InputFingerprint = "invalid"
	_, err = Apply(invalidTrust)
	if !errors.Is(err, ErrTrustInvalid) {
		t.Fatalf("expected invalid trust, got %v", err)
	}

	invalidEncounter := validRequest()
	invalidEncounter.Encounter.InputFingerprint = "invalid"
	_, err = Apply(invalidEncounter)
	if !errors.Is(err, ErrEncounterInvalid) {
		t.Fatalf("expected invalid encounter, got %v", err)
	}
}

func TestApplyRejectsAlreadyAdjustedProjection(t *testing.T) {
	t.Parallel()

	request := validRequest()
	request.Projection.Explanations = append(
		request.Projection.Explanations,
		projectioncontract.Explanation{
			Code:    weatherReasonCode,
			Message: "Already adjusted.",
		},
	)

	_, err := Apply(request)
	if !errors.Is(err, ErrAlreadyAdjusted) {
		t.Fatalf("expected already-adjusted error, got %v", err)
	}
}

func TestApplyFingerprintIsDeterministic(t *testing.T) {
	t.Parallel()

	request := validRequest()
	first, err := Apply(request)
	if err != nil {
		t.Fatalf("apply first result: %v", err)
	}
	second, err := Apply(request)
	if err != nil {
		t.Fatalf("apply second result: %v", err)
	}

	if first.InputFingerprint != second.InputFingerprint {
		t.Fatal("same weather uncertainty inputs produced different fingerprints")
	}

	changed := request
	changed.Policy.MaximumUncertaintyMultiplier = 2.75
	changedResult, err := Apply(changed)
	if err != nil {
		t.Fatalf("apply changed policy: %v", err)
	}

	if changedResult.InputFingerprint == first.InputFingerprint {
		t.Fatal("changed policy did not change fingerprint")
	}
}

func TestApplyDoesNotMutateInputsAndCloneIsDeep(t *testing.T) {
	t.Parallel()

	request := validRequest()
	originalProjection := request.Projection.Clone()

	result, err := Apply(request)
	if err != nil {
		t.Fatalf("apply weather uncertainty: %v", err)
	}

	if !projectionsEqual(request.Projection, originalProjection) {
		t.Fatal("Apply mutated the input projection")
	}

	cloned := result.Clone()
	cloned.Components[0].Score = 0
	cloned.PointAdjustments[0].AdjustedHorizontalRadiusM = 999
	*cloned.PointAdjustments[0].AdjustedVerticalRadiusM = 999
	cloned.AdjustedProjection.Points[0].Uncertainty.HorizontalRadiusM = 999
	cloned.Explanations[0].Code = "changed"

	if result.Components[0].Score == 0 ||
		result.PointAdjustments[0].AdjustedHorizontalRadiusM == 999 ||
		*result.PointAdjustments[0].AdjustedVerticalRadiusM == 999 ||
		result.AdjustedProjection.Points[0].Uncertainty.HorizontalRadiusM == 999 ||
		result.Explanations[0].Code == "changed" {
		t.Fatal("weather uncertainty clone shares backing data")
	}
}

func validRequest() Request {
	asOfTime := time.Date(2026, time.July, 16, 18, 0, 0, 0, time.UTC)

	return Request{
		Projection:  validProjection(asOfTime),
		Trust:       validTrust(asOfTime),
		Encounter:   validEncounter(asOfTime),
		Policy:      DefaultPolicy(),
		GeneratedAt: asOfTime.Add(5 * time.Minute),
	}
}

func validProjection(asOfTime time.Time) projectioncontract.Result {
	firstAltitude := 9000.0
	secondAltitude := 9100.0
	firstVertical := 400.0
	secondVertical := 600.0

	return projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        projectioncontract.ResultStatusComplete,
		TrajectoryID:  "trajectory-1",
		FlightID:      "flight-1",
		AircraftID:    "aircraft-1",
		ICAO24:        "abc123",
		Callsign:      "TEST123",
		Method: projectioncontract.Method{
			Name:          "test_projection",
			Version:       "v1",
			DecisionClass: projectioncontract.DecisionClassProjectDerived,
		},
		Horizon: projectioncontract.Horizon{
			AsOfTime: asOfTime,
			EndTime:  asOfTime.Add(20 * time.Minute),
			Step:     10 * time.Minute,
		},
		Points: []projectioncontract.ProjectionPoint{
			{
				Sequence:     0,
				ForecastTime: asOfTime.Add(10 * time.Minute),
				Position: projectioncontract.Position{
					Latitude:  40.3,
					Longitude: 49.9,
					AltitudeM: &firstAltitude,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 1000,
					VerticalRadiusM:   &firstVertical,
				},
				Confidence: projectionConfidence(0.85),
			},
			{
				Sequence:     1,
				ForecastTime: asOfTime.Add(20 * time.Minute),
				Position: projectioncontract.Position{
					Latitude:  40.5,
					Longitude: 50.1,
					AltitudeM: &secondAltitude,
				},
				Uncertainty: projectioncontract.Uncertainty{
					HorizontalRadiusM: 1800,
					VerticalRadiusM:   &secondVertical,
				},
				Confidence: projectionConfidence(0.75),
			},
		},
		Arrival: &projectioncontract.ArrivalEstimate{
			AirportICAOCode: "UBBB",
			EarliestTime:    asOfTime.Add(14 * time.Minute),
			EstimatedTime:   asOfTime.Add(16 * time.Minute),
			LatestTime:      asOfTime.Add(18 * time.Minute),
			Confidence:      projectionConfidence(0.75),
		},
		Confidence: projectionConfidence(0.80),
		Explanations: []projectioncontract.Explanation{
			{
				Code:    "projection",
				Message: "Projection is available.",
			},
		},
		ScopeGuard: projectioncontract.ScopeGuardResearchOnly,
		Provenance: projectioncontract.Provenance{
			InputFingerprint: testFingerprint("projection"),
			Inputs: []projectioncontract.InputReference{
				{
					Name:           "trajectory",
					Classification: projectioncontract.InputClassificationObserved,
					ObservedAt:     asOfTime.Add(-time.Minute),
					RetrievedAt:    asOfTime,
				},
			},
			LatestInputObservedAt: asOfTime.Add(-time.Minute),
		},
		GeneratedAt: asOfTime.Add(time.Minute),
	}
}

func unavailableProjection(asOfTime time.Time) projectioncontract.Result {
	return projectioncontract.Result{
		SchemaVersion: projectioncontract.SchemaVersionV1,
		Status:        projectioncontract.ResultStatusUnavailable,
		TrajectoryID:  "trajectory-1",
		Method: projectioncontract.Method{
			Name:          "test_projection",
			Version:       "v1",
			DecisionClass: projectioncontract.DecisionClassProjectDerived,
		},
		Horizon: projectioncontract.Horizon{
			AsOfTime: asOfTime,
			EndTime:  asOfTime.Add(20 * time.Minute),
			Step:     10 * time.Minute,
		},
		Confidence: projectioncontract.Confidence{
			Score: 0,
			Level: projectioncontract.ConfidenceLevelNone,
		},
		Limitations: []projectioncontract.Limitation{
			{
				Code:    "unavailable",
				Message: "Projection is unavailable.",
				Scope:   "projection",
			},
		},
		ScopeGuard:  projectioncontract.ScopeGuardResearchOnly,
		GeneratedAt: asOfTime.Add(time.Minute),
	}
}

func projectionConfidence(score float64) projectioncontract.Confidence {
	level := projectioncontract.ConfidenceLevelHigh
	if score < 0.80 {
		level = projectioncontract.ConfidenceLevelMedium
	}

	return projectioncontract.Confidence{
		Score: score,
		Level: level,
		Reasons: []projectioncontract.ConfidenceReason{
			{
				Code:         "baseline",
				Message:      "Baseline confidence.",
				Contribution: score,
			},
		},
	}
}

func validTrust(asOfTime time.Time) weathertrust.Result {
	return weathertrust.Result{
		Version:  weathertrust.Version,
		Decision: weathertrust.DecisionAllowed,
		Usable:   true,
		AsOfTime: asOfTime,
		Score:    0.90,
		Components: []weathertrust.Component{
			{
				Name:   weathertrust.ComponentContractConfidence,
				Score:  0.90,
				Weight: 0.35,
			},
			{
				Name:   weathertrust.ComponentTemporalFreshness,
				Score:  0.90,
				Weight: 0.30,
			},
			{
				Name:   weathertrust.ComponentFeatureCompleteness,
				Score:  1,
				Weight: 0.20,
			},
			{
				Name:   weathertrust.ComponentVerticalApplicability,
				Score:  1,
				Weight: 0.15,
			},
		},
		AllowedScopes: []weathertrust.UsageScope{
			weathertrust.UsageScopeProjectionUncertainty,
			weathertrust.UsageScopeTrajectoryContext,
		},
		Explanations: []weathertrust.Notice{
			{
				Code:    "context",
				Message: "Weather is contextual.",
			},
		},
		InputFingerprint: testFingerprint("trust"),
	}
}

func validEncounter(asOfTime time.Time) weatherencounter.Result {
	startedAt := asOfTime.Add(-20 * time.Minute)
	endedAt := asOfTime.Add(-5 * time.Minute)

	return weatherencounter.Result{
		Version:                  weatherencounter.Version,
		Status:                   weatherencounter.StatusComplete,
		TrajectoryID:             "trajectory-1",
		AsOfTime:                 asOfTime,
		AlignmentStatus:          "complete",
		AlignmentCoverageRatio:   1,
		PointCount:               2,
		EncounterPointCount:      2,
		UnprofiledPointCount:     0,
		ProfileCoverageRatio:     1,
		EncounterStartedAt:       &startedAt,
		EncounterEndedAt:         &endedAt,
		TemperatureCelsius:       metricSummary(-42, -40, -41),
		RelativeHumidityPercent:  metricSummary(40, 50, 45),
		PrecipitationMillimeters: metricSummary(0.5, 1.5, 1),
		RainMillimeters:          metricSummary(0.2, 1, 0.6),
		CloudCoverPercent:        metricSummary(50, 80, 65),
		SurfacePressureHPA:       metricSummary(290, 310, 300),
		WindSpeedMetersPerSecond: metricSummary(18, 28, 23),
		WindDirectionDegrees:     directionSummary(250),
		WindGustsMetersPerSecond: metricSummary(25, 40, 32),
		Points: []weatherencounter.EncounterPoint{
			{
				TrajectoryPointSequence: 0,
				TrajectoryPointID:       "point-1",
				TrajectoryObservedAt:    startedAt,
				WeatherSampleSequence:   0,
				WeatherValidAt:          startedAt,
				AlignmentScore:          0.90,
				FeatureCount:            9,
			},
			{
				TrajectoryPointSequence: 1,
				TrajectoryPointID:       "point-2",
				TrajectoryObservedAt:    endedAt,
				WeatherSampleSequence:   1,
				WeatherValidAt:          endedAt,
				AlignmentScore:          0.90,
				FeatureCount:            9,
			},
		},
		Explanations: []weatherencounter.Notice{
			{
				Code:    "profile",
				Message: "Encounter profile is available.",
			},
		},
		InputFingerprint: testFingerprint("encounter"),
		GeneratedAt:      asOfTime.Add(3 * time.Minute),
	}
}

func metricSummary(
	minimumValue float64,
	maximumValue float64,
	meanValue float64,
) weatherencounter.MetricSummary {
	return weatherencounter.MetricSummary{
		PresentCount:  2,
		CoverageRatio: 1,
		Minimum:       floatPointer(minimumValue),
		Maximum:       floatPointer(maximumValue),
		Mean:          floatPointer(meanValue),
	}
}

func directionSummary(value float64) weatherencounter.CircularDirectionSummary {
	concentration := 0.90
	return weatherencounter.CircularDirectionSummary{
		PresentCount:         2,
		CoverageRatio:        1,
		MeanDirectionDegrees: &value,
		Concentration:        &concentration,
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

func equalOptionalFloat(left *float64, right *float64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func projectionsEqual(
	left projectioncontract.Result,
	right projectioncontract.Result,
) bool {
	if left.Provenance.InputFingerprint != right.Provenance.InputFingerprint ||
		!left.GeneratedAt.Equal(right.GeneratedAt) ||
		len(left.Points) != len(right.Points) {
		return false
	}

	for index := range left.Points {
		leftPoint := left.Points[index]
		rightPoint := right.Points[index]

		if leftPoint.Position.Latitude != rightPoint.Position.Latitude ||
			leftPoint.Position.Longitude != rightPoint.Position.Longitude ||
			leftPoint.Uncertainty.HorizontalRadiusM != rightPoint.Uncertainty.HorizontalRadiusM ||
			leftPoint.Confidence.Score != rightPoint.Confidence.Score {
			return false
		}
	}
	return true
}

func floatPointer(value float64) *float64 {
	return &value
}

func testFingerprint(seed string) string {
	switch seed {
	case "trust":
		return "sha256:" +
			"abcdef0123456789" +
			"abcdef0123456789" +
			"abcdef0123456789" +
			"abcdef0123456789"
	case "encounter":
		return "sha256:" +
			"fedcba9876543210" +
			"fedcba9876543210" +
			"fedcba9876543210" +
			"fedcba9876543210"
	default:
		return "sha256:" +
			"0123456789abcdef" +
			"0123456789abcdef" +
			"0123456789abcdef" +
			"0123456789abcdef"
	}
}
