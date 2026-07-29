package projectionneighbors

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

var (
	ErrSelectorUnavailable = errors.New(
		"historical neighbor selector is unavailable",
	)
	ErrSimilarityEngineFailed = errors.New(
		"historical similarity engine failed",
	)
	ErrSimilarityEvidenceInvalid = errors.New(
		"historical similarity engine returned invalid consumer evidence",
	)
	ErrCurrentTrajectoryIDRequired = errors.New(
		"current trajectory identifier is required",
	)
	ErrAsOfTimeRequired = errors.New(
		"neighbor selection as-of time is required",
	)
	ErrContinuationDurationInvalid = errors.New(
		"required continuation duration must be greater than zero",
	)
	ErrRouteScopeInvalid = errors.New(
		"historical candidate route scope is invalid",
	)
	ErrCurrentTrajectoryNotComparable = errors.New(
		"current trajectory does not contain enough usable as-of points",
	)
	ErrSelectionResultInvalid = errors.New(
		"historical neighbor selection result is invalid",
	)
)

type Selector struct {
	config Config
}

func New(
	config Config,
) (*Selector, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate historical neighbor selector config: %w",
			err,
		)
	}

	return &Selector{
		config: config,
	}, nil
}

type Request struct {
	CurrentTrajectory trajectory.FlightTrajectory
	Candidates        []trajectory.FlightTrajectory

	RouteScope                   RouteScope
	AsOfTime                     time.Time
	RequiredContinuationDuration time.Duration
}

