package projectionpatternconfidence

import (
	"fmt"
	"regexp"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

const (
	Version            = "projection-pattern-confidence-v4"
	FingerprintVersion = "projection-pattern-confidence-fingerprint-v4"
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
	ComponentContinuationAgreement ComponentName = "continuation_agreement"

	// Deprecated compatibility aliases.
	ComponentSimilarity = ComponentSimilarityStrength
	ComponentFreshness  = ComponentSimilarityConsistency
)

var canonicalComponentNames = []ComponentName{
	ComponentSimilarityStrength,
	ComponentSupport,
	ComponentSimilarityConsistency,
	ComponentAnchorProximity,
	ComponentContinuationAgreement,
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

type Policy struct {
	MinimumNeighborCount int
	TargetNeighborCount  int

	MinimumSimilarityScore             float64
	MaximumSimilarityStandardDeviation float64
	AnchorDistanceNormalizationKM      float64

	MinimumUsableScore float64

	MediumConfidenceMinimum float64
	HighConfidenceMinimum   float64

	ContinuationAgreementSampleCount       int
	ContinuationDivergenceNormalizationMPS float64
	MaximumContinuationDivergenceMPS       float64

	SimilarityStrengthWeight    float64
	SupportWeight               float64
	SimilarityConsistencyWeight float64
	AnchorProximityWeight       float64
	ContinuationAgreementWeight float64
}

type Result struct {
	Version         string
	Status          Status
	SelectionStatus projectionneighbors.Status
	Usable          bool
	Policy          Policy

	NeighborCount       int
	TargetNeighborCount int

	MeanSimilarityScore         float64
	MinimumSimilarityScore      float64
	SimilarityStandardDeviation float64
	MeanAnchorDistanceKM        float64
	// Deprecated: retained for source compatibility and always zero for evaluator results.
	MeanCandidateAgeSeconds float64

	ContinuationAgreementKnown       bool
	ContinuationAgreementSampleCount int
	ContinuationAgreementPairCount   int
	ContinuationComparisonCount      int
	ContinuationHorizonSeconds       float64
	MeanContinuationSpreadM          float64
	MaximumContinuationSpreadM       float64
	MeanContinuationDivergenceMPS    float64
	MaximumContinuationDivergenceMPS float64

	Score float64
	Level projectioncontract.ConfidenceLevel

	Components            []Component
	SelectedTrajectoryIDs []string
	Limitations           []Notice

	SourceSelectionFingerprint string
	InputFingerprint           string
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
	if result.Version != Version || !result.Status.IsKnown() || !result.SelectionStatus.IsKnown() {
		return fmt.Errorf("pattern confidence version or status is invalid")
	}
	if err := validatePolicy(result.Policy); err != nil {
		return err
	}
	if err := validateResultCounts(result); err != nil {
		return err
	}
	if err := validateAggregateMeasurements(result); err != nil {
		return err
	}
	if err := validateContinuationAgreement(result); err != nil {
		return err
	}
	if err := validateComponents(result); err != nil {
		return err
	}
	if err := validateDecisionSemantics(result); err != nil {
		return err
	}
	if err := validateSelectedTrajectoryIDs(result); err != nil {
		return err
	}
	if err := validateLimitations(result.Limitations); err != nil {
		return err
	}
	if result.SourceSelectionFingerprint != "" &&
		!fingerprintPattern.MatchString(result.SourceSelectionFingerprint) {
		return fmt.Errorf("pattern confidence source selection fingerprint is invalid")
	}
	if !fingerprintPattern.MatchString(result.InputFingerprint) {
		return fmt.Errorf("pattern confidence input fingerprint is invalid")
	}
	return nil
}
