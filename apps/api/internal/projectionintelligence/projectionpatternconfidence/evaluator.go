package projectionpatternconfidence

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

var (
	ErrSelectionInvalid               = errors.New("historical neighbor selection is invalid")
	ErrPatternConfidenceResultInvalid = errors.New(
		"pattern confidence result is invalid",
	)
)

type Evaluator struct {
	config Config
}

func New(config Config) (*Evaluator, error) {
	normalized, err := normalizeAndValidateConfig(config)
	if err != nil {
		return nil, fmt.Errorf("validate pattern confidence config: %w", err)
	}
	return &Evaluator{config: normalized}, nil
}

// Evaluate preserves the legacy interface but deliberately does not authorize
// historical projection because future continuation agreement is unavailable.
func (evaluator *Evaluator) Evaluate(
	selection projectionneighbors.Result,
) (Result, error) {
	return evaluator.evaluate(
		selection,
		continuationAgreementEvidence{},
	)
}

func (evaluator *Evaluator) EvaluateWithContinuations(
	selection projectionneighbors.Result,
	candidates []trajectory.FlightTrajectory,
) (Result, error) {
	if evaluator == nil {
		return Result{}, ErrPatternConfidenceResultInvalid
	}
	if err := selection.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrSelectionInvalid, err)
	}

	continuation, err := extractContinuationAgreement(
		selection,
		candidates,
		evaluator.config,
	)
	if err != nil {
		return Result{}, err
	}
	return evaluator.evaluate(selection, continuation)
}

func (evaluator *Evaluator) evaluate(
	selection projectionneighbors.Result,
	continuation continuationAgreementEvidence,
) (Result, error) {
	if evaluator == nil {
		return Result{}, ErrPatternConfidenceResultInvalid
	}
	if err := selection.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrSelectionInvalid, err)
	}

	evidence := extractPatternEvidence(
		selection,
		evaluator.config,
		continuation,
	)
	components := buildComponents(evaluator.config, evidence)
	score := weightedComponentScore(components)
	decision := decidePattern(selection, evidence, score, evaluator.config)

	result := Result{
		Version:         Version,
		Status:          decision.status,
		SelectionStatus: selection.Status,
		Usable:          decision.usable,
		Policy:          evaluator.config.policySnapshot(),

		NeighborCount:       len(evidence.neighbors),
		TargetNeighborCount: evaluator.config.TargetNeighborCount,

		MeanSimilarityScore:         evidence.meanSimilarityScore,
		MinimumSimilarityScore:      evidence.minimumSimilarityScore,
		SimilarityStandardDeviation: evidence.similarityStandardDeviation,
		MeanAnchorDistanceKM:        evidence.meanAnchorDistanceKM,
		MeanCandidateAgeSeconds:     0,

		ContinuationAgreementKnown:       continuation.known,
		ContinuationAgreementSampleCount: continuation.sampleCount,
		ContinuationAgreementPairCount:   continuation.pairCount,
		ContinuationComparisonCount:      continuation.comparisonCount,
		ContinuationHorizonSeconds:       continuation.horizonSeconds,
		MeanContinuationSpreadM:          continuation.meanSpreadM,
		MaximumContinuationSpreadM:       continuation.maximumSpreadM,
		MeanContinuationDivergenceMPS:    continuation.meanDivergenceMPS,
		MaximumContinuationDivergenceMPS: continuation.maximumDivergenceMPS,

		Score: score,
		Level: confidenceLevelForDecision(
			score,
			decision.status,
			evaluator.config.MediumConfidenceMinimum,
			evaluator.config.HighConfidenceMinimum,
		),

		Components:            append([]Component(nil), components...),
		SelectedTrajectoryIDs: append([]string(nil), evidence.trajectoryIDs...),
		Limitations:           append([]Notice(nil), decision.limitations...),

		InputFingerprint: inputFingerprint(
			selection,
			evaluator.config,
			evidence,
		),
	}
	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrPatternConfidenceResultInvalid,
			err,
		)
	}
	return result.Clone(), nil
}

func clampUnit(value float64) float64 {
	if !finite(value) || value <= 0 {
		return 0
	}
	if value >= 1 {
		return 1
	}
	return value
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
