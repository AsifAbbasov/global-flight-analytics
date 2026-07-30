package projectionfreshness

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	Version                  = "projection-pattern-freshness-guard-v2"
	FingerprintVersion       = "projection-pattern-freshness-fingerprint-v2"
	scoreComparisonTolerance = 1e-9
)

type Decision string

const (
	DecisionBlocked Decision = "blocked"
	DecisionLimited Decision = "limited"
	DecisionAllowed Decision = "allowed"
)

func (decision Decision) IsKnown() bool {
	switch decision {
	case DecisionBlocked, DecisionLimited, DecisionAllowed:
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

var canonicalComponentNames = []ComponentName{
	ComponentNewestAge,
	ComponentMeanAge,
	ComponentOldestAge,
	ComponentRecentSupport,
}

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
	cloned.Components = append([]Component(nil), result.Components...)
	cloned.SelectedTrajectoryIDs = append([]string(nil), result.SelectedTrajectoryIDs...)
	cloned.Limitations = append([]Notice(nil), result.Limitations...)
	return cloned
}

var fingerprintPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func (result Result) Validate() error {
	if result.Version != Version || !result.Decision.IsKnown() {
		return fmt.Errorf("pattern freshness version or decision is invalid")
	}
	if err := validateCountsAndAges(result); err != nil {
		return err
	}
	if err := validateResultComponents(result); err != nil {
		return err
	}
	if err := validateSelectedTrajectoryIDs(result.SelectedTrajectoryIDs); err != nil {
		return err
	}
	if err := validateLimitations(result.Limitations); err != nil {
		return err
	}
	if !fingerprintPattern.MatchString(result.InputFingerprint) {
		return fmt.Errorf("pattern freshness input fingerprint is invalid")
	}
	return validateDecisionShape(result)
}

func validateCountsAndAges(result Result) error {
	if result.AsOfTime.IsZero() ||
		result.NeighborCount < 0 ||
		result.RecentNeighborCount < 0 ||
		result.RecentNeighborCount > result.NeighborCount ||
		result.NeighborCount != len(result.SelectedTrajectoryIDs) {
		return fmt.Errorf("pattern freshness counts or as-of time are invalid")
	}
	if result.NewestNeighborAge < 0 || result.MeanNeighborAge < 0 || result.OldestNeighborAge < 0 {
		return fmt.Errorf("pattern freshness age measurements are invalid")
	}
	if result.NeighborCount == 0 {
		if result.NewestNeighborAge != 0 || result.MeanNeighborAge != 0 || result.OldestNeighborAge != 0 {
			return fmt.Errorf("empty pattern freshness result must have zero age measurements")
		}
		return nil
	}
	if result.NewestNeighborAge > result.MeanNeighborAge || result.MeanNeighborAge > result.OldestNeighborAge {
		return fmt.Errorf("pattern freshness age measurements are invalid")
	}
	return nil
}

func validateResultComponents(result Result) error {
	if !unitInterval(result.Score) || len(result.Components) != len(canonicalComponentNames) {
		return fmt.Errorf("pattern freshness score or components are invalid")
	}
	weightTotal := 0.0
	weightedScore := 0.0
	for index, component := range result.Components {
		if component.Name != canonicalComponentNames[index] {
			return fmt.Errorf("pattern freshness component catalog is invalid at index %d", index)
		}
		if !unitInterval(component.Score) || !finite(component.Weight) || component.Weight < 0 {
			return fmt.Errorf("pattern freshness component is invalid")
		}
		weightTotal += component.Weight
		weightedScore += component.Score * component.Weight
	}
	if math.Abs(weightTotal-1) > scoreComparisonTolerance {
		return fmt.Errorf("pattern freshness component weights do not sum to one")
	}
	if math.Abs(result.Score-clampUnit(weightedScore)) > scoreComparisonTolerance {
		return fmt.Errorf("pattern freshness score does not match weighted components")
	}
	return nil
}

func validateSelectedTrajectoryIDs(items []string) error {
	seen := make(map[string]struct{}, len(items))
	for _, trajectoryID := range items {
		normalized := strings.TrimSpace(trajectoryID)
		if normalized == "" || normalized != trajectoryID {
			return fmt.Errorf("selected trajectory identifiers must be normalized")
		}
		if _, exists := seen[trajectoryID]; exists {
			return fmt.Errorf("duplicate selected trajectory identifier")
		}
		seen[trajectoryID] = struct{}{}
	}
	if !sort.StringsAreSorted(items) {
		return fmt.Errorf("selected trajectory identifiers must be sorted")
	}
	return nil
}

func validateLimitations(items []Notice) error {
	previousKey := ""
	for index, limitation := range items {
		code := strings.TrimSpace(limitation.Code)
		message := strings.TrimSpace(limitation.Message)
		if code == "" || message == "" || code != limitation.Code || message != limitation.Message {
			return fmt.Errorf("pattern freshness limitation is invalid or not normalized")
		}
		key := code + "\x00" + message
		if index > 0 && key <= previousKey {
			return fmt.Errorf("pattern freshness limitations must be sorted and unique")
		}
		previousKey = key
	}
	return nil
}

func validateDecisionShape(result Result) error {
	switch result.Decision {
	case DecisionBlocked:
		if result.Usable || len(result.Limitations) == 0 {
			return fmt.Errorf("blocked freshness result must be unusable and explain limitations")
		}
	case DecisionLimited:
		if !result.Usable || len(result.Limitations) == 0 {
			return fmt.Errorf("limited freshness result must remain usable and explain limitations")
		}
	case DecisionAllowed:
		if !result.Usable || len(result.Limitations) != 0 {
			return fmt.Errorf("allowed freshness result must be usable without limitations")
		}
	}
	return nil
}

func normalizeNotices(items []Notice) []Notice {
	seen := make(map[string]Notice, len(items))
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		message := strings.TrimSpace(item.Message)
		if code == "" || message == "" {
			continue
		}
		key := code + "\x00" + message
		seen[key] = Notice{Code: code, Message: message}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]Notice, 0, len(keys))
	for _, key := range keys {
		result = append(result, seen[key])
	}
	return result
}
