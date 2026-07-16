package weatherencounter

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathercontract"
)

var (
	ErrInvalidPolicy = errors.New(
		"weather encounter policy is invalid",
	)
	ErrWeatherContractInvalid = errors.New(
		"weather encounter weather contract is invalid",
	)
	ErrAlignmentInvalid = errors.New(
		"weather encounter alignment is invalid",
	)
	ErrInputMismatch = errors.New(
		"weather encounter inputs do not describe the same evidence boundary",
	)
	ErrResultInvalid = errors.New(
		"weather encounter result is invalid",
	)
)

type Request struct {
	Weather     weathercontract.Result
	Alignment   weatheralignment.Result
	Policy      Policy
	GeneratedAt time.Time
}

func Build(
	request Request,
) (Result, error) {
	if err := request.Policy.Validate(); err != nil {
		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrInvalidPolicy,
			err,
		)
	}

	weatherReport := weathercontract.Validate(
		request.Weather,
	)
	if weatherReport.Status !=
		weathercontract.ValidationStatusValid {
		return Result{}, fmt.Errorf(
			"%w: issues=%v",
			ErrWeatherContractInvalid,
			weatherReport.Issues,
		)
	}
	if err := request.Alignment.Validate(); err != nil {
		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrAlignmentInvalid,
			err,
		)
	}
	if err := validateInputBoundary(
		request.Weather,
		request.Alignment,
	); err != nil {
		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrInputMismatch,
			err,
		)
	}

	generatedAt := request.GeneratedAt.UTC()
	if generatedAt.IsZero() ||
		generatedAt.Before(
			request.Weather.AsOfTime,
		) {
		return Result{}, fmt.Errorf(
			"%w: generated-at time is invalid",
			ErrInputMismatch,
		)
	}

	result := Result{
		Version: Version,
		TrajectoryID: strings.TrimSpace(
			request.Weather.TrajectoryID,
		),
		AsOfTime: request.Weather.
			AsOfTime.UTC(),
		AlignmentStatus:        request.Alignment.Status,
		AlignmentCoverageRatio: request.Alignment.CoverageRatio,
		PointCount:             request.Alignment.PointCount,
		Explanations: []Notice{
			{
				Code:    "weather_encounter_profile",
				Message: "The profile summarizes weather features attached to aligned trajectory points.",
			},
			{
				Code:    "trajectory_point_weighted",
				Message: "Repeated use of a weather sample is weighted by the number of aligned trajectory points that encountered that sample.",
			},
			{
				Code:    "weather_context_only",
				Message: "The profile describes contextual evidence and does not prove pilot intent, controller intent, rerouting reason, or maneuver cause.",
			},
		},
		GeneratedAt: generatedAt,
	}

	if request.Alignment.Status ==
		weatheralignment.StatusUnavailable ||
		request.Alignment.AlignedCount == 0 {
		result.Status = StatusUnavailable
		result.UnprofiledPointCount =
			result.PointCount
		result.Limitations = []Notice{
			{
				Code:    "weather_alignment_unavailable",
				Message: "No aligned trajectory point is available for a Weather Encounter Profile.",
			},
		}
		result.InputFingerprint = inputFingerprint(
			request.Weather,
			request.Alignment,
			request.Policy,
			nil,
		)
		return validateAndClone(result)
	}

	samplesBySequence := make(
		map[int]weathercontract.Sample,
		len(request.Weather.Samples),
	)
	for _, sample := range request.Weather.Samples {
		samplesBySequence[sample.Sequence] =
			sample
	}

	accumulators := profileAccumulators{
		conditions: make(
			map[conditionKey]int,
		),
	}

	for _, match := range request.Alignment.Matches {
		if match.Status !=
			weatheralignment.MatchStatusAligned {
			continue
		}
		if match.WeatherSampleSequence == nil ||
			match.WeatherValidAt == nil {
			return Result{}, fmt.Errorf(
				"%w: aligned match does not identify weather evidence",
				ErrInputMismatch,
			)
		}

		sample, exists := samplesBySequence[*match.WeatherSampleSequence]
		if !exists {
			return Result{}, fmt.Errorf(
				"%w: alignment references missing weather sample %d",
				ErrInputMismatch,
				*match.WeatherSampleSequence,
			)
		}
		if !match.WeatherValidAt.Equal(
			sample.ValidAt,
		) {
			return Result{}, fmt.Errorf(
				"%w: alignment and weather valid times differ for sample %d",
				ErrInputMismatch,
				sample.Sequence,
			)
		}

		point := EncounterPoint{
			TrajectoryPointSequence: match.TrajectoryPointSequence,
			TrajectoryPointID: strings.TrimSpace(
				match.TrajectoryPointID,
			),
			TrajectoryObservedAt:  match.TrajectoryObservedAt.UTC(),
			WeatherSampleSequence: sample.Sequence,
			WeatherValidAt:        sample.ValidAt.UTC(),
			AlignmentScore:        match.Score,
			FeatureCount:          sample.Features.PresentCount(),
		}
		result.Points = append(
			result.Points,
			point,
		)
		accumulators.add(sample.Features)
	}

	sort.Slice(
		result.Points,
		func(left int, right int) bool {
			return result.Points[left].
				TrajectoryPointSequence <
				result.Points[right].
					TrajectoryPointSequence
		},
	)

	result.EncounterPointCount =
		len(result.Points)
	result.UnprofiledPointCount =
		result.PointCount -
			result.EncounterPointCount
	if result.PointCount > 0 {
		result.ProfileCoverageRatio =
			float64(
				result.EncounterPointCount,
			) /
				float64(result.PointCount)
	}

	if result.EncounterPointCount == 0 {
		result.Status = StatusUnavailable
		result.Limitations = []Notice{
			{
				Code:    "weather_encounter_points_unavailable",
				Message: "Alignment reported evidence, but no auditable weather encounter point could be built.",
			},
		}
		result.InputFingerprint = inputFingerprint(
			request.Weather,
			request.Alignment,
			request.Policy,
			result.Points,
		)
		return validateAndClone(result)
	}

	startedAt := result.Points[0].
		TrajectoryObservedAt.UTC()
	endedAt := result.Points[len(result.Points)-1].TrajectoryObservedAt.UTC()
	result.EncounterStartedAt = &startedAt
	result.EncounterEndedAt = &endedAt

	accumulators.apply(&result)

	if request.Policy.complete(
		string(result.AlignmentStatus),
		result.ProfileCoverageRatio,
		result.TemperatureCelsius.
			CoverageRatio,
		result.WindSpeedMetersPerSecond.
			CoverageRatio,
		result.WindDirectionDegrees.
			CoverageRatio,
	) {
		result.Status = StatusComplete
	} else {
		result.Status = StatusLimited
		addCompletenessLimitations(
			&result,
			request.Policy,
		)
	}

	for _, limitation := range request.Alignment.Limitations {
		result.Limitations = append(
			result.Limitations,
			Notice{
				Code: "alignment_" +
					limitation.Code,
				Message: limitation.Message,
			},
		)
	}
	result.Limitations = normalizeNotices(
		result.Limitations,
	)
	result.InputFingerprint = inputFingerprint(
		request.Weather,
		request.Alignment,
		request.Policy,
		result.Points,
	)

	return validateAndClone(result)
}

