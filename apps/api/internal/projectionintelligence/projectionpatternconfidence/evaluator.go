package projectionpatternconfidence

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

var (
	ErrSelectionInvalid = errors.New(
		"historical neighbor selection is invalid",
	)
	ErrPatternConfidenceResultInvalid = errors.New(
		"pattern confidence result is invalid",
	)
)

type Evaluator struct {
	config Config
}

func New(
	config Config,
) (*Evaluator, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate pattern confidence config: %w",
			err,
		)
	}

	return &Evaluator{
		config: config,
	}, nil
}

func (
	evaluator *Evaluator,
) Evaluate(
	selection projectionneighbors.Result,
) (Result, error) {
	if evaluator == nil {
		return Result{},
			ErrPatternConfidenceResultInvalid
	}
	if err := selection.Validate(); err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %v",
				ErrSelectionInvalid,
				err,
			)
	}

	neighborCount := len(
		selection.Neighbors,
	)
	trajectoryIDs := make(
		[]string,
		0,
		neighborCount,
	)
	for _, neighbor := range selection.Neighbors {
		trajectoryIDs = append(
			trajectoryIDs,
			strings.TrimSpace(
				neighbor.TrajectoryID,
			),
		)
	}
	sort.Strings(trajectoryIDs)

	components := []Component{
		{
			Name: ComponentSimilarity,
			Weight: evaluator.config.
				SimilarityWeight,
		},
		{
			Name: ComponentSupport,
			Weight: evaluator.config.
				SupportWeight,
		},
		{
			Name: ComponentFreshness,
			Weight: evaluator.config.
				FreshnessWeight,
		},
		{
			Name: ComponentAnchorProximity,
			Weight: evaluator.config.
				AnchorProximityWeight,
		},
	}

	limitations := make(
		[]Notice,
		0,
		4,
	)
	for _, limitation := range selection.Limitations {
		limitations = append(
			limitations,
			Notice{
				Code: "neighbor_selection_" +
					limitation.Code,
				Message: limitation.Message,
			},
		)
	}

	meanSimilarity := 0.0
	meanAgeSeconds := 0.0
	meanAnchorDistanceKM := 0.0
	freshnessScore := 0.0

	for _, neighbor := range selection.Neighbors {
		meanSimilarity +=
			neighbor.SimilarityScore
		meanAgeSeconds +=
			neighbor.CandidateAge.Seconds()
		meanAnchorDistanceKM +=
			neighbor.AnchorDistanceKM

		freshness := 1 -
			neighbor.CandidateAge.Seconds()/
				evaluator.config.
					MaximumCandidateAge.Seconds()
		freshnessScore +=
			clampUnit(freshness)
	}

	if neighborCount > 0 {
		divisor := float64(
			neighborCount,
		)
		meanSimilarity /= divisor
		meanAgeSeconds /= divisor
		meanAnchorDistanceKM /=
			divisor
		freshnessScore /= divisor
	}

	supportScore := clampUnit(
		float64(neighborCount) /
			float64(
				evaluator.config.
					TargetNeighborCount,
			),
	)
	anchorProximityScore := 0.0
	if neighborCount > 0 {
		anchorProximityScore =
			clampUnit(
				1 -
					meanAnchorDistanceKM/
						evaluator.config.
							MaximumMeanAnchorDistanceKM,
			)
	}

	components[0].Score =
		clampUnit(meanSimilarity)
	components[1].Score =
		supportScore
	components[2].Score =
		freshnessScore
	components[3].Score =
		anchorProximityScore

	score := 0.0
	for _, component := range components {
		score += component.Score *
			component.Weight
	}
	score = clampUnit(score)

	usable := neighborCount >=
		evaluator.config.
			MinimumNeighborCount &&
		score >=
			evaluator.config.
				MinimumUsableScore

	status := StatusUnavailable
	switch {
	case !usable:
		if neighborCount <
			evaluator.config.
				MinimumNeighborCount {
			limitations = append(
				limitations,
				Notice{
					Code: "insufficient_historical_neighbor_support",
					Message: fmt.Sprintf(
						"Pattern requires at least %d neighbors, but %d were selected.",
						evaluator.config.
							MinimumNeighborCount,
						neighborCount,
					),
				},
			)
		}
		if score <
			evaluator.config.
				MinimumUsableScore {
			limitations = append(
				limitations,
				Notice{
					Code: "pattern_confidence_below_minimum",
					Message: fmt.Sprintf(
						"Pattern confidence score %.6f is below the configured minimum %.6f.",
						score,
						evaluator.config.
							MinimumUsableScore,
					),
				},
			)
		}

	case neighborCount >=
		evaluator.config.
			TargetNeighborCount &&
		selection.Status ==
			projectionneighbors.
				StatusComplete:
		status = StatusComplete

	default:
		status = StatusLimited
		limitations = append(
			limitations,
			Notice{
				Code:    "pattern_support_partial",
				Message: "Historical pattern is usable but does not satisfy complete target support.",
			},
		)
	}

	result := Result{
		Version: Version,
		Status:  status,
		Usable:  usable,

		NeighborCount: neighborCount,
		TargetNeighborCount: evaluator.config.
			TargetNeighborCount,

		MeanSimilarityScore:     meanSimilarity,
		MeanCandidateAgeSeconds: meanAgeSeconds,
		MeanAnchorDistanceKM:    meanAnchorDistanceKM,

		Score: score,
		Level: evaluator.
			confidenceLevel(score),

		Components: append(
			[]Component(nil),
			components...,
		),
		SelectedTrajectoryIDs: trajectoryIDs,
		Limitations: normalizeNotices(
			limitations,
		),

		InputFingerprint: inputFingerprint(
			selection,
			evaluator.config,
		),
	}

	if err := result.Validate(); err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %v",
				ErrPatternConfidenceResultInvalid,
				err,
			)
	}

	return result.Clone(), nil
}

func (
	evaluator *Evaluator,
) confidenceLevel(
	score float64,
) projectioncontract.ConfidenceLevel {
	switch {
	case score >= evaluator.config.
		HighConfidenceMinimum:
		return projectioncontract.
			ConfidenceLevelHigh
	case score >= evaluator.config.
		MediumConfidenceMinimum:
		return projectioncontract.
			ConfidenceLevelMedium
	case score > 0:
		return projectioncontract.
			ConfidenceLevelLow
	default:
		return projectioncontract.
			ConfidenceLevelNone
	}
}

func clampUnit(
	value float64,
) float64 {
	if !finite(value) ||
		value <= 0 {
		return 0
	}
	if value >= 1 {
		return 1
	}

	return value
}

func normalizeNotices(
	items []Notice,
) []Notice {
	seen := make(
		map[string]Notice,
		len(items),
	)
	for _, item := range items {
		code := strings.TrimSpace(
			item.Code,
		)
		message := strings.TrimSpace(
			item.Message,
		)
		if code == "" ||
			message == "" {
			continue
		}
		key := code + "\x00" +
			message
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
