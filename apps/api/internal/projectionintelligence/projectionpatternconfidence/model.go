package projectionpatternconfidence

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

const (
	Version            = "projection-pattern-confidence-v1"
	FingerprintVersion = "projection-pattern-confidence-fingerprint-v1"
)

type Status string

const (
	StatusUnavailable Status = "unavailable"
	StatusLimited     Status = "limited"
	StatusComplete    Status = "complete"
)

func (status Status) IsKnown() bool {
	switch status {
	case StatusUnavailable,
		StatusLimited,
		StatusComplete:
		return true
	default:
		return false
	}
}

type ComponentName string

const (
	ComponentSimilarity      ComponentName = "similarity"
	ComponentSupport         ComponentName = "support"
	ComponentFreshness       ComponentName = "freshness"
	ComponentAnchorProximity ComponentName = "anchor_proximity"
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
	Version string
	Status  Status
	Usable  bool

	NeighborCount       int
	TargetNeighborCount int

	MeanSimilarityScore     float64
	MeanCandidateAgeSeconds float64
	MeanAnchorDistanceKM    float64

	Score float64
	Level projectioncontract.ConfidenceLevel

	Components            []Component
	SelectedTrajectoryIDs []string
	Limitations           []Notice

	InputFingerprint string
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
		!result.Status.IsKnown() {
		return fmt.Errorf(
			"pattern confidence version or status is invalid",
		)
	}
	if result.NeighborCount < 0 ||
		result.TargetNeighborCount < 1 ||
		result.NeighborCount !=
			len(result.SelectedTrajectoryIDs) {
		return fmt.Errorf(
			"pattern confidence neighbor counts are invalid",
		)
	}
	if !unitInterval(result.Score) ||
		!result.Level.IsKnown() ||
		(result.Score == 0 &&
			result.Level !=
				projectioncontract.
					ConfidenceLevelNone) ||
		(result.Score > 0 &&
			result.Level ==
				projectioncontract.
					ConfidenceLevelNone) {
		return fmt.Errorf(
			"pattern confidence score and level are inconsistent",
		)
	}
	if !finite(
		result.MeanSimilarityScore,
	) ||
		result.MeanSimilarityScore < 0 ||
		result.MeanSimilarityScore > 1 ||
		!finite(
			result.MeanCandidateAgeSeconds,
		) ||
		result.MeanCandidateAgeSeconds < 0 ||
		!finite(
			result.MeanAnchorDistanceKM,
		) ||
		result.MeanAnchorDistanceKM < 0 {
		return fmt.Errorf(
			"pattern confidence aggregate measurements are invalid",
		)
	}
	if len(result.Components) != 4 {
		return fmt.Errorf(
			"pattern confidence requires four components",
		)
	}
	seenComponents := make(
		map[ComponentName]struct{},
		4,
	)
	weightTotal := 0.0
	for _, component := range result.Components {
		if _, exists := seenComponents[component.Name]; exists {
			return fmt.Errorf(
				"duplicate pattern confidence component",
			)
		}
		seenComponents[component.Name] =
			struct{}{}
		if !unitInterval(component.Score) ||
			!finite(component.Weight) ||
			component.Weight < 0 {
			return fmt.Errorf(
				"pattern confidence component is invalid",
			)
		}
		weightTotal += component.Weight
	}
	if absolute(
		weightTotal-1,
	) > 1e-9 {
		return fmt.Errorf(
			"pattern confidence component weights do not sum to one",
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
				"pattern confidence limitation is invalid",
			)
		}
	}
	if !fingerprintPattern.MatchString(
		result.InputFingerprint,
	) {
		return fmt.Errorf(
			"pattern confidence input fingerprint is invalid",
		)
	}

	switch result.Status {
	case StatusUnavailable:
		if result.Usable ||
			len(result.Limitations) == 0 {
			return fmt.Errorf(
				"unavailable pattern confidence must be unusable and explain limitations",
			)
		}
	case StatusComplete:
		if !result.Usable ||
			result.NeighborCount <
				result.TargetNeighborCount {
			return fmt.Errorf(
				"complete pattern confidence requires usable target support",
			)
		}
	case StatusLimited:
		if !result.Usable {
			return fmt.Errorf(
				"limited pattern confidence must remain usable",
			)
		}
	}

	return nil
}

func absolute(value float64) float64 {
	if value < 0 {
		return -value
	}

	return value
}
