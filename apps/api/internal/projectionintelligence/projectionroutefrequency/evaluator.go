package projectionroutefrequency

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

var (
	ErrRouteContractInvalid = errors.New(
		"route intelligence contract is invalid",
	)
	ErrRouteHistoryInvalid = errors.New(
		"route history summary is invalid",
	)
	ErrRouteHistoryAsOfMismatch = errors.New(
		"route history as-of time must equal route intelligence as-of time",
	)
	ErrRouteHistoryKeyMismatch = errors.New(
		"route history key does not match the resolved route",
	)
	ErrRouteFrequencyResultInvalid = errors.New(
		"route-frequency guard result is invalid",
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
			"validate route-frequency config: %w",
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
	route routecontract.Result,
	history HistorySummary,
) (Result, error) {
	if evaluator == nil {
		return Result{},
			ErrRouteFrequencyResultInvalid
	}

	routeReport :=
		routecontract.Validate(route)
	if routeReport.Status !=
		routecontract.ValidationStatusValid {
		return Result{},
			fmt.Errorf(
				"%w: %#v",
				ErrRouteContractInvalid,
				routeReport.Issues,
			)
	}
	if err := history.Validate(); err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %v",
				ErrRouteHistoryInvalid,
				err,
			)
	}
	if !history.AsOfTime.UTC().Equal(
		route.Window.AsOfTime.UTC(),
	) {
		return Result{},
			ErrRouteHistoryAsOfMismatch
	}

	expectedRouteKey,
		routeAvailable :=
		resolvedRouteKey(route)
	if routeAvailable &&
		!strings.EqualFold(
			expectedRouteKey,
			history.RouteKey,
		) {
		return Result{},
			ErrRouteHistoryKeyMismatch
	}

	latestAge := history.AsOfTime.UTC().
		Sub(
			history.LastObservedAt.UTC(),
		)
	if history.ObservationCount == 0 {
		latestAge = 0
	}

	observationScore := clampUnit(
		float64(
			history.ObservationCount,
		) /
			float64(
				evaluator.config.
					TargetObservationCount,
			),
	)
	distinctDayScore := clampUnit(
		float64(
			history.DistinctDayCount,
		) /
			float64(
				evaluator.config.
					TargetDistinctDayCount,
			),
	)
	recentObservationScore := clampUnit(
		float64(
			history.RecentObservationCount,
		) /
			float64(
				evaluator.config.
					TargetRecentObservationCount,
			),
	)
	latestObservationScore := 0.0
	if history.ObservationCount > 0 {
		latestObservationScore = clampUnit(
			1 -
				float64(latestAge)/
					float64(
						evaluator.config.
							MaximumLatestObservationAge,
					),
		)
	}
	routeConfidenceScore :=
		clampUnit(
			route.Confidence.Score,
		)

	components := []Component{
		{
			Name:  ComponentObservationCount,
			Score: observationScore,
			Weight: evaluator.config.
				ObservationCountWeight,
		},
		{
			Name:  ComponentDistinctDays,
			Score: distinctDayScore,
			Weight: evaluator.config.
				DistinctDayWeight,
		},
		{
			Name:  ComponentRecentObservations,
			Score: recentObservationScore,
			Weight: evaluator.config.
				RecentObservationWeight,
		},
		{
			Name:  ComponentLatestObservation,
			Score: latestObservationScore,
			Weight: evaluator.config.
				LatestObservationWeight,
		},
		{
			Name:  ComponentRouteConfidence,
			Score: routeConfidenceScore,
			Weight: evaluator.config.
				RouteConfidenceWeight,
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
		10,
	)

	switch {
	case !routeAvailable:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code:    "complete_route_unavailable",
				Message: "Historical route continuation is blocked because Route Intelligence did not resolve both route endpoints.",
			},
		)
	case route.Summary.SameAirport:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code:    "same_airport_route_not_supported",
				Message: "Historical route continuation is blocked because origin and destination resolve to the same airport.",
			},
		)
	case route.Confidence.Score <
		evaluator.config.
			MinimumRouteConfidenceScore:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code: "route_confidence_below_minimum",
				Message: fmt.Sprintf(
					"Route confidence %.6f is below the configured minimum %.6f.",
					route.Confidence.Score,
					evaluator.config.
						MinimumRouteConfidenceScore,
				),
			},
		)
	case history.ObservationCount <
		evaluator.config.
			MinimumObservationCount:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code: "route_observation_count_below_minimum",
				Message: fmt.Sprintf(
					"Historical route observation count %d is below the configured minimum %d.",
					history.ObservationCount,
					evaluator.config.
						MinimumObservationCount,
				),
			},
		)
	case history.DistinctDayCount <
		evaluator.config.
			MinimumDistinctDayCount:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code: "route_distinct_day_count_below_minimum",
				Message: fmt.Sprintf(
					"Historical route distinct-day count %d is below the configured minimum %d.",
					history.DistinctDayCount,
					evaluator.config.
						MinimumDistinctDayCount,
				),
			},
		)
	case history.RecentObservationCount <
		evaluator.config.
			MinimumRecentObservationCount:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code: "recent_route_observation_count_below_minimum",
				Message: fmt.Sprintf(
					"Recent route observation count %d in the configured %s window is below the minimum %d.",
					history.RecentObservationCount,
					evaluator.config.RecentWindow,
					evaluator.config.
						MinimumRecentObservationCount,
				),
			},
		)
	case history.ObservationCount > 0 &&
		latestAge >
			evaluator.config.
				MaximumLatestObservationAge:
		decision = DecisionBlocked
		usable = false
		limitations = append(
			limitations,
			Notice{
				Code: "latest_route_observation_too_old",
				Message: fmt.Sprintf(
					"Latest route observation age %s exceeds the configured maximum %s.",
					latestAge,
					evaluator.config.
						MaximumLatestObservationAge,
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
				Code: "route_frequency_score_below_minimum",
				Message: fmt.Sprintf(
					"Route-frequency score %.6f is below the configured usable minimum %.6f.",
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
				Code: "route_frequency_support_limited",
				Message: fmt.Sprintf(
					"Route-frequency score %.6f is below the configured complete threshold %.6f.",
					score,
					evaluator.config.
						CompleteScoreMinimum,
				),
			},
		)
	}

	if route.Status !=
		routecontract.RouteStatusComplete &&
		usable {
		decision = DecisionLimited
		limitations = append(
			limitations,
			Notice{
				Code:    "route_intelligence_not_complete",
				Message: "Route Intelligence is not complete, so route-frequency approval is limited.",
			},
		)
	}

	result := Result{
		Version:  Version,
		Decision: decision,
		Usable:   usable,

		RouteKey: strings.ToUpper(
			strings.TrimSpace(
				history.RouteKey,
			),
		),
		AsOfTime: history.AsOfTime.UTC(),

		ObservationCount:       history.ObservationCount,
		DistinctFlightCount:    history.DistinctFlightCount,
		DistinctDayCount:       history.DistinctDayCount,
		RecentObservationCount: history.RecentObservationCount,
		LatestObservationAge:   latestAge,
		RouteConfidenceScore:   route.Confidence.Score,

		Score:      score,
		Components: components,
		Limitations: normalizeNotices(
			limitations,
		),

		HistoryInputFingerprint: history.InputFingerprint,
		InputFingerprint: routeFrequencyFingerprint(
			route,
			history,
			evaluator.config,
		),
	}

	if err := result.Validate(); err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %v",
				ErrRouteFrequencyResultInvalid,
				err,
			)
	}

	return result.Clone(), nil
}

func resolvedRouteKey(
	route routecontract.Result,
) (string, bool) {
	if route.Origin == nil ||
		route.Destination == nil {
		return "", false
	}

	origin := strings.ToUpper(
		strings.TrimSpace(
			route.Origin.Airport.ICAOCode,
		),
	)
	destination := strings.ToUpper(
		strings.TrimSpace(
			route.Destination.Airport.ICAOCode,
		),
	)
	if len(origin) != 4 ||
		len(destination) != 4 {
		return "", false
	}

	return origin + ">" + destination,
		true
}
