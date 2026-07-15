package projectionfreshness

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	Version            = "projection-pattern-freshness-guard-v1"
	FingerprintVersion = "projection-pattern-freshness-fingerprint-v1"
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
	ComponentNewestAge     ComponentName = "newest_neighbor_age"
	ComponentMeanAge       ComponentName = "mean_neighbor_age"
	ComponentOldestAge     ComponentName = "oldest_neighbor_age"
	ComponentRecentSupport ComponentName = "recent_neighbor_support"
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

type Result struct {
	Version  string
	Decision Decision
	Usable   bool

	AsOfTime time.Time

	NeighborCount       int
	RecentNeighborCount int

	NewestNeighborAge time.Duration
	MeanNeighborAge   time.Duration
	OldestNeighborAge time.Duration

	Score      float64
	Components []Component

	SelectedTrajectoryIDs []string
	Limitations           []Notice
	InputFingerprint      string
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Components = append(
		[]Component(nil),
		result.Components...,
	)
	cloned.SelectedTrajectoryIDs = append(
		[]string(nil),
		result.SelectedTrajectoryIDs...,
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
		!result.Decision.IsKnown() {
		return fmt.Errorf(
			"pattern freshness version or decision is invalid",
		)
	}
	if result.AsOfTime.IsZero() ||
		result.NeighborCount < 0 ||
		result.RecentNeighborCount < 0 ||
		result.RecentNeighborCount >
			result.NeighborCount ||
		result.NeighborCount !=
			len(result.SelectedTrajectoryIDs) {
		return fmt.Errorf(
			"pattern freshness counts or as-of time are invalid",
		)
	}
	if result.NewestNeighborAge < 0 ||
		result.MeanNeighborAge < 0 ||
		result.OldestNeighborAge < 0 ||
		(result.NeighborCount > 0 &&
			(result.NewestNeighborAge >
				result.MeanNeighborAge ||
				result.MeanNeighborAge >
					result.OldestNeighborAge)) {
		return fmt.Errorf(
			"pattern freshness age measurements are invalid",
		)
	}
	if !unitInterval(result.Score) ||
		len(result.Components) != 4 {
		return fmt.Errorf(
			"pattern freshness score or components are invalid",
		)
	}

	seenComponents := make(
		map[ComponentName]struct{},
		len(result.Components),
	)
	weightTotal := 0.0
	for _, component := range result.Components {
		switch component.Name {
		case ComponentNewestAge,
			ComponentMeanAge,
			ComponentOldestAge,
			ComponentRecentSupport:
		default:
			return fmt.Errorf(
				"pattern freshness component name is invalid",
			)
		}
		if _, exists :=
			seenComponents[component.Name]; exists {
			return fmt.Errorf(
				"duplicate pattern freshness component",
			)
		}
		seenComponents[component.Name] =
			struct{}{}
		if !unitInterval(component.Score) ||
			!finite(component.Weight) ||
			component.Weight < 0 {
			return fmt.Errorf(
				"pattern freshness component is invalid",
			)
		}
		weightTotal += component.Weight
	}
	if absolute(weightTotal-1) > 1e-9 {
		return fmt.Errorf(
			"pattern freshness component weights do not sum to one",
		)
	}

	seenIDs := make(map[string]struct{})
	for _, trajectoryID := range result.SelectedTrajectoryIDs {
		normalized := strings.TrimSpace(
			trajectoryID,
		)
		if normalized == "" {
			return fmt.Errorf(
				"selected trajectory identifier is required",
			)
		}
		if _, exists := seenIDs[normalized]; exists {
			return fmt.Errorf(
				"duplicate selected trajectory identifier",
			)
		}
		seenIDs[normalized] =
			struct{}{}
	}
	if !sort.StringsAreSorted(
		result.SelectedTrajectoryIDs,
	) {
		return fmt.Errorf(
			"selected trajectory identifiers must be sorted",
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
				"pattern freshness limitation is invalid",
			)
		}
	}
	if !fingerprintPattern.MatchString(
		result.InputFingerprint,
	) {
		return fmt.Errorf(
			"pattern freshness input fingerprint is invalid",
		)
	}

	switch result.Decision {
	case DecisionBlocked:
		if result.Usable ||
			len(result.Limitations) == 0 {
			return fmt.Errorf(
				"blocked freshness result must be unusable and explain limitations",
			)
		}
	case DecisionLimited:
		if !result.Usable ||
			len(result.Limitations) == 0 {
			return fmt.Errorf(
				"limited freshness result must remain usable and explain limitations",
			)
		}
	case DecisionAllowed:
		if !result.Usable {
			return fmt.Errorf(
				"allowed freshness result must be usable",
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
