package projectionpatternconfidence

import (
	"fmt"
	"regexp"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

const (
	Version            = "projection-pattern-confidence-v3"
	FingerprintVersion = "projection-pattern-confidence-fingerprint-v3"
)

const scoreComparisonTolerance = 1e-9

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

type ComponentName string

const (
	ComponentSimilarityStrength    ComponentName = "similarity_strength"
	ComponentSupport               ComponentName = "support"
	ComponentSimilarityConsistency ComponentName = "similarity_consistency"
	ComponentAnchorProximity       ComponentName = "anchor_proximity"

	// Deprecated compatibility aliases.
	ComponentSimilarity = ComponentSimilarityStrength
	ComponentFreshness  = ComponentSimilarityConsistency
)

var canonicalComponentNames = []ComponentName{
	ComponentSimilarityStrength,
	ComponentSupport,
	ComponentSimilarityConsistency,
	ComponentAnchorProximity,
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
	Version string
	Status  Status
	Usable  bool

	NeighborCount       int
	TargetNeighborCount int

	MeanSimilarityScore         float64
	MinimumSimilarityScore      float64
	SimilarityStandardDeviation float64
	MeanAnchorDistanceKM        float64
	// Deprecated: retained for source compatibility and always zero for evaluator results.
	MeanCandidateAgeSeconds float64

	Score float64
	Level projectioncontract.ConfidenceLevel

	Components            []Component
	SelectedTrajectoryIDs []string
	Limitations           []Notice

	InputFingerprint string
}

func (result Result) Clone() Result {
	cloned := result
	cloned.Components = append([]Component(nil), result.Components...)
	cloned.SelectedTrajectoryIDs = append(
		[]string(nil),
		result.SelectedTrajectoryIDs...,
	)
	cloned.Limitations = append([]Notice(nil), result.Limitations...)
	return cloned
}

var fingerprintPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func (result Result) Validate() error {
	if result.Version != Version || !result.Status.IsKnown() {
		return fmt.Errorf("pattern confidence version or status is invalid")
	}
	if err := validateResultCounts(result); err != nil {
		return err
	}
	if err := validateAggregateMeasurements(result); err != nil {
		return err
	}
	if err := validateScoreAndLevel(result); err != nil {
		return err
	}
	if err := validateComponents(result); err != nil {
		return err
	}
	if err := validateSelectedTrajectoryIDs(result); err != nil {
		return err
	}
	if err := validateLimitations(result.Limitations); err != nil {
		return err
	}
	if !fingerprintPattern.MatchString(result.InputFingerprint) {
		return fmt.Errorf("pattern confidence input fingerprint is invalid")
	}
	if err := validateStatusSemantics(result); err != nil {
		return err
	}
	return nil
}
