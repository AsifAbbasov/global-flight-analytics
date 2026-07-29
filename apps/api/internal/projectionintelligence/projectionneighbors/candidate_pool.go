package projectionneighbors

import (
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type preparedCandidate struct {
	Trajectory trajectory.FlightTrajectory
	ID         string
	StartTime  time.Time
	EndTime    time.Time
	Age        time.Duration
}

type preparedCandidatePool struct {
	Candidates               []preparedCandidate
	Rejections               []Rejection
	ExcludedFuturePointCount int
	Truncated                bool
}

func prepareCandidatePool(
	inputs []trajectory.FlightTrajectory,
	currentID string,
	currentStartTime time.Time,
	asOfTime time.Time,
	config Config,
	routeScope routeScopeIndex,
) preparedCandidatePool {
	ordered := append(
		[]trajectory.FlightTrajectory(nil),
		inputs...,
	)
	sort.SliceStable(
		ordered,
		func(left int, right int) bool {
			return strings.TrimSpace(ordered[left].ID) <
				strings.TrimSpace(ordered[right].ID)
		},
	)

	candidateIDCounts := make(
		map[string]int,
		len(ordered),
	)
	for _, candidate := range ordered {
		candidateIDCounts[strings.TrimSpace(candidate.ID)]++
	}

	pool := preparedCandidatePool{
		Candidates: make(
			[]preparedCandidate,
			0,
			len(ordered),
		),
		Rejections: make(
			[]Rejection,
			0,
			len(ordered),
		),
	}

	for _, candidateInput := range ordered {
		candidateID := strings.TrimSpace(candidateInput.ID)
		if candidateID == "" {
			pool.Rejections = append(
				pool.Rejections,
				rejection(
					candidateID,
					RejectionIdentifierMissing,
					"Historical candidate identifier is required.",
				),
			)
			continue
		}
		if candidateID == currentID {
			pool.Rejections = append(
				pool.Rejections,
				rejection(
					candidateID,
					RejectionSameTrajectory,
					"Current trajectory cannot be selected as its own historical neighbor.",
				),
			)
			continue
		}
		if candidateIDCounts[candidateID] > 1 {
			pool.Rejections = append(
				pool.Rejections,
				rejection(
					candidateID,
					RejectionDuplicateCandidate,
					"Every historical candidate with a duplicated identifier was rejected.",
				),
			)
			continue
		}

		routeEvidence := routeScope.evidenceByCandidate[candidateID]
		if !routeEvidence.Route.Equal(routeScope.scope.Route) {
			pool.Rejections = append(
				pool.Rejections,
				rejection(
					candidateID,
					RejectionCrossRoute,
					"Historical candidate route does not match the current origin-destination scope.",
				),
			)
			continue
		}

		candidate, excludedFuturePointCount := snapshotAt(
			candidateInput,
			asOfTime,
		)
		pool.ExcludedFuturePointCount += excludedFuturePointCount
		if len(candidate.Points) < config.MinimumCurrentPointCount+1 {
			pool.Rejections = append(
				pool.Rejections,
				rejection(
					candidateID,
					RejectionInsufficientPoints,
					"Historical candidate does not contain enough usable points for a comparable prefix and continuation.",
				),
			)
			continue
		}

		candidateStartTime := candidate.Points[0].ObservedAt.UTC()
		candidateEndTime := candidate.Points[len(candidate.Points)-1].ObservedAt.UTC()
		if !candidateEndTime.Before(currentStartTime) {
			pool.Rejections = append(
				pool.Rejections,
				rejection(
					candidateID,
					RejectionNotHistorical,
					"Historical candidate must end before the current trajectory begins.",
				),
			)
			continue
		}

		candidateAge := asOfTime.Sub(candidateEndTime)
		if candidateAge < 0 {
			pool.Rejections = append(
				pool.Rejections,
				rejection(
					candidateID,
					RejectionNotHistorical,
					"Historical candidate contains evidence after the as-of time.",
				),
			)
			continue
		}
		if config.MaximumCandidateAge > 0 &&
			candidateAge > config.MaximumCandidateAge {
			pool.Rejections = append(
				pool.Rejections,
				rejection(
					candidateID,
					RejectionTooOld,
					"Historical candidate exceeds the configured maximum age.",
				),
			)
			continue
		}

		pool.Candidates = append(
			pool.Candidates,
			preparedCandidate{
				Trajectory: candidate,
				ID:         candidateID,
				StartTime:  candidateStartTime,
				EndTime:    candidateEndTime,
				Age:        candidateAge,
			},
		)
	}

	// The expensive similarity budget is applied only after deterministic,
	// source-independent eligibility and duplicate checks. Newer historical
	// evidence is evaluated first; the identifier is the stable tie-breaker.
	sort.SliceStable(
		pool.Candidates,
		func(left int, right int) bool {
			if !pool.Candidates[left].EndTime.Equal(
				pool.Candidates[right].EndTime,
			) {
				return pool.Candidates[left].EndTime.After(
					pool.Candidates[right].EndTime,
				)
			}
			return pool.Candidates[left].ID <
				pool.Candidates[right].ID
		},
	)

	if len(pool.Candidates) > config.MaximumCandidateCount {
		pool.Candidates = append(
			[]preparedCandidate(nil),
			pool.Candidates[:config.MaximumCandidateCount]...,
		)
		pool.Truncated = true
	}

	return pool
}
