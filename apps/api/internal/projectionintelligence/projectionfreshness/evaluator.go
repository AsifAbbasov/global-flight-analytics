package projectionfreshness

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

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
		"pattern confidence selected trajectories do not match neighbor selection",
	)
	ErrFreshnessResultInvalid = errors.New(
		"pattern freshness result is invalid",
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
			"validate pattern freshness config: %w",
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
	pattern projectionpatternconfidence.Result,
) (Result, error) {
	if evaluator == nil {
		return Result{},
			ErrFreshnessResultInvalid
	}
	if err := selection.Validate(); err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %v",
				ErrNeighborSelectionInvalid,
				err,
			)
	}
	if err := pattern.Validate(); err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %v",
				ErrPatternConfidenceInvalid,
				err,
			)
	}
	if !sameSelectedTrajectoryIDs(
		selection,
		pattern,
	) {
		return Result{},
			ErrPatternSelectionMismatch
	}

	ages := make(
		[]time.Duration,
		0,
		len(selection.Neighbors),
	)
	selectedIDs := make(
		[]string,
		0,
		len(selection.Neighbors),
	)
	recentCount := 0
	for _, neighbor := range selection.Neighbors {
		ages = append(
			ages,
			neighbor.CandidateAge,
		)
		selectedIDs = append(
			selectedIDs,
			strings.TrimSpace(
				neighbor.TrajectoryID,
			),
		)
		if neighbor.CandidateAge <=
			evaluator.config.
				RecentNeighborAgeLimit {
			recentCount++
		}
	}
	sort.Strings(selectedIDs)
	sort.Slice(
		ages,
		func(left int, right int) bool {
			return ages[left] < ages[right]
		},
	)

	newestAge := time.Duration(0)
	meanAge := time.Duration(0)
	oldestAge := time.Duration(0)
	if len(ages) > 0 {
		newestAge = ages[0]
		oldestAge = ages[len(ages)-1]
		var total int64
		for _, age := range ages {
			total += age.Nanoseconds()
		}
		meanAge = time.Duration(
			total / int64(len(ages)),
		)
	}

	newestScore :=
		ageScore(
			newestAge,
			evaluator.config.
				MaximumNewestNeighborAge,
		)
	meanScore :=
		ageScore(
			meanAge,
			evaluator.config.
				MaximumMeanNeighborAge,
		)
	oldestScore :=
		ageScore(
			oldestAge,
			evaluator.config.
				MaximumOldestNeighborAge,
		)
	recentSupportScore :=
		clampUnit(
			float64(recentCount) /
				float64(
					evaluator.config.
						TargetRecentNeighborCount,
				),
		)

	components := []Component{
		{
			Name:  ComponentNewestAge,
			Score: newestScore,
			Weight: evaluator.config.
				NewestAgeWeight,
		},
		{
			Name:  ComponentMeanAge,
			Score: meanScore,
			Weight: evaluator.config.
				MeanAgeWeight,
		},
		{
			Name:  ComponentOldestAge,
			Score: oldestScore,
			Weight: evaluator.config.
				OldestAgeWeight,
		},
		{
			Name:  ComponentRecentSupport,
			Score: recentSupportScore,
			Weight: evaluator.config.
				RecentSupportWeight,
		},
	}

	score := 0.0
	for _, component := range components {
		score += component.Score *
			component.Weight
	}
	score = clampUnit(score)

	decision := DecisionAllowed
	usable := true
	limitations := make(
		[]Notice,
		0,
		8,
	)

	switch {
	case len(ages) == 0:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code:    "historical_neighbors_unavailable",
				Message: "Historical continuation is blocked because no selected historical neighbors are available.",
			},
		)
	case newestAge >
		evaluator.config.
			MaximumNewestNeighborAge:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code: "newest_historical_neighbor_too_old",
				Message: fmt.Sprintf(
					"Newest selected historical neighbor age %s exceeds the configured maximum %s.",
					newestAge,
					evaluator.config.
						MaximumNewestNeighborAge,
				),
			},
		)
	case meanAge >
		evaluator.config.
			MaximumMeanNeighborAge:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code: "mean_historical_neighbor_age_too_old",
				Message: fmt.Sprintf(
					"Mean selected historical neighbor age %s exceeds the configured maximum %s.",
					meanAge,
					evaluator.config.
						MaximumMeanNeighborAge,
				),
			},
		)
	case oldestAge >
		evaluator.config.
			MaximumOldestNeighborAge:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code: "oldest_historical_neighbor_too_old",
				Message: fmt.Sprintf(
					"Oldest selected historical neighbor age %s exceeds the configured maximum %s.",
					oldestAge,
					evaluator.config.
						MaximumOldestNeighborAge,
				),
			},
		)
	case recentCount <
		evaluator.config.
			MinimumRecentNeighborCount:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code: "recent_historical_neighbor_support_insufficient",
				Message: fmt.Sprintf(
					"Recent historical neighbor count %d is below the configured minimum %d.",
					recentCount,
					evaluator.config.
						MinimumRecentNeighborCount,
				),
			},
		)
	case score <
		evaluator.config.
			MinimumUsableScore:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code: "pattern_freshness_score_below_minimum",
				Message: fmt.Sprintf(
					"Pattern freshness score %.6f is below the configured usable minimum %.6f.",
					score,
					evaluator.config.
						MinimumUsableScore,
				),
			},
		)
	case score <
		evaluator.config.
			CompleteScoreMinimum:
		decision = DecisionLimited
		limitations = append(
			limitations,
			Notice{
				Code: "pattern_freshness_limited",
				Message: fmt.Sprintf(
					"Pattern freshness score %.6f is below the configured complete threshold %.6f.",
					score,
					evaluator.config.
						CompleteScoreMinimum,
				),
			},
		)
	}

	if pattern.Status !=
		projectionpatternconfidence.
			StatusComplete &&
		usable {
		decision = DecisionLimited
		limitations = append(
			limitations,
			Notice{
				Code:    "pattern_confidence_not_complete",
				Message: "Pattern confidence remains usable but is not complete, so freshness approval is limited.",
			},
		)
	}
	if selection.Status !=
		projectionneighbors.StatusComplete &&
		usable {
		decision = DecisionLimited
		limitations = append(
			limitations,
			Notice{
				Code:    "neighbor_selection_not_complete",
				Message: "Historical neighbor selection remains usable but did not fill the configured selection target.",
			},
		)
	}

	result := Result{
		Version:  Version,
		Decision: decision,
		Usable:   usable,

		AsOfTime: selection.AsOfTime.UTC(),

		NeighborCount:       len(selection.Neighbors),
		RecentNeighborCount: recentCount,

		NewestNeighborAge: newestAge,
		MeanNeighborAge:   meanAge,
		OldestNeighborAge: oldestAge,

		Score:      score,
		Components: components,

		SelectedTrajectoryIDs: selectedIDs,
		Limitations: normalizeNotices(
			limitations,
		),
		InputFingerprint: freshnessFingerprint(
			selection,
			pattern,
			evaluator.config,
		),
	}

	if err := result.Validate(); err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %v",
				ErrFreshnessResultInvalid,
				err,
			)
	}

	return result.Clone(), nil
}

func sameSelectedTrajectoryIDs(
	selection projectionneighbors.Result,
	pattern projectionpatternconfidence.Result,
) bool {
	selectionIDs := make(
		[]string,
		0,
		len(selection.Neighbors),
	)
	for _, neighbor := range selection.Neighbors {
		selectionIDs = append(
			selectionIDs,
			strings.TrimSpace(
				neighbor.TrajectoryID,
			),
		)
	}
	sort.Strings(selectionIDs)

	if len(selectionIDs) !=
		len(pattern.SelectedTrajectoryIDs) {
		return false
	}
	for index := range selectionIDs {
		if selectionIDs[index] !=
			pattern.SelectedTrajectoryIDs[index] {
			return false
		}
	}

	return true
}

func ageScore(
	age time.Duration,
	maximum time.Duration,
) float64 {
	if maximum <= 0 ||
		age < 0 {
		return 0
	}

	return clampUnit(
		1 -
			float64(age)/
				float64(maximum),
	)
}
