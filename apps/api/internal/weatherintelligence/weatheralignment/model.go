package weatheralignment

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/weatherintelligence/weathertrust"
)

const (
	Version            = "weather-trajectory-alignment-v1"
	FingerprintVersion = "weather-trajectory-alignment-fingerprint-v1"
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

type AltitudeBasis string

const (
	AltitudeBasisGround      AltitudeBasis = "ground"
	AltitudeBasisGeometric   AltitudeBasis = "geometric"
	AltitudeBasisBarometric  AltitudeBasis = "barometric"
	AltitudeBasisUnavailable AltitudeBasis = "unavailable"
)

func (basis AltitudeBasis) IsKnown() bool {
	switch basis {
	case AltitudeBasisGround,
		AltitudeBasisGeometric,
		AltitudeBasisBarometric,
		AltitudeBasisUnavailable:
		return true
	default:
		return false
	}
}

type MatchStatus string

const (
	MatchStatusAligned   MatchStatus = "aligned"
	MatchStatusUnmatched MatchStatus = "unmatched"
)

func (status MatchStatus) IsKnown() bool {
	switch status {
	case MatchStatusAligned, MatchStatusUnmatched:
		return true
	default:
		return false
	}
}

type ComponentName string

const (
	ComponentHorizontal ComponentName = "horizontal"
	ComponentTemporal   ComponentName = "temporal"
	ComponentVertical   ComponentName = "vertical"
)

type Component struct {
	Name   ComponentName
	Score  float64
	Weight float64
}

type Notice struct {
	Code    string
	Message string
}

type Match struct {
	TrajectoryPointSequence int
	TrajectoryPointID       string
	TrajectoryObservedAt    time.Time

	WeatherSampleSequence *int
	WeatherValidAt        *time.Time

	Status MatchStatus

	AltitudeBasis  AltitudeBasis
	AltitudeMeters *float64

	HorizontalDistanceKilometers *float64
	TemporalDistance             *time.Duration
	VerticalDistanceMeters       *float64

	Score      float64
	Components []Component

	Limitations []Notice
}

type Result struct {
	Version string
	Status  Status

	TrajectoryID string
	AsOfTime     time.Time

	TrustDecision weathertrust.Decision
	TrustScore    float64

	PointCount     int
	AlignedCount   int
	UnmatchedCount int
	CoverageRatio  float64

	Matches []Match

	Limitations  []Notice
	Explanations []Notice

	InputFingerprint string
	GeneratedAt      time.Time
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Matches = make([]Match, 0, len(result.Matches))
	for _, match := range result.Matches {
		cloned.Matches = append(cloned.Matches, cloneMatch(match))
	}
	cloned.Limitations = append([]Notice(nil), result.Limitations...)
	cloned.Explanations = append([]Notice(nil), result.Explanations...)
	return cloned
}

func cloneMatch(match Match) Match {
	cloned := match
	cloned.WeatherSampleSequence = cloneInt(match.WeatherSampleSequence)
	cloned.WeatherValidAt = cloneTime(match.WeatherValidAt)
	cloned.AltitudeMeters = cloneFloat64(match.AltitudeMeters)
	cloned.HorizontalDistanceKilometers = cloneFloat64(match.HorizontalDistanceKilometers)
	cloned.TemporalDistance = cloneDuration(match.TemporalDistance)
	cloned.VerticalDistanceMeters = cloneFloat64(match.VerticalDistanceMeters)
	cloned.Components = append([]Component(nil), match.Components...)
	cloned.Limitations = append([]Notice(nil), match.Limitations...)
	return cloned
}

var fingerprintPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func (result Result) Validate() error {
	if result.Version != Version || !result.Status.IsKnown() {
		return fmt.Errorf("weather alignment version or status is invalid")
	}
	if strings.TrimSpace(result.TrajectoryID) == "" ||
		result.AsOfTime.IsZero() ||
		result.GeneratedAt.IsZero() ||
		result.GeneratedAt.Before(result.AsOfTime) {
		return fmt.Errorf("weather alignment identity or time boundary is invalid")
	}
	if !result.TrustDecision.IsKnown() || !unitInterval(result.TrustScore) {
		return fmt.Errorf("weather alignment trust evidence is invalid")
	}
	if result.PointCount < 0 ||
		result.AlignedCount < 0 ||
		result.UnmatchedCount < 0 ||
		result.AlignedCount+result.UnmatchedCount != result.PointCount ||
		len(result.Matches) != result.PointCount {
		return fmt.Errorf("weather alignment counts are inconsistent")
	}
	if !unitInterval(result.CoverageRatio) {
		return fmt.Errorf("weather alignment coverage ratio is invalid")
	}
	expectedCoverage := 0.0
	if result.PointCount > 0 {
		expectedCoverage = float64(result.AlignedCount) / float64(result.PointCount)
	}
	if absolute(result.CoverageRatio-expectedCoverage) > 1e-9 {
		return fmt.Errorf("weather alignment coverage ratio does not match counts")
	}

	for index, match := range result.Matches {
		if match.TrajectoryPointSequence != index ||
			!match.Status.IsKnown() ||
			!match.AltitudeBasis.IsKnown() ||
			match.TrajectoryObservedAt.IsZero() ||
			!unitInterval(match.Score) {
			return fmt.Errorf("weather alignment match identity is invalid")
		}
		if err := validateComponents(match.Components); err != nil {
			return err
		}
		for _, limitation := range match.Limitations {
			if !validNotice(limitation) {
				return fmt.Errorf("weather alignment match limitation is invalid")
			}
		}

		switch match.Status {
		case MatchStatusAligned:
			if match.WeatherSampleSequence == nil ||
				match.WeatherValidAt == nil ||
				match.HorizontalDistanceKilometers == nil ||
				match.TemporalDistance == nil ||
				match.VerticalDistanceMeters == nil ||
				match.AltitudeMeters == nil {
				return fmt.Errorf("aligned weather match is incomplete")
			}
			if *match.WeatherSampleSequence < 0 ||
				match.WeatherValidAt.IsZero() ||
				!finite(*match.HorizontalDistanceKilometers) ||
				*match.HorizontalDistanceKilometers < 0 ||
				*match.TemporalDistance < 0 ||
				!finite(*match.VerticalDistanceMeters) ||
				*match.VerticalDistanceMeters < 0 ||
				!finite(*match.AltitudeMeters) {
				return fmt.Errorf("aligned weather distances are invalid")
			}
		case MatchStatusUnmatched:
			if match.WeatherSampleSequence != nil ||
				match.WeatherValidAt != nil ||
				match.HorizontalDistanceKilometers != nil ||
				match.TemporalDistance != nil ||
				match.VerticalDistanceMeters != nil ||
				len(match.Limitations) == 0 {
				return fmt.Errorf("unmatched weather point must omit match evidence and explain limitations")
			}
		}
	}

	for _, collection := range [][]Notice{result.Limitations, result.Explanations} {
		for _, notice := range collection {
			if !validNotice(notice) {
				return fmt.Errorf("weather alignment notice is invalid")
			}
		}
	}
	if len(result.Explanations) == 0 {
		return fmt.Errorf("weather alignment result requires explanations")
	}
	if !fingerprintPattern.MatchString(result.InputFingerprint) {
		return fmt.Errorf("weather alignment input fingerprint is invalid")
	}

	switch result.Status {
	case StatusUnavailable:
		if result.AlignedCount != 0 ||
			(result.PointCount > 0 && result.UnmatchedCount != result.PointCount) ||
			len(result.Limitations) == 0 {
			return fmt.Errorf("unavailable weather alignment must contain no aligned points and explain limitations")
		}
	case StatusLimited:
		if result.AlignedCount == 0 || len(result.Limitations) == 0 {
			return fmt.Errorf("limited weather alignment must contain aligned points and limitations")
		}
	case StatusComplete:
		if result.PointCount == 0 || result.AlignedCount != result.PointCount {
			return fmt.Errorf("complete weather alignment requires full non-empty coverage")
		}
	}
	return nil
}

func validateComponents(components []Component) error {
	if len(components) != 3 {
		return fmt.Errorf("weather alignment match requires three components")
	}
	expected := map[ComponentName]struct{}{
		ComponentHorizontal: {},
		ComponentTemporal:   {},
		ComponentVertical:   {},
	}
	seen := make(map[ComponentName]struct{})
	weightTotal := 0.0
	for _, component := range components {
		if _, exists := expected[component.Name]; !exists {
			return fmt.Errorf("weather alignment component name is invalid")
		}
		if _, exists := seen[component.Name]; exists {
			return fmt.Errorf("weather alignment component is duplicated")
		}
		seen[component.Name] = struct{}{}
		if !unitInterval(component.Score) || !finite(component.Weight) || component.Weight < 0 {
			return fmt.Errorf("weather alignment component value is invalid")
		}
		weightTotal += component.Weight
	}
	if absolute(weightTotal-1) > 1e-9 {
		return fmt.Errorf("weather alignment component weights must sum to one")
	}
	return nil
}

func normalizeNotices(notices []Notice) []Notice {
	unique := make(map[string]Notice)
	for _, notice := range notices {
		code := strings.TrimSpace(notice.Code)
		message := strings.TrimSpace(notice.Message)
		if code == "" || message == "" {
			continue
		}
		key := code + "\x00" + message
		unique[key] = Notice{Code: code, Message: message}
	}
	keys := make([]string, 0, len(unique))
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]Notice, 0, len(keys))
	for _, key := range keys {
		result = append(result, unique[key])
	}
	return result
}

func validNotice(notice Notice) bool {
	return strings.TrimSpace(notice.Code) != "" && strings.TrimSpace(notice.Message) != ""
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
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

func cloneDuration(value *time.Duration) *time.Duration {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
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
