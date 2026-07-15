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
	ErrCurrentTrajectoryIDRequired = errors.New(
		"current trajectory identifier is required",
	)
	ErrAsOfTimeRequired = errors.New(
		"neighbor selection as-of time is required",
	)
	ErrContinuationDurationInvalid = errors.New(
		"required continuation duration must be greater than zero",
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
			ErrSimilarityEngineRequired
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

	candidates := append(
		[]trajectory.FlightTrajectory(nil),
		request.Candidates...,
	)
	sort.SliceStable(
		candidates,
		func(left int, right int) bool {
			leftID := strings.TrimSpace(
				candidates[left].ID,
			)
			rightID := strings.TrimSpace(
				candidates[right].ID,
			)
			return leftID < rightID
		},
	)

	truncated := false
	if len(candidates) >
		selector.config.
			MaximumCandidateCount {
		candidates = candidates[:selector.config.MaximumCandidateCount]
		truncated = true
	}

	latestCurrentPoint := current.Points[len(current.Points)-1]
	currentStartTime :=
		current.Points[0].
			ObservedAt.UTC()

	neighbors := make(
		[]Neighbor,
		0,
		selector.config.SelectionLimit,
	)
	rejections := make(
		[]Rejection,
		0,
		len(candidates),
	)
	excludedCandidateFuturePointCount := 0

	candidateIDCounts := make(
		map[string]int,
		len(candidates),
	)
	for _, candidate := range candidates {
		candidateID := strings.TrimSpace(
			candidate.ID,
		)
		candidateIDCounts[candidateID]++
	}

	for _, candidateInput := range candidates {
		candidateID := strings.TrimSpace(
			candidateInput.ID,
		)
		if candidateID == "" {
			rejections = append(
				rejections,
				rejection(
					candidateID,
					RejectionIdentifierMissing,
					"Historical candidate identifier is required.",
				),
			)
			continue
		}
		if candidateID == currentID {
			rejections = append(
				rejections,
				rejection(
					candidateID,
					RejectionSameTrajectory,
					"Current trajectory cannot be selected as its own historical neighbor.",
				),
			)
			continue
		}
		if candidateIDCounts[candidateID] > 1 {
			rejections = append(
				rejections,
				rejection(
					candidateID,
					RejectionDuplicateCandidate,
					"Every historical candidate with a duplicated identifier was rejected.",
				),
			)
			continue
		}

		candidate,
			excludedFuturePointCount :=
			snapshotAt(
				candidateInput,
				asOfTime,
			)
		excludedCandidateFuturePointCount +=
			excludedFuturePointCount
		if len(candidate.Points) <
			selector.config.
				MinimumCurrentPointCount+1 {
			rejections = append(
				rejections,
				rejection(
					candidateID,
					RejectionInsufficientPoints,
					"Historical candidate does not contain enough usable points for a comparable prefix and continuation.",
				),
			)
			continue
		}

		candidateStartTime :=
			candidate.Points[0].
				ObservedAt.UTC()
		candidateEndTime :=
			candidate.Points[len(candidate.Points)-1].ObservedAt.UTC()

		if !candidateEndTime.Before(
			currentStartTime,
		) {
			rejections = append(
				rejections,
				rejection(
					candidateID,
					RejectionNotHistorical,
					"Historical candidate must end before the current trajectory begins.",
				),
			)
			continue
		}

		candidateAge := asOfTime.Sub(
			candidateEndTime,
		)
		if candidateAge < 0 {
			rejections = append(
				rejections,
				rejection(
					candidateID,
					RejectionNotHistorical,
					"Historical candidate contains evidence after the as-of time.",
				),
			)
			continue
		}
		if selector.config.
			MaximumCandidateAge > 0 &&
			candidateAge >
				selector.config.
					MaximumCandidateAge {
			rejections = append(
				rejections,
				rejection(
					candidateID,
					RejectionTooOld,
					"Historical candidate exceeds the configured maximum age.",
				),
			)
			continue
		}

		anchorIndex, anchorDistanceKM,
			continuationPointCount,
			continuationEndTime,
			found := findAnchor(
			latestCurrentPoint,
			candidate.Points,
			selector.config.
				MinimumCurrentPointCount,
			request.
				RequiredContinuationDuration,
		)
		if !found {
			rejections = append(
				rejections,
				rejection(
					candidateID,
					RejectionContinuationUnavailable,
					"Historical candidate does not provide enough observed continuation after a comparable prefix.",
				),
			)
			continue
		}
		if anchorDistanceKM >
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
			anchorIndex,
		)
		similarity, err := selector.config.
			SimilarityEngine.Compare(
			current,
			prefix,
		)
		if err != nil {
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

				AnchorPointIndex: anchorIndex,
				AnchorObservedAt: candidate.Points[anchorIndex].ObservedAt.UTC(),
				AnchorDistanceKM: anchorDistanceKM,

				CandidateStartTime: candidateStartTime,
				CandidateEndTime:   candidateEndTime,
				CandidateAge:       candidateAge,

				PrefixPointCount:       anchorIndex + 1,
				ContinuationPointCount: continuationPointCount,
				ContinuationEndTime:    continuationEndTime,
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

		InputCandidateCount:     len(request.Candidates),
		CheckedCandidateCount:   len(candidates),
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
	type indexedPoint struct {
		point trajectory.TrackPoint4D
		index int
	}

	valid := make(
		[]indexedPoint,
		0,
		len(item.Points),
	)
	excludedFutureCount := 0

	for index, point := range item.Points {
		if point.ObservedAt.IsZero() ||
			!validLatitude(
				point.Latitude,
			) ||
			!validLongitude(
				point.Longitude,
			) {
			continue
		}
		if point.ObservedAt.UTC().After(
			asOfTime,
		) {
			excludedFutureCount++
			continue
		}

		point.ObservedAt =
			point.ObservedAt.UTC()
		valid = append(
			valid,
			indexedPoint{
				point: point,
				index: index,
			},
		)
	}

	sort.SliceStable(
		valid,
		func(left int, right int) bool {
			if valid[left].point.
				ObservedAt.Equal(
				valid[right].point.
					ObservedAt,
			) {
				return valid[left].index <
					valid[right].index
			}

			return valid[left].point.
				ObservedAt.Before(
				valid[right].point.
					ObservedAt,
			)
		},
	)

	points := make(
		[]trajectory.TrackPoint4D,
		len(valid),
	)
	for index, item := range valid {
		points[index] = item.point
	}

	snapshot := item
	snapshot.Points = points
	snapshot.PointCount = len(points)

	if len(points) == 0 {
		snapshot.StartTime = time.Time{}
		snapshot.EndTime = time.Time{}
		snapshot.DurationSeconds = 0
		return snapshot,
			excludedFutureCount
	}

	snapshot.StartTime =
		points[0].ObservedAt
	snapshot.EndTime =
		points[len(points)-1].
			ObservedAt
	snapshot.DurationSeconds = int64(
		snapshot.EndTime.Sub(
			snapshot.StartTime,
		).Seconds(),
	)
	if snapshot.UpdatedAt.After(
		asOfTime,
	) {
		snapshot.UpdatedAt = asOfTime
	}

	return snapshot,
		excludedFutureCount
}

func findAnchor(
	currentEndpoint trajectory.TrackPoint4D,
	candidatePoints []trajectory.TrackPoint4D,
	minimumPrefixPointCount int,
	requiredContinuationDuration time.Duration,
) (
	int,
	float64,
	int,
	time.Time,
	bool,
) {
	if len(candidatePoints) <
		minimumPrefixPointCount+1 {
		return 0, 0, 0,
			time.Time{}, false
	}

	currentGeoPoint := geoPoint{
		latitude:  currentEndpoint.Latitude,
		longitude: currentEndpoint.Longitude,
	}

	bestIndex := -1
	bestDistance := 0.0
	bestContinuationPointCount := 0
	bestContinuationEndTime :=
		time.Time{}

	for index :=
		minimumPrefixPointCount - 1; index < len(candidatePoints)-1; index++ {
		anchorTime := candidatePoints[index].ObservedAt.UTC()
		requiredEndTime := anchorTime.Add(
			requiredContinuationDuration,
		)

		continuationEndIndex := -1
		for futureIndex :=
			index + 1; futureIndex <
			len(candidatePoints); futureIndex++ {
			if !candidatePoints[futureIndex].ObservedAt.UTC().Before(
				requiredEndTime,
			) {
				continuationEndIndex =
					futureIndex
				break
			}
		}
		if continuationEndIndex < 0 {
			continue
		}

		distance := haversineKM(
			currentGeoPoint,
			geoPoint{
				latitude:  candidatePoints[index].Latitude,
				longitude: candidatePoints[index].Longitude,
			},
		)
		if bestIndex < 0 ||
			distance < bestDistance ||
			(distance == bestDistance &&
				anchorTime.Before(
					candidatePoints[bestIndex].ObservedAt.UTC(),
				)) {
			bestIndex = index
			bestDistance = distance
			bestContinuationPointCount =
				continuationEndIndex - index
			bestContinuationEndTime =
				candidatePoints[continuationEndIndex].ObservedAt.UTC()
		}
	}

	if bestIndex < 0 {
		return 0, 0, 0,
			time.Time{}, false
	}

	return bestIndex,
		bestDistance,
		bestContinuationPointCount,
		bestContinuationEndTime,
		true
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
