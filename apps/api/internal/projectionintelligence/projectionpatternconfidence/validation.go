package projectionpatternconfidence

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func validateResultCounts(result Result) error {
	if result.NeighborCount < 0 ||
		result.TargetNeighborCount < 1 ||
		result.NeighborCount != len(result.SelectedTrajectoryIDs) {
		return fmt.Errorf("pattern confidence neighbor counts are invalid")
	}
	return nil
}

func validateAggregateMeasurements(result Result) error {
	if !unitInterval(result.MeanSimilarityScore) ||
		!finite(result.MeanCandidateAgeSeconds) ||
		result.MeanCandidateAgeSeconds < 0 ||
		!finite(result.MeanAnchorDistanceKM) ||
		result.MeanAnchorDistanceKM < 0 {
		return fmt.Errorf("pattern confidence aggregate measurements are invalid")
	}
	return nil
}

func validateScoreAndLevel(result Result) error {
	if !unitInterval(result.Score) || !result.Level.IsKnown() {
		return fmt.Errorf("pattern confidence score and level are invalid")
	}
	if result.Score == 0 && result.Level != projectioncontract.ConfidenceLevelNone {
		return fmt.Errorf("zero pattern confidence score requires none level")
	}
	if result.Score > 0 && result.Level == projectioncontract.ConfidenceLevelNone {
		return fmt.Errorf("positive pattern confidence score requires a non-none level")
	}
	return nil
}

func validateComponents(result Result) error {
	if len(result.Components) != len(canonicalComponentNames) {
		return fmt.Errorf("pattern confidence requires the canonical component catalog")
	}

	weightTotal := 0.0
	weightedScore := 0.0
	for index, component := range result.Components {
		if component.Name != canonicalComponentNames[index] {
			return fmt.Errorf(
				"pattern confidence component catalog is invalid at index %d",
				index,
			)
		}
		if !unitInterval(component.Score) || !positiveFinite(component.Weight) {
			return fmt.Errorf("pattern confidence component is invalid")
		}
		weightTotal += component.Weight
		weightedScore += component.Score * component.Weight
	}
	if math.Abs(weightTotal-1) > scoreComparisonTolerance {
		return fmt.Errorf("pattern confidence component weights do not sum to one")
	}
	if math.Abs(result.Score-clampUnit(weightedScore)) > scoreComparisonTolerance {
		return fmt.Errorf("pattern confidence score does not match weighted components")
	}
	return nil
}

func validateSelectedTrajectoryIDs(result Result) error {
	seen := make(map[string]struct{}, len(result.SelectedTrajectoryIDs))
	for _, trajectoryID := range result.SelectedTrajectoryIDs {
		normalized := strings.TrimSpace(trajectoryID)
		if normalized == "" || trajectoryID != normalized {
			return fmt.Errorf("selected trajectory identifiers must be normalized")
		}
		if _, exists := seen[trajectoryID]; exists {
			return fmt.Errorf("duplicate selected trajectory identifier")
		}
		seen[trajectoryID] = struct{}{}
	}
	if !sort.StringsAreSorted(result.SelectedTrajectoryIDs) {
		return fmt.Errorf("selected trajectory identifiers must be sorted")
	}
	return nil
}

func validateLimitations(items []Notice) error {
	previousKey := ""
	for index, item := range items {
		code := strings.TrimSpace(item.Code)
		message := strings.TrimSpace(item.Message)
		if code == "" || message == "" || code != item.Code || message != item.Message {
			return fmt.Errorf("pattern confidence limitation is invalid or not normalized")
		}
		key := code + "\x00" + message
		if index > 0 && key <= previousKey {
			return fmt.Errorf("pattern confidence limitations must be sorted and unique")
		}
		previousKey = key
	}
	return nil
}

func validateStatusSemantics(result Result) error {
	if result.Usable &&
		(result.NeighborCount == 0 || result.Score == 0 ||
			result.Level == projectioncontract.ConfidenceLevelNone) {
		return fmt.Errorf("usable pattern confidence requires positive evidence")
	}

	switch result.Status {
	case StatusUnavailable:
		if result.Usable || len(result.Limitations) == 0 {
			return fmt.Errorf(
				"unavailable pattern confidence must be unusable and explain limitations",
			)
		}
	case StatusComplete:
		if !result.Usable || result.NeighborCount < result.TargetNeighborCount {
			return fmt.Errorf(
				"complete pattern confidence requires usable target support",
			)
		}
	case StatusLimited:
		if !result.Usable || result.NeighborCount == 0 {
			return fmt.Errorf(
				"limited pattern confidence requires usable non-empty support",
			)
		}
	}
	return nil
}