func (
	selector *Selector,
) Select(
	request Request,
) (Result, error) {
	if selector == nil {
		return Result{},
			ErrSelectorUnavailable
	}

	currentID := strings.TrimSpace(
		request.CurrentTrajectory.ID,
	)
	if currentID == "" {
		return Result{},
			ErrCurrentTrajectoryIDRequired
	}
	if request.AsOfTime.IsZero() {
		return Result{},
			ErrAsOfTimeRequired
	}
	if request.RequiredContinuationDuration <= 0 {
		return Result{},
			ErrContinuationDurationInvalid
	}

	routeScope, err := prepareRouteScope(
		request.RouteScope,
		request.Candidates,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrRouteScopeInvalid,
			err,
		)
	}

	asOfTime := request.AsOfTime.UTC()
	current, excludedCurrentPointCount :=
		snapshotAt(
			request.CurrentTrajectory,
			asOfTime,
		)
	if len(current.Points) <
		selector.config.
			MinimumCurrentPointCount {
		return Result{},
			fmt.Errorf(
				"%w: usable=%d minimum=%d",
				ErrCurrentTrajectoryNotComparable,
				len(current.Points),
				selector.config.
					MinimumCurrentPointCount,
			)
	}

	latestCurrentPoint := current.Points[len(current.Points)-1]
	currentStartTime := current.Points[0].ObservedAt.UTC()

	pool := prepareCandidatePool(
		request.Candidates,
		currentID,
		currentStartTime,
		asOfTime,
		selector.config,
		routeScope,
	)
	candidates := pool.Candidates
	truncated := pool.Truncated

	neighbors := make(
		[]Neighbor,
		0,
		selector.config.SelectionLimit,
	)
	rejections := append(
		[]Rejection(nil),
		pool.Rejections...,
	)
	excludedCandidateFuturePointCount :=
		pool.ExcludedFuturePointCount

	for _, prepared := range candidates {
		candidateID := prepared.ID
		candidate := prepared.Trajectory
		candidateStartTime := prepared.StartTime
		candidateEndTime := prepared.EndTime
		candidateAge := prepared.Age

		anchorSearch := findAnchor(
			latestCurrentPoint,
			candidate.Points,
			selector.config.MinimumCurrentPointCount,
			request.RequiredContinuationDuration,
			selector.config.effectiveMaximumContinuationGap(),
		)
		if !anchorSearch.Found() {
			code := RejectionContinuationUnavailable
			message := "Historical candidate does not provide enough observed continuation after a comparable prefix."
			if anchorSearch.Failure == anchorSearchFailureDiscontinuous {
				code = RejectionContinuationDiscontinuous
				message = "Historical candidate continuation crosses an observation gap larger than the configured maximum."
			}
			rejections = append(
				rejections,
				rejection(candidateID, code, message),
			)
			continue
		}
		anchor := anchorSearch.Evidence
		if anchor.DistanceKM >
			selector.config.
				MaximumAnchorDistanceKM {
			rejections = append(
				rejections,
				rejection(
					candidateID,
					RejectionAnchorTooDistant,
					"Historical candidate anchor exceeds the configured maximum distance from the current endpoint.",
				),
			)
			continue
		}

		prefix := candidatePrefix(
			candidate,
			anchor.PointIndex,
		)
		similarityResult, err := selector.config.
			SimilarityEngine.Compare(
			current,
			prefix,
		)
		if err != nil {
			if candidateSimilarityFailure(err) {
				rejections = append(
					rejections,
					rejection(
						candidateID,
						RejectionSimilarityUnavailable,
						"Historical candidate prefix could not be compared with the current trajectory.",
					),
				)
				continue
			}
			return Result{}, fmt.Errorf(
				"%w: candidate=%s: %v",
				ErrSimilarityEngineFailed,
				candidateID,
				err,
			)
		}
		similarity, err := similarityEvidenceFromResult(
			similarityResult,
			current,
			prefix,
		)
		if err != nil {
			return Result{}, fmt.Errorf(
				"%w: candidate=%s: %v",
				ErrSimilarityEvidenceInvalid,
				candidateID,
				err,
			)
		}
		if similarity.Score <
			selector.config.
				MinimumSimilarityScore {
			rejections = append(
				rejections,
				rejection(
					candidateID,
					RejectionSimilarityBelowMinimum,
					"Historical candidate similarity is below the configured minimum.",
				),
			)
			continue
		}

		neighbors = append(
			neighbors,
			Neighbor{
				TrajectoryID: candidateID,

				SimilarityScore: similarity.Score,
				SimilarityLevel: similarity.Level,
				SimilarityInputFingerprint: similarity.
					InputFingerprint,

				AnchorPointIndex: anchor.PointIndex,
				AnchorObservedAt: anchor.ObservedAt,
				AnchorDistanceKM: anchor.DistanceKM,

				CandidateStartTime: candidateStartTime,
				CandidateEndTime:   candidateEndTime,
				CandidateAge:       candidateAge,

				PrefixPointCount:       anchor.PointIndex + 1,
				ContinuationPointCount: anchor.ContinuationPointCount,
				ContinuationEndTime:    anchor.ContinuationEndTime,
			},
		)
	}

	sort.SliceStable(
		neighbors,
		func(left int, right int) bool {
			if neighbors[left].
				SimilarityScore !=
				neighbors[right].
					SimilarityScore {
				return neighbors[left].
					SimilarityScore >
					neighbors[right].
						SimilarityScore
			}
			if neighbors[left].
				AnchorDistanceKM !=
				neighbors[right].
					AnchorDistanceKM {
				return neighbors[left].
					AnchorDistanceKM <
					neighbors[right].
						AnchorDistanceKM
			}

			return neighbors[left].
				TrajectoryID <
				neighbors[right].
					TrajectoryID
		},
	)

	qualifiedCandidateCount :=
		len(neighbors)
	if len(neighbors) >
		selector.config.SelectionLimit {
		neighbors = neighbors[:selector.config.SelectionLimit]
	}

	limitations := make(
		[]Notice,
		0,
		3,
	)
	if excludedCurrentPointCount > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "future_current_points_excluded",
				Message: fmt.Sprintf(
					"%d current-trajectory points after the as-of time were excluded.",
					excludedCurrentPointCount,
				),
			},
		)
	}
	if excludedCandidateFuturePointCount > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "future_candidate_points_excluded",
				Message: fmt.Sprintf(
					"%d historical-candidate points after the as-of time were excluded.",
					excludedCandidateFuturePointCount,
				),
			},
		)
	}
	if qualifiedCandidateCount >
		len(neighbors) {
		limitations = append(
			limitations,
			Notice{
				Code: "qualified_neighbors_limited",
				Message: fmt.Sprintf(
					"%d qualified historical neighbors were reduced to the configured selection limit of %d.",
					qualifiedCandidateCount,
					selector.config.SelectionLimit,
				),
			},
		)
	}
	if truncated {
		limitations = append(
			limitations,
			Notice{
				Code:    "candidate_evaluation_truncated",
				Message: "Historical candidate evaluation was truncated at the configured maximum candidate count.",
			},
		)
	}
	if len(rejections) > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "historical_candidates_rejected",
				Message: fmt.Sprintf(
					"%d historical candidates were rejected by deterministic selection guards.",
					len(rejections),
				),
			},
		)
	}

	status := StatusUnavailable
	switch {
	case len(neighbors) == 0:
		limitations = append(
			limitations,
			Notice{
				Code:    "historical_neighbor_unavailable",
				Message: "No historical trajectory satisfied the configured selection policy.",
			},
		)
	case len(neighbors) ==
		selector.config.SelectionLimit &&
		!truncated:
		status = StatusComplete
	default:
		status = StatusPartial
	}

	result := Result{
		Version: Version,
		Status:  status,

		CurrentTrajectoryID: currentID,
		AsOfTime:            asOfTime,
		RequiredContinuationDuration: request.
			RequiredContinuationDuration,

		InputCandidateCount: len(request.Candidates),
		CheckedCandidateCount: qualifiedCandidateCount +
			len(rejections),
		QualifiedCandidateCount: qualifiedCandidateCount,
		RejectedCandidateCount:  len(rejections),

		SelectionLimit: selector.config.SelectionLimit,
		Truncated:      truncated,

		Neighbors: append(
			[]Neighbor(nil),
			neighbors...,
		),
		Rejections: append(
			[]Rejection(nil),
			rejections...,
		),
		Limitations: normalizeNotices(
			limitations,
		),

		InputFingerprint: selectionFingerprint(
			current,
			request.Candidates,
			asOfTime,
			request.
				RequiredContinuationDuration,
			selector.config,
			routeScope.scope,
		),
	}

	if err := result.Validate(); err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %v",
				ErrSelectionResultInvalid,
				err,
			)
	}

	return result.Clone(), nil
}

