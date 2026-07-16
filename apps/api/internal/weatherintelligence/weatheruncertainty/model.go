package weatheruncertainty

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

const (
	Version            = "weather-adjusted-uncertainty-v1"
	FingerprintVersion = "weather-adjusted-uncertainty-fingerprint-v1"
)

type Status string

const (
	StatusUnavailable    Status = "unavailable"
	StatusWithheld       Status = "withheld"
	StatusAppliedLimited Status = "applied_limited"
	StatusApplied        Status = "applied"
)

func (status Status) IsKnown() bool {
	switch status {
	case StatusUnavailable, StatusWithheld, StatusAppliedLimited, StatusApplied:
		return true
	default:
		return false
	}
}

type ComponentName string

const (
	ComponentWindSpeed       ComponentName = "wind_speed"
	ComponentWindGust        ComponentName = "wind_gust"
	ComponentPrecipitation   ComponentName = "precipitation"
	ComponentCloudCover      ComponentName = "cloud_cover"
	ComponentEvidenceQuality ComponentName = "evidence_quality"
)

type Component struct {
	Name   ComponentName
	Score  float64
	Weight float64
}

type PointAdjustment struct {
	Sequence     int
	ForecastTime time.Time

	HorizonProgress float64
	Multiplier      float64

	OriginalHorizontalRadiusM float64
	AdjustedHorizontalRadiusM float64

	OriginalVerticalRadiusM *float64
	AdjustedVerticalRadiusM *float64

	OriginalConfidenceScore float64
	AdjustedConfidenceScore float64
}

type ArrivalAdjustment struct {
	Multiplier float64

	OriginalEarliestTime  time.Time
	OriginalEstimatedTime time.Time
	OriginalLatestTime    time.Time

	AdjustedEarliestTime  time.Time
	AdjustedEstimatedTime time.Time
	AdjustedLatestTime    time.Time

	OriginalConfidenceScore float64
	AdjustedConfidenceScore float64
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

	SeverityScore     float64
	WeatherMultiplier float64
	Components        []Component

	PointAdjustments  []PointAdjustment
	ArrivalAdjustment *ArrivalAdjustment

	AdjustedProjection projectioncontract.Result

	Limitations  []Notice
	Explanations []Notice

	InputFingerprint string
	GeneratedAt      time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Components = append([]Component(nil), result.Components...)

	cloned.PointAdjustments = make([]PointAdjustment, 0, len(result.PointAdjustments))
	for _, adjustment := range result.PointAdjustments {
		cloned.PointAdjustments = append(cloned.PointAdjustments, clonePointAdjustment(adjustment))
	}

	if result.ArrivalAdjustment != nil {
		arrival := *result.ArrivalAdjustment
		cloned.ArrivalAdjustment = &arrival
	}

	cloned.AdjustedProjection = result.AdjustedProjection.Clone()
	cloned.Limitations = append([]Notice(nil), result.Limitations...)
	cloned.Explanations = append([]Notice(nil), result.Explanations...)
	return cloned
}

func clonePointAdjustment(adjustment PointAdjustment) PointAdjustment {
	cloned := adjustment
	cloned.OriginalVerticalRadiusM = cloneFloat64(adjustment.OriginalVerticalRadiusM)
	cloned.AdjustedVerticalRadiusM = cloneFloat64(adjustment.AdjustedVerticalRadiusM)
	return cloned
}

var fingerprintPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func (result Result) Validate() error {
	if result.Version != Version || !result.Status.IsKnown() {
		return fmt.Errorf("weather uncertainty version or status is invalid")
	}
	if strings.TrimSpace(result.TrajectoryID) == "" ||
		result.AsOfTime.IsZero() ||
		result.GeneratedAt.IsZero() ||
		result.GeneratedAt.Before(result.AsOfTime) {
		return fmt.Errorf("weather uncertainty identity or time boundary is invalid")
	}
	if !unitInterval(result.SeverityScore) ||
		!finite(result.WeatherMultiplier) ||
		result.WeatherMultiplier < 1 {
		return fmt.Errorf("weather uncertainty severity or multiplier is invalid")
	}
	if err := validateComponents(result.Components); err != nil {
		return err
	}
	if !fingerprintPattern.MatchString(result.InputFingerprint) {
		return fmt.Errorf("weather uncertainty input fingerprint is invalid")
	}
	if len(result.Explanations) == 0 {
		return fmt.Errorf("weather uncertainty explanations are required")
	}
	for _, collection := range [][]Notice{result.Limitations, result.Explanations} {
		for _, notice := range collection {
			if strings.TrimSpace(notice.Code) == "" || strings.TrimSpace(notice.Message) == "" {
				return fmt.Errorf("weather uncertainty notice is invalid")
			}
		}
	}

	projectionReport := projectioncontract.Validate(result.AdjustedProjection)
	if projectionReport.Status != projectioncontract.ValidationStatusValid {
		return fmt.Errorf("adjusted projection contract is invalid: %v", projectionReport.Issues)
	}
	if strings.TrimSpace(result.AdjustedProjection.TrajectoryID) != strings.TrimSpace(result.TrajectoryID) ||
		!result.AdjustedProjection.Horizon.AsOfTime.Equal(result.AsOfTime) {
		return fmt.Errorf("adjusted projection does not match weather uncertainty identity")
	}

	switch result.Status {
	case StatusUnavailable:
		if result.AdjustedProjection.Status != projectioncontract.ResultStatusUnavailable ||
			result.WeatherMultiplier != 1 ||
			result.SeverityScore != 0 ||
			len(result.PointAdjustments) != 0 ||
			result.ArrivalAdjustment != nil ||
			len(result.Limitations) == 0 {
			return fmt.Errorf("unavailable weather uncertainty result is inconsistent")
		}
	case StatusWithheld:
		if result.WeatherMultiplier != 1 ||
			result.SeverityScore != 0 ||
			len(result.PointAdjustments) != 0 ||
			result.ArrivalAdjustment != nil ||
			len(result.Limitations) == 0 {
			return fmt.Errorf("withheld weather uncertainty result must preserve projection values and explain why")
		}
	case StatusAppliedLimited, StatusApplied:
		if result.AdjustedProjection.Status == projectioncontract.ResultStatusUnavailable ||
			len(result.PointAdjustments) != len(result.AdjustedProjection.Points) ||
			result.AdjustedProjection.Provenance.InputFingerprint != result.InputFingerprint ||
			!result.AdjustedProjection.GeneratedAt.Equal(result.GeneratedAt) {
			return fmt.Errorf("applied weather uncertainty result is incomplete")
		}
		if err := validatePointAdjustments(result.PointAdjustments, result.AdjustedProjection.Points); err != nil {
			return err
		}
		if err := validateArrivalAdjustment(result.ArrivalAdjustment, result.AdjustedProjection.Arrival); err != nil {
			return err
		}
		if result.Status == StatusAppliedLimited && len(result.Limitations) == 0 {
			return fmt.Errorf("limited weather uncertainty result requires limitations")
		}
	}

	return nil
}

func validateComponents(components []Component) error {
	if len(components) != 5 {
		return fmt.Errorf("weather uncertainty requires five score components")
	}

	expected := map[ComponentName]struct{}{
		ComponentWindSpeed:       {},
		ComponentWindGust:        {},
		ComponentPrecipitation:   {},
		ComponentCloudCover:      {},
		ComponentEvidenceQuality: {},
	}
	seen := make(map[ComponentName]struct{})
	weightTotal := 0.0

	for _, component := range components {
		if _, exists := expected[component.Name]; !exists {
			return fmt.Errorf("weather uncertainty component name is invalid")
		}
		if _, exists := seen[component.Name]; exists {
			return fmt.Errorf("weather uncertainty component is duplicated")
		}
		seen[component.Name] = struct{}{}
		if !unitInterval(component.Score) || !finite(component.Weight) || component.Weight < 0 {
			return fmt.Errorf("weather uncertainty component value is invalid")
		}
		weightTotal += component.Weight
	}

	if absolute(weightTotal-1) > 1e-9 {
		return fmt.Errorf("weather uncertainty component weights must sum to one")
	}
	return nil
}