func validateInputBoundary(
	weather weathercontract.Result,
	alignment weatheralignment.Result,
) error {
	if strings.TrimSpace(
		weather.TrajectoryID,
	) != strings.TrimSpace(
		alignment.TrajectoryID,
	) {
		return fmt.Errorf(
			"trajectory identifiers differ",
		)
	}
	if !weather.AsOfTime.Equal(
		alignment.AsOfTime,
	) {
		return fmt.Errorf(
			"as-of times differ",
		)
	}
	if alignment.AlignedCount >
		len(alignment.Matches) {
		return fmt.Errorf(
			"alignment count exceeds match count",
		)
	}
	return nil
}

type metricAccumulator struct {
	count int
	sum   float64
	min   float64
	max   float64
}

func (accumulator *metricAccumulator) add(
	value *float64,
) {
	if value == nil {
		return
	}
	if accumulator.count == 0 {
		accumulator.min = *value
		accumulator.max = *value
	} else {
		if *value < accumulator.min {
			accumulator.min = *value
		}
		if *value > accumulator.max {
			accumulator.max = *value
		}
	}
	accumulator.count++
	accumulator.sum += *value
}

func (accumulator metricAccumulator) summary(
	denominator int,
) MetricSummary {
	summary := MetricSummary{
		PresentCount: accumulator.count,
	}
	if denominator > 0 {
		summary.CoverageRatio =
			float64(accumulator.count) /
				float64(denominator)
	}
	if accumulator.count == 0 {
		return summary
	}

	minimum := accumulator.min
	maximum := accumulator.max
	mean := accumulator.sum /
		float64(accumulator.count)
	summary.Minimum = &minimum
	summary.Maximum = &maximum
	summary.Mean = &mean
	return summary
}

