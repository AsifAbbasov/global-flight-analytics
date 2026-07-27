package trajectorybuilder

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func calculatePathEfficiency(
	ctx context.Context,
	evidence canonicalEvidence,
) (ratioMetric, []flightfeatures.FeatureLimitation, error) {
	parts, limitations, usedPointEvidence, err := collectPathParts(ctx, evidence)
	if err != nil {
		return ratioMetric{}, nil, err
	}
	usablePartCount := 0
	for _, part := range parts {
		if len(part.coordinates) >= 2 {
			usablePartCount++
		}
	}
	if usablePartCount == 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code:    flightfeatures.TrajectoryLimitationPathEvidenceInsufficient,
			Message: "At least one continuous path part with two usable coordinates is required for path efficiency calculation.",
		})
		return ratioMetric{}, limitations, nil
	}

	observedDistance := kahanAccumulator{}
	directDistance := kahanAccumulator{}
	for partIndex, part := range parts {
		if partIndex%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return ratioMetric{}, nil, err
			}
		}
		if len(part.coordinates) < 2 {
			continue
		}
		for index := 1; index < len(part.coordinates); index++ {
			observedDistance.Add(haversineDistanceKM(part.coordinates[index-1], part.coordinates[index]))
		}
		directDistance.Add(haversineDistanceKM(part.coordinates[0], part.coordinates[len(part.coordinates)-1]))
	}
	observed := observedDistance.Value()
	direct := directDistance.Value()
	if observed <= 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code:    flightfeatures.TrajectoryLimitationPathZeroDistance,
			Message: "Observed continuous path distance is zero, so path efficiency ratio is undefined.",
		})
		return ratioMetric{}, limitations, nil
	}
	ratio := direct / observed
	if math.IsNaN(ratio) || math.IsInf(ratio, 0) {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code:    flightfeatures.TrajectoryLimitationPathAggregateNonFinite,
			Message: "Path efficiency aggregation produced a non-finite ratio.",
		})
		return ratioMetric{}, limitations, nil
	}
	if ratio < -pathRatioTolerance || ratio > 1+pathRatioTolerance {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationPathRatioOutOfRange,
			Message: fmt.Sprintf(
				"Path efficiency ratio %.12f is outside the inclusive zero-to-one range beyond numerical tolerance.",
				ratio,
			),
		})
		return ratioMetric{}, limitations, nil
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	if !usedPointEvidence {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code:    flightfeatures.TrajectoryLimitationPathSegmentFallback,
			Message: "Path efficiency uses independent usable segment endpoint parts because fewer than two continuous point coordinates were available; distances between segments are never treated as observed movement.",
		})
	}
	return ratioMetric{available: true, value: ratio}, limitations, nil
}

func collectPathParts(
	ctx context.Context,
	evidence canonicalEvidence,
) ([]pathPart, []flightfeatures.FeatureLimitation, bool, error) {
	pointParts, pointLimitations, err := pointPathParts(ctx, evidence.points, evidence.gaps)
	if err != nil {
		return nil, nil, false, err
	}
	for _, part := range pointParts {
		if len(part.coordinates) >= 2 {
			return pointParts, pointLimitations, true, nil
		}
	}

	segmentParts, segmentLimitations, err := segmentPathParts(ctx, evidence)
	if err != nil {
		return nil, nil, false, err
	}
	limitations := append(pointLimitations, segmentLimitations...)
	return segmentParts, limitations, false, nil
}

func pointPathParts(
	ctx context.Context,
	points []canonicalPoint,
	gaps []trajectory.CoverageGap,
) ([]pathPart, []flightfeatures.FeatureLimitation, error) {
	parts := make([]pathPart, 0)
	current := pathPart{}
	discontinuityCount := 0
	var previousTime time.Time

	flush := func() {
		if len(current.coordinates) > 0 {
			parts = append(parts, current)
			current = pathPart{}
		}
	}
	for index, point := range points {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return nil, nil, err
			}
		}
		if !point.coordinateAvailable {
			flush()
			previousTime = time.Time{}
			continue
		}
		if !previousTime.IsZero() && gapSeparatesInterval(previousTime, point.observedAt, gaps) {
			discontinuityCount++
			flush()
		}
		current.coordinates = append(current.coordinates, point.coordinate)
		previousTime = point.observedAt
	}
	flush()

	limitations := make([]flightfeatures.FeatureLimitation, 0, 2)
	if len(points) == 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code:    flightfeatures.TrajectoryLimitationPathPointEvidenceUnavailable,
			Message: "No canonical point records were available for path efficiency calculation.",
		})
	}
	if discontinuityCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationPathDiscontinuityExcluded,
			Message: fmt.Sprintf(
				"%d coverage-gap discontinuities split the point path; distances across those gaps were excluded from observed path efficiency.",
				discontinuityCount,
			),
		})
	}
	return parts, limitations, nil
}