func validatePointAdjustments(
	adjustments []PointAdjustment,
	points []projectioncontract.ProjectionPoint,
) error {
	for index, adjustment := range adjustments {
		point := points[index]

		if adjustment.Sequence != index ||
			point.Sequence != index ||
			!adjustment.ForecastTime.Equal(point.ForecastTime) ||
			!unitInterval(adjustment.HorizonProgress) ||
			!finite(adjustment.Multiplier) ||
			adjustment.Multiplier < 1 ||
			!positiveFinite(adjustment.OriginalHorizontalRadiusM) ||
			!positiveFinite(adjustment.AdjustedHorizontalRadiusM) ||
			adjustment.AdjustedHorizontalRadiusM < adjustment.OriginalHorizontalRadiusM ||
			absolute(adjustment.AdjustedHorizontalRadiusM-point.Uncertainty.HorizontalRadiusM) > 1e-6 ||
			!unitInterval(adjustment.OriginalConfidenceScore) ||
			!unitInterval(adjustment.AdjustedConfidenceScore) ||
			adjustment.AdjustedConfidenceScore > adjustment.OriginalConfidenceScore ||
			absolute(adjustment.AdjustedConfidenceScore-point.Confidence.Score) > 1e-9 {
			return fmt.Errorf("weather uncertainty point adjustment is invalid")
		}

		if (adjustment.OriginalVerticalRadiusM == nil) != (adjustment.AdjustedVerticalRadiusM == nil) ||
			(adjustment.AdjustedVerticalRadiusM == nil) != (point.Uncertainty.VerticalRadiusM == nil) {
			return fmt.Errorf("weather uncertainty vertical adjustment dimensions are inconsistent")
		}

		if adjustment.OriginalVerticalRadiusM != nil {
			if !positiveFinite(*adjustment.OriginalVerticalRadiusM) ||
				!positiveFinite(*adjustment.AdjustedVerticalRadiusM) ||
				*adjustment.AdjustedVerticalRadiusM < *adjustment.OriginalVerticalRadiusM ||
				absolute(*adjustment.AdjustedVerticalRadiusM-*point.Uncertainty.VerticalRadiusM) > 1e-6 {
				return fmt.Errorf("weather uncertainty vertical radius adjustment is invalid")
			}
		}
	}
	return nil
}

func validateArrivalAdjustment(
	adjustment *ArrivalAdjustment,
	arrival *projectioncontract.ArrivalEstimate,
) error {
	if adjustment == nil {
		if arrival == nil {
			return nil
		}
		return fmt.Errorf("weather uncertainty arrival adjustment is missing")
	}
	if arrival == nil {
		return fmt.Errorf("weather uncertainty arrival adjustment has no projection arrival")
	}
	if !finite(adjustment.Multiplier) ||
		adjustment.Multiplier < 1 ||
		adjustment.OriginalEarliestTime.IsZero() ||
		adjustment.OriginalEstimatedTime.IsZero() ||
		adjustment.OriginalLatestTime.IsZero() ||
		adjustment.AdjustedEarliestTime.IsZero() ||
		adjustment.AdjustedEstimatedTime.IsZero() ||
		adjustment.AdjustedLatestTime.IsZero() ||
		adjustment.OriginalEarliestTime.After(adjustment.OriginalEstimatedTime) ||
		adjustment.OriginalEstimatedTime.After(adjustment.OriginalLatestTime) ||
		adjustment.AdjustedEarliestTime.After(adjustment.AdjustedEstimatedTime) ||
		adjustment.AdjustedEstimatedTime.After(adjustment.AdjustedLatestTime) ||
		!adjustment.AdjustedEarliestTime.Equal(arrival.EarliestTime) ||
		!adjustment.AdjustedEstimatedTime.Equal(arrival.EstimatedTime) ||
		!adjustment.AdjustedLatestTime.Equal(arrival.LatestTime) ||
		!unitInterval(adjustment.OriginalConfidenceScore) ||
		!unitInterval(adjustment.AdjustedConfidenceScore) ||
		adjustment.AdjustedConfidenceScore > adjustment.OriginalConfidenceScore ||
		absolute(adjustment.AdjustedConfidenceScore-arrival.Confidence.Score) > 1e-9 {
		return fmt.Errorf("weather uncertainty arrival adjustment is invalid")
	}
	return nil
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func positiveFinite(value float64) bool {
	return finite(value) && value > 0
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func unitInterval(value float64) bool {
	return finite(value) && value >= 0 && value <= 1
}

func clampUnit(value float64) float64 {
	switch {
	case !finite(value):
		return 0
	case value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}

func absolute(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