type circularAccumulator struct {
	count  int
	sine   float64
	cosine float64
}

func (accumulator *circularAccumulator) add(
	value *float64,
) {
	if value == nil {
		return
	}
	radians := *value * math.Pi / 180
	accumulator.sine += math.Sin(radians)
	accumulator.cosine += math.Cos(radians)
	accumulator.count++
}

func (accumulator circularAccumulator) summary(
	denominator int,
) CircularDirectionSummary {
	summary := CircularDirectionSummary{
		PresentCount: accumulator.count,
	}
	if denominator > 0 {
		summary.CoverageRatio =
			float64(accumulator.count) /
				float64(denominator)
	}
	if accumulator.count == 0 {
		return summary
	}

	meanRadians := math.Atan2(
		accumulator.sine,
		accumulator.cosine,
	)
	meanDegrees :=
		meanRadians * 180 / math.Pi
	if meanDegrees < 0 {
		meanDegrees += 360
	}
	if meanDegrees >= 360 {
		meanDegrees -= 360
	}
	concentration := math.Sqrt(
		accumulator.sine*
			accumulator.sine+
			accumulator.cosine*
				accumulator.cosine,
	) / float64(accumulator.count)
	concentration = math.Min(
		1,
		math.Max(0, concentration),
	)

	summary.MeanDirectionDegrees =
		&meanDegrees
	summary.Concentration = &concentration
	return summary
}

type conditionKey struct {
	scheme string
	code   int
}

type profileAccumulators struct {
	temperature   metricAccumulator
	humidity      metricAccumulator
	precipitation metricAccumulator
	rain          metricAccumulator
	cloudCover    metricAccumulator
	pressure      metricAccumulator
	windSpeed     metricAccumulator
	windDirection circularAccumulator
	windGusts     metricAccumulator
	conditions    map[conditionKey]int
}

func (accumulators *profileAccumulators) add(
	features weathercontract.FeatureVector,
) {
	accumulators.temperature.add(
		features.TemperatureCelsius,
	)
	accumulators.humidity.add(
		features.RelativeHumidityPercent,
	)
	accumulators.precipitation.add(
		features.PrecipitationMillimeters,
	)
	accumulators.rain.add(
		features.RainMillimeters,
	)
	accumulators.cloudCover.add(
		features.CloudCoverPercent,
	)
	accumulators.pressure.add(
		features.SurfacePressureHPA,
	)
	accumulators.windSpeed.add(
		features.WindSpeedMetersPerSecond,
	)
	accumulators.windDirection.add(
		features.WindDirectionDegrees,
	)
	accumulators.windGusts.add(
		features.WindGustsMetersPerSecond,
	)

	if features.ConditionCode != nil {
		key := conditionKey{
			scheme: strings.TrimSpace(
				features.ConditionCodeScheme,
			),
			code: *features.ConditionCode,
		}
		accumulators.conditions[key]++
	}
}

