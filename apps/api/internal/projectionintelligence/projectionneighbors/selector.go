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

	context, err := selector.prepareSelectionContext(
		request,
	)
	if err != nil {
		return Result{}, err
	}

	evaluation, err := selector.evaluateCandidatePool(
		context,
	)
	if err != nil {
		return Result{}, err
	}

	return selector.assembleSelectionResult(
		context,
		evaluation,
	)
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
