package weatherencounter

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weatheralignment"
)

const (
	Version            = "weather-encounter-profile-v1"
	FingerprintVersion = "weather-encounter-profile-fingerprint-v1"
)

type Status string

const (
	StatusUnavailable Status = "unavailable"
	StatusLimited     Status = "limited"
	StatusComplete    Status = "complete"
)

func (status Status) IsKnown() bool {
	switch status {
	case StatusUnavailable, StatusLimited, StatusComplete:
		return true
	default:
		return false
	}
}

type MetricSummary struct {
	PresentCount  int
	CoverageRatio float64
	Minimum       *float64
	Maximum       *float64
	Mean          *float64
}

type CircularDirectionSummary struct {
	PresentCount         int
	CoverageRatio        float64
	MeanDirectionDegrees *float64
	Concentration        *float64
}

type ConditionFrequency struct {
	Scheme string
	Code   int
	Count  int
	Share  float64
}

type EncounterPoint struct {
	TrajectoryPointSequence int
	TrajectoryPointID       string
	TrajectoryObservedAt    time.Time

	WeatherSampleSequence int
	WeatherValidAt        time.Time

	AlignmentScore float64
	FeatureCount   int
}

type Notice struct {
	Code    string
	Message string
}

type Result struct {
	Version string
	Status  Status

	TrajectoryID string
	AsOfTime     time.Time

	AlignmentStatus        weatheralignment.Status
	AlignmentCoverageRatio float64

	PointCount           int
	EncounterPointCount  int
	UnprofiledPointCount int
	ProfileCoverageRatio float64

	EncounterStartedAt *time.Time
	EncounterEndedAt   *time.Time

	TemperatureCelsius       MetricSummary
	RelativeHumidityPercent  MetricSummary
	PrecipitationMillimeters MetricSummary
	RainMillimeters          MetricSummary
	CloudCoverPercent        MetricSummary
	SurfacePressureHPA       MetricSummary
	WindSpeedMetersPerSecond MetricSummary
	WindDirectionDegrees     CircularDirectionSummary
	WindGustsMetersPerSecond MetricSummary

	Conditions        []ConditionFrequency
	DominantCondition *ConditionFrequency
	Points            []EncounterPoint

	Limitations  []Notice
	Explanations []Notice

	InputFingerprint string
	GeneratedAt      time.Time
}

