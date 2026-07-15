package projectionroutefrequency

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	Version            = "projection-low-frequency-route-guard-v1"
	FingerprintVersion = "projection-low-frequency-route-fingerprint-v1"
)

type Decision string

const (
	DecisionBlocked Decision = "blocked"
	DecisionLimited Decision = "limited"
	DecisionAllowed Decision = "allowed"
)

func (decision Decision) IsKnown() bool {
	switch decision {
	case DecisionBlocked,
		DecisionLimited,
		DecisionAllowed:
		return true
	default:
		return false
	}
}

type ComponentName string

const (
	ComponentObservationCount   ComponentName = "observation_count"
	ComponentDistinctDays       ComponentName = "distinct_days"
	ComponentRecentObservations ComponentName = "recent_observations"
	ComponentLatestObservation  ComponentName = "latest_observation"
	ComponentRouteConfidence    ComponentName = "route_confidence"
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

type HistorySummary struct {
	RouteKey string

	WindowStart time.Time
	WindowEnd   time.Time
	AsOfTime    time.Time

	ObservationCount       int
	DistinctFlightCount    int
	DistinctDayCount       int
	RecentObservationCount int
	LastObservedAt         time.Time

	SourceNames      []string
	InputFingerprint string
}

func (summary HistorySummary) Clone() HistorySummary {
	cloned := summary
	cloned.SourceNames = append(
		[]string(nil),
		summary.SourceNames...,
	)

	return cloned
}

func (summary HistorySummary) Validate() error {
	if strings.TrimSpace(
		summary.RouteKey,
	) == "" ||
		summary.WindowStart.IsZero() ||
		summary.WindowEnd.IsZero() ||
		summary.AsOfTime.IsZero() ||
		!summary.WindowStart.Before(
			summary.WindowEnd,
		) ||
		summary.WindowEnd.After(
			summary.AsOfTime,
		) {
		return fmt.Errorf(
			"route history identity or window is invalid",
		)
	}
	if summary.ObservationCount < 0 ||
		summary.DistinctFlightCount < 0 ||
		summary.DistinctDayCount < 0 ||
		summary.RecentObservationCount < 0 ||
		summary.DistinctFlightCount >
			summary.ObservationCount ||
		summary.RecentObservationCount >
			summary.ObservationCount {
		return fmt.Errorf(
			"route history counts are invalid",
		)
	}
	if summary.ObservationCount == 0 {
		if !summary.LastObservedAt.IsZero() {
			return fmt.Errorf(
				"empty route history must not publish a latest observation",
			)
		}
	} else if summary.LastObservedAt.IsZero() ||
		summary.LastObservedAt.Before(
			summary.WindowStart,
		) ||
		summary.LastObservedAt.After(
			summary.WindowEnd,
		) {
		return fmt.Errorf(
			"route history latest observation is invalid",
		)
	}

	seenSources := make(map[string]struct{})
	for _, sourceName := range summary.SourceNames {
		normalized := strings.TrimSpace(
			sourceName,
		)
		if normalized == "" {
			return fmt.Errorf(
				"route history source name is required",
			)
		}
		if _, exists := seenSources[normalized]; exists {
			return fmt.Errorf(
				"duplicate route history source name",
			)
		}
		seenSources[normalized] =
			struct{}{}
	}
	if !sort.StringsAreSorted(
		summary.SourceNames,
	) {
		return fmt.Errorf(
			"route history source names must be sorted",
		)
	}
	if !fingerprintPattern.MatchString(
		summary.InputFingerprint,
	) {
		return fmt.Errorf(
			"route history fingerprint is invalid",
		)
	}

	return nil
}

type Result struct {
	Version  string
	Decision Decision
	Usable   bool

	RouteKey string
	AsOfTime time.Time

	ObservationCount       int
	DistinctFlightCount    int
	DistinctDayCount       int
	RecentObservationCount int
	LatestObservationAge   time.Duration
	RouteConfidenceScore   float64

	Score       float64
	Components  []Component
	Limitations []Notice

	HistoryInputFingerprint string
	InputFingerprint        string
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Components = append(
		[]Component(nil),
		result.Components...,
	)
	cloned.Limitations = append(
		[]Notice(nil),
		result.Limitations...,
	)

	return cloned
}

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func (result Result) Validate() error {
	if result.Version != Version ||
		!result.Decision.IsKnown() ||
		strings.TrimSpace(result.RouteKey) == "" ||
		result.AsOfTime.IsZero() {
		return fmt.Errorf(
			"route-frequency result metadata is invalid",
		)
	}
	if result.ObservationCount < 0 ||
		result.DistinctFlightCount < 0 ||
		result.DistinctDayCount < 0 ||
		result.RecentObservationCount < 0 ||
		result.DistinctFlightCount >
			result.ObservationCount ||
		result.RecentObservationCount >
			result.ObservationCount ||
		result.LatestObservationAge < 0 ||
		!unitInterval(
			result.RouteConfidenceScore,
		) ||
		!unitInterval(result.Score) {
		return fmt.Errorf(
			"route-frequency result measurements are invalid",
		)
	}
	if len(result.Components) != 5 {
		return fmt.Errorf(
			"route-frequency result requires five components",
		)
	}

	seenComponents := make(
		map[ComponentName]struct{},
		len(result.Components),
	)
	weightTotal := 0.0
	for _, component := range result.Components {
		switch component.Name {
		case ComponentObservationCount,
			ComponentDistinctDays,
			ComponentRecentObservations,
			ComponentLatestObservation,
			ComponentRouteConfidence:
		default:
			return fmt.Errorf(
				"route-frequency component name is invalid",
			)
		}
		if _, exists :=
			seenComponents[component.Name]; exists {
			return fmt.Errorf(
				"duplicate route-frequency component",
			)
		}
		seenComponents[component.Name] =
			struct{}{}
		if !unitInterval(component.Score) ||
			!finite(component.Weight) ||
			component.Weight < 0 {
			return fmt.Errorf(
				"route-frequency component is invalid",
			)
		}
		weightTotal += component.Weight
	}
	if absolute(weightTotal-1) > 1e-9 {
		return fmt.Errorf(
			"route-frequency component weights do not sum to one",
		)
	}

	for _, limitation := range result.Limitations {
		if strings.TrimSpace(
			limitation.Code,
		) == "" ||
			strings.TrimSpace(
				limitation.Message,
			) == "" {
			return fmt.Errorf(
				"route-frequency limitation is invalid",
			)
		}
	}
	if !fingerprintPattern.MatchString(
		result.HistoryInputFingerprint,
	) ||
		!fingerprintPattern.MatchString(
			result.InputFingerprint,
		) {
		return fmt.Errorf(
			"route-frequency fingerprints are invalid",
		)
	}

	switch result.Decision {
	case DecisionBlocked:
		if result.Usable ||
			len(result.Limitations) == 0 {
			return fmt.Errorf(
				"blocked route-frequency result must be unusable and explain limitations",
			)
		}
	case DecisionLimited:
		if !result.Usable ||
			len(result.Limitations) == 0 {
			return fmt.Errorf(
				"limited route-frequency result must remain usable and explain limitations",
			)
		}
	case DecisionAllowed:
		if !result.Usable {
			return fmt.Errorf(
				"allowed route-frequency result must be usable",
			)
		}
	}

	return nil
}

func normalizeNotices(
	items []Notice,
) []Notice {
	seen := make(
		map[string]Notice,
		len(items),
	)
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		message := strings.TrimSpace(
			item.Message,
		)
		if code == "" || message == "" {
			continue
		}
		key := code + "\x00" + message
		seen[key] = Notice{
			Code:    code,
			Message: message,
		}
	}

	keys := make(
		[]string,
		0,
		len(seen),
	)
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make(
		[]Notice,
		0,
		len(keys),
	)
	for _, key := range keys {
		result = append(
			result,
			seen[key],
		)
	}

	return result
}

func absolute(value float64) float64 {
	if value < 0 {
		return -value
	}

	return value
}
