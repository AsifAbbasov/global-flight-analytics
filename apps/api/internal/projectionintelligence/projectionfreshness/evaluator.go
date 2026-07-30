package projectionfreshness

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

var (
	ErrNeighborSelectionInvalid = errors.New(
		"historical neighbor selection is invalid",
	)
	ErrPatternConfidenceInvalid = errors.New(
		"historical pattern confidence is invalid",
	)
	ErrPatternSelectionMismatch = errors.New(
		"pattern confidence lineage does not match neighbor selection",
	)
	ErrFreshnessResultInvalid = errors.New(
		"pattern freshness result is invalid",
	)
)

type Evaluator struct {
	config Config
}

func New(config Config) (*Evaluator, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate pattern freshness config: %w", err)
	}
	return &Evaluator{config: config}, nil
}

func (evaluator *Evaluator) Evaluate(
	selection projectionneighbors.Result,
	pattern projectionpatternconfidence.Result,
) (Result, error) {
	if evaluator == nil {
		return Result{}, ErrFreshnessResultInvalid
	}
	if err := selection.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrNeighborSelectionInvalid, err)
	}
	if err := pattern.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrPatternConfidenceInvalid, err)
	}
	if err := validateLineage(selection, pattern); err != nil {
		return Result{}, err
	}

	metrics := measureFreshness(selection, evaluator.config)
	components := buildFreshnessComponents(metrics, evaluator.config)
	score := weightedComponentScore(components)
	decision := evaluateFreshnessPolicy(
		metrics,
		score,
		selection,
		pattern,
		evaluator.config,
	)
	result := assembleFreshnessResult(
		selection,
		pattern,
		metrics,
		components,
		score,
		decision,
		evaluator.config,
	)
	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrFreshnessResultInvalid, err)
	}
	return result.Clone(), nil
}

func validateLineage(
	selection projectionneighbors.Result,
	pattern projectionpatternconfidence.Result,
) error {
	if pattern.SourceSelectionFingerprint == "" ||
		pattern.SourceSelectionFingerprint != selection.InputFingerprint {
		return ErrPatternSelectionMismatch
	}
	if !sameSelectedTrajectoryIDs(selection, pattern) {
		return ErrPatternSelectionMismatch
	}
	return nil
}

func sameSelectedTrajectoryIDs(
	selection projectionneighbors.Result,
	pattern projectionpatternconfidence.Result,
) bool {
	selectionIDs := make([]string, 0, len(selection.Neighbors))
	for _, neighbor := range selection.Neighbors {
		selectionIDs = append(selectionIDs, strings.TrimSpace(neighbor.TrajectoryID))
	}
	sort.Strings(selectionIDs)
	if len(selectionIDs) != len(pattern.SelectedTrajectoryIDs) {
		return false
	}
	for index := range selectionIDs {
		if selectionIDs[index] != pattern.SelectedTrajectoryIDs[index] {
			return false
		}
	}
	return true
}

func assembleFreshnessResult(
	selection projectionneighbors.Result,
	pattern projectionpatternconfidence.Result,
	metrics freshnessMetrics,
	components []Component,
	score float64,
	decision freshnessDecision,
	config Config,
) Result {
	return Result{
		Version:  Version,
		Decision: decision.decision,
		Usable:   decision.usable,

		AsOfTime: selection.AsOfTime.UTC(),

		NeighborCount:       len(metrics.ages),
		RecentNeighborCount: metrics.recentCount,

		NewestNeighborAge: metrics.newestAge,
		MeanNeighborAge:   metrics.meanAge,
		OldestNeighborAge: metrics.oldestAge,

		Score:      score,
		Components: append([]Component(nil), components...),

		SelectedTrajectoryIDs: append([]string(nil), metrics.selectedIDs...),
		Limitations:           append([]Notice(nil), decision.limitations...),
		InputFingerprint: freshnessFingerprint(
			selection,
			pattern,
			config,
		),
	}
}