func (result Result) Clone() Result {
	cloned := result

	cloned.EncounterStartedAt = cloneTime(
		result.EncounterStartedAt,
	)
	cloned.EncounterEndedAt = cloneTime(
		result.EncounterEndedAt,
	)

	cloned.TemperatureCelsius = cloneMetricSummary(
		result.TemperatureCelsius,
	)
	cloned.RelativeHumidityPercent = cloneMetricSummary(
		result.RelativeHumidityPercent,
	)
	cloned.PrecipitationMillimeters = cloneMetricSummary(
		result.PrecipitationMillimeters,
	)
	cloned.RainMillimeters = cloneMetricSummary(
		result.RainMillimeters,
	)
	cloned.CloudCoverPercent = cloneMetricSummary(
		result.CloudCoverPercent,
	)
	cloned.SurfacePressureHPA = cloneMetricSummary(
		result.SurfacePressureHPA,
	)
	cloned.WindSpeedMetersPerSecond = cloneMetricSummary(
		result.WindSpeedMetersPerSecond,
	)
	cloned.WindDirectionDegrees =
		cloneCircularDirectionSummary(
			result.WindDirectionDegrees,
		)
	cloned.WindGustsMetersPerSecond = cloneMetricSummary(
		result.WindGustsMetersPerSecond,
	)

	cloned.Conditions = append(
		[]ConditionFrequency(nil),
		result.Conditions...,
	)
	if result.DominantCondition != nil {
		dominant := *result.DominantCondition
		cloned.DominantCondition = &dominant
	}
	cloned.Points = append(
		[]EncounterPoint(nil),
		result.Points...,
	)
	cloned.Limitations = append(
		[]Notice(nil),
		result.Limitations...,
	)
	cloned.Explanations = append(
		[]Notice(nil),
		result.Explanations...,
	)

	return cloned
}

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func (result Result) Validate() error {
	if result.Version != Version ||
		!result.Status.IsKnown() {
		return fmt.Errorf(
			"weather encounter version or status is invalid",
		)
	}
	if strings.TrimSpace(result.TrajectoryID) == "" ||
		result.AsOfTime.IsZero() ||
		result.GeneratedAt.IsZero() ||
		result.GeneratedAt.Before(result.AsOfTime) {
		return fmt.Errorf(
			"weather encounter identity or time boundary is invalid",
		)
	}
	if !result.AlignmentStatus.IsKnown() ||
		!unitInterval(
			result.AlignmentCoverageRatio,
		) {
		return fmt.Errorf(
			"weather encounter alignment evidence is invalid",
		)
	}
	if result.PointCount < 0 ||
		result.EncounterPointCount < 0 ||
		result.UnprofiledPointCount < 0 ||
		result.EncounterPointCount+
			result.UnprofiledPointCount !=
			result.PointCount ||
		len(result.Points) !=
			result.EncounterPointCount {
		return fmt.Errorf(
			"weather encounter counts are inconsistent",
		)
	}
	if !unitInterval(
		result.ProfileCoverageRatio,
	) {
		return fmt.Errorf(
			"weather encounter profile coverage is invalid",
		)
	}
	expectedCoverage := 0.0
	if result.PointCount > 0 {
		expectedCoverage =
			float64(
				result.EncounterPointCount,
			) /
				float64(result.PointCount)
	}
	if absolute(
		result.ProfileCoverageRatio-
			expectedCoverage,
	) > 1e-9 {
		return fmt.Errorf(
			"weather encounter coverage does not match counts",
		)
	}

	if err := validateEncounterTimes(result); err != nil {
		return err
	}

	for index, point := range result.Points {
		if point.TrajectoryPointSequence < 0 ||
			(index > 0 &&
				point.TrajectoryPointSequence <=
					result.Points[index-1].
						TrajectoryPointSequence) ||
			point.TrajectoryObservedAt.IsZero() ||
			point.WeatherSampleSequence < 0 ||
			point.WeatherValidAt.IsZero() ||
			!unitInterval(point.AlignmentScore) ||
			point.FeatureCount <= 0 {
			return fmt.Errorf(
				"weather encounter point is invalid",
			)
		}
	}

	for _, summary := range []MetricSummary{
		result.TemperatureCelsius,
		result.RelativeHumidityPercent,
		result.PrecipitationMillimeters,
		result.RainMillimeters,
		result.CloudCoverPercent,
		result.SurfacePressureHPA,
		result.WindSpeedMetersPerSecond,
		result.WindGustsMetersPerSecond,
	} {
		if err := validateMetricSummary(
			summary,
			result.EncounterPointCount,
		); err != nil {
			return err
		}
	}
	if err := validateCircularDirectionSummary(
		result.WindDirectionDegrees,
		result.EncounterPointCount,
	); err != nil {
		return err
	}
	if err := validateConditions(
		result.Conditions,
		result.DominantCondition,
		result.EncounterPointCount,
	); err != nil {
		return err
	}

	for _, collection := range [][]Notice{
		result.Limitations,
		result.Explanations,
	} {
		for _, notice := range collection {
			if !validNotice(notice) {
				return fmt.Errorf(
					"weather encounter notice is invalid",
				)
			}
		}
	}
	if len(result.Explanations) == 0 {
		return fmt.Errorf(
			"weather encounter explanations are required",
		)
	}
	if !fingerprintPattern.MatchString(
		result.InputFingerprint,
	) {
		return fmt.Errorf(
			"weather encounter input fingerprint is invalid",
		)
	}

	switch result.Status {
	case StatusUnavailable:
		if result.EncounterPointCount != 0 ||
			result.EncounterStartedAt != nil ||
			result.EncounterEndedAt != nil ||
			len(result.Limitations) == 0 {
			return fmt.Errorf(
				"unavailable weather encounter must contain no profile points and explain limitations",
			)
		}
	case StatusLimited:
		if result.EncounterPointCount == 0 ||
			len(result.Limitations) == 0 {
			return fmt.Errorf(
				"limited weather encounter requires evidence and limitations",
			)
		}
	case StatusComplete:
		if result.PointCount == 0 ||
			result.EncounterPointCount !=
				result.PointCount ||
			result.AlignmentStatus !=
				weatheralignment.StatusComplete {
			return fmt.Errorf(
				"complete weather encounter requires complete non-empty alignment coverage",
			)
		}
	}

	return nil
}