func (accumulators profileAccumulators) apply(
	result *Result,
) {
	denominator := result.EncounterPointCount
	result.TemperatureCelsius =
		accumulators.temperature.summary(
			denominator,
		)
	result.RelativeHumidityPercent =
		accumulators.humidity.summary(
			denominator,
		)
	result.PrecipitationMillimeters =
		accumulators.precipitation.summary(
			denominator,
		)
	result.RainMillimeters =
		accumulators.rain.summary(
			denominator,
		)
	result.CloudCoverPercent =
		accumulators.cloudCover.summary(
			denominator,
		)
	result.SurfacePressureHPA =
		accumulators.pressure.summary(
			denominator,
		)
	result.WindSpeedMetersPerSecond =
		accumulators.windSpeed.summary(
			denominator,
		)
	result.WindDirectionDegrees =
		accumulators.windDirection.summary(
			denominator,
		)
	result.WindGustsMetersPerSecond =
		accumulators.windGusts.summary(
			denominator,
		)

	conditionTotal := 0
	for _, count := range accumulators.conditions {
		conditionTotal += count
	}

	result.Conditions = make(
		[]ConditionFrequency,
		0,
		len(accumulators.conditions),
	)
	for key, count := range accumulators.conditions {
		result.Conditions = append(
			result.Conditions,
			ConditionFrequency{
				Scheme: key.scheme,
				Code:   key.code,
				Count:  count,
				Share: float64(count) /
					float64(conditionTotal),
			},
		)
	}
	sort.Slice(
		result.Conditions,
		func(left int, right int) bool {
			if result.Conditions[left].Scheme ==
				result.Conditions[right].Scheme {
				return result.Conditions[left].Code <
					result.Conditions[right].Code
			}
			return result.Conditions[left].Scheme <
				result.Conditions[right].Scheme
		},
	)

	for index := range result.Conditions {
		condition := result.Conditions[index]
		if result.DominantCondition == nil ||
			condition.Count >
				result.DominantCondition.Count ||
			(condition.Count ==
				result.DominantCondition.Count &&
				conditionLess(
					condition,
					*result.DominantCondition,
				)) {
			dominant := condition
			result.DominantCondition =
				&dominant
		}
	}
}

func conditionLess(
	left ConditionFrequency,
	right ConditionFrequency,
) bool {
	if left.Scheme == right.Scheme {
		return left.Code < right.Code
	}
	return left.Scheme < right.Scheme
}

func addCompletenessLimitations(
	result *Result,
	policy Policy,
) {
	if result.AlignmentStatus !=
		weatheralignment.StatusComplete {
		result.Limitations = append(
			result.Limitations,
			Notice{
				Code:    "weather_alignment_not_complete",
				Message: "The upstream four-dimensional alignment is not complete.",
			},
		)
	}
	if result.ProfileCoverageRatio <
		policy.MinimumCompleteProfileCoverage {
		result.Limitations = append(
			result.Limitations,
			Notice{
				Code:    "weather_encounter_coverage_below_complete_threshold",
				Message: "Weather encounter coverage is below the production threshold for a complete profile.",
			},
		)
	}

	coreMetrics := []struct {
		code     string
		message  string
		coverage float64
	}{
		{
			code:    "temperature_coverage_below_complete_threshold",
			message: "Temperature coverage is below the production threshold for a complete profile.",
			coverage: result.
				TemperatureCelsius.
				CoverageRatio,
		},
		{
			code:    "wind_speed_coverage_below_complete_threshold",
			message: "Wind-speed coverage is below the production threshold for a complete profile.",
			coverage: result.
				WindSpeedMetersPerSecond.
				CoverageRatio,
		},
		{
			code:    "wind_direction_coverage_below_complete_threshold",
			message: "Wind-direction coverage is below the production threshold for a complete profile.",
			coverage: result.
				WindDirectionDegrees.
				CoverageRatio,
		},
	}
	for _, metric := range coreMetrics {
		if metric.coverage <
			policy.
				MinimumCompleteCoreMetricCoverage {
			result.Limitations = append(
				result.Limitations,
				Notice{
					Code:    metric.code,
					Message: metric.message,
				},
			)
		}
	}
}

func validateAndClone(
	result Result,
) (Result, error) {
	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrResultInvalid,
			err,
		)
	}
	return result.Clone(), nil
}