func snapshotAt(
	item trajectory.FlightTrajectory,
	asOfTime time.Time,
) (trajectory.FlightTrajectory, int) {
	points := make(
		[]trajectory.TrackPoint4D,
		0,
		len(item.Points),
	)
	excludedFutureCount := 0

	for _, point := range item.Points {
		if point.ObservedAt.IsZero() ||
			!validLatitude(point.Latitude) ||
			!validLongitude(point.Longitude) {
			continue
		}
		if point.ObservedAt.UTC().After(asOfTime) {
			excludedFutureCount++
			continue
		}

		point.ObservedAt = point.ObservedAt.UTC()
		points = append(points, point)
	}

	sort.SliceStable(
		points,
		func(left int, right int) bool {
			return canonicalPointLess(
				points[left],
				points[right],
			)
		},
	)

	snapshot := item
	snapshot.Points = points
	snapshot.PointCount = len(points)

	if len(points) == 0 {
		snapshot.StartTime = time.Time{}
		snapshot.EndTime = time.Time{}
		snapshot.DurationSeconds = 0
		return snapshot, excludedFutureCount
	}

	snapshot.StartTime = points[0].ObservedAt
	snapshot.EndTime = points[len(points)-1].ObservedAt
	snapshot.DurationSeconds = int64(
		snapshot.EndTime.Sub(snapshot.StartTime).Seconds(),
	)
	if snapshot.UpdatedAt.After(asOfTime) {
		snapshot.UpdatedAt = asOfTime
	}

	return snapshot, excludedFutureCount
}

func candidatePrefix(
	candidate trajectory.FlightTrajectory,
	anchorIndex int,
) trajectory.FlightTrajectory {
	prefix := candidate
	prefix.ID = candidate.ID +
		"#projection-prefix"
	prefix.Points = append(
		[]trajectory.TrackPoint4D(nil),
		candidate.Points[:anchorIndex+1]...,
	)
	prefix.PointCount =
		len(prefix.Points)
	prefix.StartTime =
		prefix.Points[0].
			ObservedAt.UTC()
	prefix.EndTime =
		prefix.Points[len(prefix.Points)-1].ObservedAt.UTC()
	prefix.DurationSeconds = int64(
		prefix.EndTime.Sub(
			prefix.StartTime,
		).Seconds(),
	)

	return prefix
}

func rejection(
	trajectoryID string,
	code RejectionCode,
	message string,
) Rejection {
	return Rejection{
		TrajectoryID: strings.TrimSpace(
			trajectoryID,
		),
		Code: code,
		Message: strings.TrimSpace(
			message,
		),
	}
}