func validateEncounterTimes(
	result Result,
) error {
	if result.EncounterPointCount == 0 {
		if result.EncounterStartedAt != nil ||
			result.EncounterEndedAt != nil {
			return fmt.Errorf(
				"weather encounter without points must not publish encounter times",
			)
		}
		return nil
	}
	if result.EncounterStartedAt == nil ||
		result.EncounterEndedAt == nil ||
		result.EncounterStartedAt.After(
			*result.EncounterEndedAt,
		) {
		return fmt.Errorf(
			"weather encounter time range is invalid",
		)
	}

	first := result.Points[0].
		TrajectoryObservedAt.UTC()
	last := result.Points[len(result.Points)-1].
		TrajectoryObservedAt.UTC()
	if !result.EncounterStartedAt.Equal(first) ||
		!result.EncounterEndedAt.Equal(last) {
		return fmt.Errorf(
			"weather encounter time range does not match profile points",
		)
	}

	return nil
}

func validateMetricSummary(
	summary MetricSummary,
	denominator int,
) error {
	if summary.PresentCount < 0 ||
		summary.PresentCount > denominator ||
		!unitInterval(summary.CoverageRatio) {
		return fmt.Errorf(
			"weather encounter metric count or coverage is invalid",
		)
	}

	expectedCoverage := 0.0
	if denominator > 0 {
		expectedCoverage =
			float64(summary.PresentCount) /
				float64(denominator)
	}
	if absolute(
		summary.CoverageRatio-
			expectedCoverage,
	) > 1e-9 {
		return fmt.Errorf(
			"weather encounter metric coverage does not match count",
		)
	}

	if summary.PresentCount == 0 {
		if summary.Minimum != nil ||
			summary.Maximum != nil ||
			summary.Mean != nil {
			return fmt.Errorf(
				"empty weather encounter metric must not publish values",
			)
		}
		return nil
	}

	if summary.Minimum == nil ||
		summary.Maximum == nil ||
		summary.Mean == nil ||
		!finite(*summary.Minimum) ||
		!finite(*summary.Maximum) ||
		!finite(*summary.Mean) ||
		*summary.Minimum > *summary.Mean ||
		*summary.Mean > *summary.Maximum {
		return fmt.Errorf(
			"weather encounter metric values are invalid",
		)
	}

	return nil
}

func validateCircularDirectionSummary(
	summary CircularDirectionSummary,
	denominator int,
) error {
	if summary.PresentCount < 0 ||
		summary.PresentCount > denominator ||
		!unitInterval(summary.CoverageRatio) {
		return fmt.Errorf(
			"weather encounter wind-direction count or coverage is invalid",
		)
	}

	expectedCoverage := 0.0
	if denominator > 0 {
		expectedCoverage =
			float64(summary.PresentCount) /
				float64(denominator)
	}
	if absolute(
		summary.CoverageRatio-
			expectedCoverage,
	) > 1e-9 {
		return fmt.Errorf(
			"weather encounter wind-direction coverage does not match count",
		)
	}

	if summary.PresentCount == 0 {
		if summary.MeanDirectionDegrees != nil ||
			summary.Concentration != nil {
			return fmt.Errorf(
				"empty weather direction summary must not publish values",
			)
		}
		return nil
	}

	if summary.MeanDirectionDegrees == nil ||
		summary.Concentration == nil ||
		!finite(*summary.MeanDirectionDegrees) ||
		*summary.MeanDirectionDegrees < 0 ||
		*summary.MeanDirectionDegrees >= 360 ||
		!unitInterval(*summary.Concentration) {
		return fmt.Errorf(
			"weather encounter wind-direction values are invalid",
		)
	}

	return nil
}