func segmentPathParts(
	ctx context.Context,
	evidence canonicalEvidence,
) ([]pathPart, []flightfeatures.FeatureLimitation, error) {
	if !evidence.windowAvailable {
		return nil, nil, nil
	}
	parts := make([]pathPart, 0, len(evidence.segments))
	invalidStatusCount := 0
	invalidCoordinateCount := 0
	outsideWindowCount := 0
	missingTimestampCount := 0
	discontinuityCount := 0
	current := pathPart{}
	var previousEnd coordinate
	var previousEndTime time.Time
	hasPrevious := false

	flush := func() {
		if len(current.coordinates) > 0 {
			parts = append(parts, current)
			current = pathPart{}
		}
		hasPrevious = false
		previousEndTime = time.Time{}
	}

	for index, segment := range evidence.segments {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return nil, nil, err
			}
		}
		switch segment.Status {
		case trajectory.SegmentStatusObserved,
			trajectory.SegmentStatusInterpolated,
			trajectory.SegmentStatusEstimated:
		default:
			invalidStatusCount++
			flush()
			continue
		}

		segmentStartTime := segment.StartTime.UTC()
		segmentEndTime := segment.EndTime.UTC()
		if evidence.windowAvailable {
			if segment.StartTime.IsZero() || segment.EndTime.IsZero() {
				missingTimestampCount++
				flush()
				continue
			}
			if segmentEndTime.Before(segmentStartTime) ||
				segmentStartTime.Before(evidence.windowStart) ||
				segmentEndTime.After(evidence.windowEnd) {
				outsideWindowCount++
				flush()
				continue
			}
		}

		startCoordinate, startValid := normalizeCoordinate(
			segment.StartLatitude,
			segment.StartLongitude,
		)
		endCoordinate, endValid := normalizeCoordinate(
			segment.EndLatitude,
			segment.EndLongitude,
		)
		if !startValid || !endValid {
			invalidCoordinateCount++
			flush()
			continue
		}

		continuous := hasPrevious &&
			previousEnd.equal(startCoordinate) &&
			!gapSeparatesInterval(previousEndTime, segmentStartTime, evidence.gaps)
		if !continuous {
			if hasPrevious {
				discontinuityCount++
			}
			flush()
			current.coordinates = append(current.coordinates, startCoordinate, endCoordinate)
		} else if !current.coordinates[len(current.coordinates)-1].equal(endCoordinate) {
			current.coordinates = append(current.coordinates, endCoordinate)
		}
		previousEnd = endCoordinate
		previousEndTime = segmentEndTime
		hasPrevious = true
	}
	flush()

	limitations := make([]flightfeatures.FeatureLimitation, 0, 5)
	if invalidStatusCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationPathInvalidSegmentStatus,
			Message: fmt.Sprintf(
				"%d invalid or unknown trajectory segments were excluded from path fallback evidence.",
				invalidStatusCount,
			),
		})
	}
	if invalidCoordinateCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationPathInvalidSegmentCoordinates,
			Message: fmt.Sprintf(
				"%d trajectory segments have invalid endpoint coordinates and were excluded from path fallback evidence.",
				invalidCoordinateCount,
			),
		})
	}
	if missingTimestampCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationPathSegmentTimestampMissing,
			Message: fmt.Sprintf(
				"%d trajectory segments have missing timestamps and were excluded from path fallback evidence.",
				missingTimestampCount,
			),
		})
	}
	if outsideWindowCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationPathSegmentOutsideWindow,
			Message: fmt.Sprintf(
				"%d trajectory segments are reversed or outside the authoritative window and were excluded from path fallback evidence.",
				outsideWindowCount,
			),
		})
	}
	if discontinuityCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationPathDiscontinuityExcluded,
			Message: fmt.Sprintf(
				"%d segment discontinuities split fallback path evidence; no distance between disconnected segments was treated as observed movement.",
				discontinuityCount,
			),
		})
	}
	return parts, limitations, nil
}

func gapSeparatesInterval(
	left time.Time,
	right time.Time,
	gaps []trajectory.CoverageGap,
) bool {
	if !right.After(left) {
		return false
	}
	for _, gap := range gaps {
		if gap.StartTime.IsZero() || gap.EndTime.IsZero() || !gap.EndTime.After(gap.StartTime) {
			continue
		}
		if gap.EndTime.UTC().After(left) && gap.StartTime.UTC().Before(right) {
			return true
		}
	}
	return false
}

func normalizeCoordinate(latitude float64, longitude float64) (coordinate, bool) {
	if math.IsNaN(latitude) || math.IsInf(latitude, 0) ||
		math.IsNaN(longitude) || math.IsInf(longitude, 0) ||
		latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return coordinate{}, false
	}
	return coordinate{latitude: latitude, longitude: normalizeLongitude(longitude)}, true
}

func haversineDistanceKM(left coordinate, right coordinate) float64 {
	leftLatitude := degreesToRadians(left.latitude)
	rightLatitude := degreesToRadians(right.latitude)
	latitudeDifference := rightLatitude - leftLatitude
	longitudeDifference := degreesToRadians(normalizeLongitude(right.longitude - left.longitude))
	sineLatitude := math.Sin(latitudeDifference / 2)
	sineLongitude := math.Sin(longitudeDifference / 2)
	value := sineLatitude*sineLatitude +
		math.Cos(leftLatitude)*math.Cos(rightLatitude)*sineLongitude*sineLongitude
	value = math.Min(1, math.Max(0, value))
	return earthMeanRadiusKM * 2 * math.Atan2(math.Sqrt(value), math.Sqrt(1-value))
}

func normalizeLongitude(value float64) float64 {
	normalized := math.Mod(value+180, 360)
	if normalized < 0 {
		normalized += 360
	}
	normalized -= 180
	if normalized == 0 {
		return 0
	}
	return normalized
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}
