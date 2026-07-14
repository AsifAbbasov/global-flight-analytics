package routepipeline

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func usableSegments(
	items []trajectory.TrajectorySegment,
) []trajectory.TrajectorySegment {
	result := make(
		[]trajectory.TrajectorySegment,
		0,
		len(items),
	)

	for _, item := range items {
		if item.Status == trajectory.SegmentStatusInvalid {
			continue
		}
		if !validCoordinates(
			item.StartLatitude,
			item.StartLongitude,
		) || !validCoordinates(
			item.EndLatitude,
			item.EndLongitude,
		) {
			continue
		}
		if item.PointCount < 1 {
			continue
		}

		result = append(result, item)
	}

	sort.SliceStable(
		result,
		func(left int, right int) bool {
			if result[left].SequenceNumber ==
				result[right].SequenceNumber {
				if result[left].StartTime.Equal(
					result[right].StartTime,
				) {
					return result[left].ID <
						result[right].ID
				}

				return result[left].StartTime.Before(
					result[right].StartTime,
				)
			}

			return result[left].SequenceNumber <
				result[right].SequenceNumber
		},
	)

	return result
}

func validCoordinates(
	latitude float64,
	longitude float64,
) bool {
	return !math.IsNaN(latitude) &&
		!math.IsInf(latitude, 0) &&
		latitude >= -90 &&
		latitude <= 90 &&
		!math.IsNaN(longitude) &&
		!math.IsInf(longitude, 0) &&
		longitude >= -180 &&
		longitude <= 180
}

func analyticalAsOfTime(
	item trajectory.FlightTrajectory,
) time.Time {
	candidates := []time.Time{
		item.EndTime,
		item.CreatedAt,
		item.UpdatedAt,
	}

	for _, segment := range item.Segments {
		candidates = append(
			candidates,
			segment.EndTime,
			segment.CreatedAt,
		)
	}
	for _, gap := range item.CoverageGaps {
		candidates = append(
			candidates,
			gap.EndTime,
			gap.CreatedAt,
		)
	}

	var selected time.Time
	for _, candidate := range candidates {
		if candidate.IsZero() {
			continue
		}

		candidate = candidate.UTC()
		if selected.IsZero() ||
			candidate.After(selected) {
			selected = candidate
		}
	}

	return selected
}

func trajectoryUpdatedAt(
	item trajectory.FlightTrajectory,
	asOfTime time.Time,
) time.Time {
	for _, candidate := range []time.Time{
		item.UpdatedAt,
		item.CreatedAt,
		item.EndTime,
	} {
		if candidate.IsZero() {
			continue
		}

		normalized := candidate.UTC()
		if normalized.After(asOfTime) {
			continue
		}

		return normalized
	}

	return asOfTime
}

func sourceNames(
	item trajectory.FlightTrajectory,
	airportSourceName string,
) []string {
	unique := make(map[string]struct{})

	add := func(value string) {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			return
		}

		unique[normalized] = struct{}{}
	}

	add(airportSourceName)
	add(item.SourceName)

	for _, segment := range item.Segments {
		add(segment.SourceName)
	}
	for _, point := range item.Points {
		add(point.SourceName)
	}

	if len(unique) == 0 {
		add("trajectory")
	}

	result := make(
		[]string,
		0,
		len(unique),
	)
	for value := range unique {
		result = append(result, value)
	}
	sort.Strings(result)

	return result
}

func cloneTrajectory(
	item trajectory.FlightTrajectory,
) trajectory.FlightTrajectory {
	cloned := item
	cloned.Points = append(
		[]trajectory.TrackPoint4D(nil),
		item.Points...,
	)
	cloned.Segments = append(
		[]trajectory.TrajectorySegment(nil),
		item.Segments...,
	)
	cloned.CoverageGaps = append(
		[]trajectory.CoverageGap(nil),
		item.CoverageGaps...,
	)

	return cloned
}