func validateConditions(
	conditions []ConditionFrequency,
	dominant *ConditionFrequency,
	denominator int,
) error {
	if !sort.SliceIsSorted(
		conditions,
		func(left int, right int) bool {
			if conditions[left].Scheme ==
				conditions[right].Scheme {
				return conditions[left].Code <
					conditions[right].Code
			}
			return conditions[left].Scheme <
				conditions[right].Scheme
		},
	) {
		return fmt.Errorf(
			"weather encounter conditions must be sorted",
		)
	}

	total := 0
	maximumCount := 0
	seen := make(map[string]struct{})
	for _, condition := range conditions {
		key := strings.TrimSpace(
			condition.Scheme,
		) + "\x00" + fmt.Sprintf(
			"%d",
			condition.Code,
		)
		if strings.TrimSpace(
			condition.Scheme,
		) == "" ||
			condition.Code < 0 ||
			condition.Count <= 0 ||
			condition.Count > denominator ||
			!unitInterval(condition.Share) {
			return fmt.Errorf(
				"weather encounter condition frequency is invalid",
			)
		}
		if _, exists := seen[key]; exists {
			return fmt.Errorf(
				"weather encounter condition frequency is duplicated",
			)
		}
		seen[key] = struct{}{}
		total += condition.Count
		if condition.Count > maximumCount {
			maximumCount = condition.Count
		}
	}

	for _, condition := range conditions {
		expectedShare := 0.0
		if total > 0 {
			expectedShare =
				float64(condition.Count) /
					float64(total)
		}
		if absolute(
			condition.Share-
				expectedShare,
		) > 1e-9 {
			return fmt.Errorf(
				"weather encounter condition share is invalid",
			)
		}
	}

	if len(conditions) == 0 {
		if dominant != nil {
			return fmt.Errorf(
				"weather encounter without conditions must not publish a dominant condition",
			)
		}
		return nil
	}
	if dominant == nil ||
		dominant.Count != maximumCount {
		return fmt.Errorf(
			"weather encounter dominant condition is invalid",
		)
	}

	found := false
	for _, condition := range conditions {
		if condition == *dominant {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf(
			"weather encounter dominant condition is not present in frequencies",
		)
	}

	return nil
}

func cloneMetricSummary(
	summary MetricSummary,
) MetricSummary {
	cloned := summary
	cloned.Minimum = cloneFloat64(summary.Minimum)
	cloned.Maximum = cloneFloat64(summary.Maximum)
	cloned.Mean = cloneFloat64(summary.Mean)
	return cloned
}

func cloneCircularDirectionSummary(
	summary CircularDirectionSummary,
) CircularDirectionSummary {
	cloned := summary
	cloned.MeanDirectionDegrees = cloneFloat64(
		summary.MeanDirectionDegrees,
	)
	cloned.Concentration = cloneFloat64(
		summary.Concentration,
	)
	return cloned
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}

func normalizeNotices(
	notices []Notice,
) []Notice {
	unique := make(map[string]Notice)
	for _, notice := range notices {
		code := strings.TrimSpace(notice.Code)
		message := strings.TrimSpace(
			notice.Message,
		)
		if code == "" || message == "" {
			continue
		}
		key := code + "\x00" + message
		unique[key] = Notice{
			Code:    code,
			Message: message,
		}
	}

	keys := make([]string, 0, len(unique))
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(
		[]Notice,
		0,
		len(keys),
	)
	for _, key := range keys {
		result = append(result, unique[key])
	}
	return result
}

func validNotice(notice Notice) bool {
	return strings.TrimSpace(notice.Code) != "" &&
		strings.TrimSpace(notice.Message) != ""
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}

func unitInterval(value float64) bool {
	return finite(value) &&
		value >= 0 &&
		value <= 1
}

func absolute(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
