package trajectorybuilder

import (
	"context"
	"math"
	"sort"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func calculatePathEfficiency(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (
	ratioMetric,
	[]flightfeatures.FeatureLimitation,
) {
	coordinates, limitations, err :=
		collectPathCoordinates(ctx, item)
	if err != nil {
		return ratioMetric{}, nil
	}
	if len(coordinates) < 2 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "trajectory_path_efficiency_evidence_insufficient",
				Message: "At least two usable ordered coordinates are required for path efficiency calculation.",
			},
		)

		return ratioMetric{}, limitations
	}

	observedPathDistance := 0.0
	for index := 1; index < len(coordinates); index++ {
		observedPathDistance += haversineDistanceKM(
			coordinates[index-1],
			coordinates[index],
		)
	}
	if observedPathDistance <= 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "trajectory_path_efficiency_zero_path",
				Message: "Observed path distance is zero, so path efficiency ratio is undefined.",
			},
		)

		return ratioMetric{}, limitations
	}

	directDistance := haversineDistanceKM(
		coordinates[0],
		coordinates[len(coordinates)-1],
	)
	ratio := directDistance / observedPathDistance
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	return ratioMetric{
		available: true,
		value:     ratio,
	}, limitations
}

func collectPathCoordinates(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (
	[]coordinate,
	[]flightfeatures.FeatureLimitation,
	error,
) {
	pointCoordinates := make(
		[]coordinate,
		0,
		len(item.Points),
	)
	invalidPointCount := 0

	for index, point := range item.Points {
		if index%1024 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, nil, err
			}
		}

		value, valid := normalizeCoordinate(
			point.Latitude,
			point.Longitude,
		)
		if !valid {
			invalidPointCount++
			continue
		}
		pointCoordinates = append(
			pointCoordinates,
			value,
		)
	}

	if len(pointCoordinates) > 0 {
		limitations := make(
			[]flightfeatures.FeatureLimitation,
			0,
			1,
		)
		if invalidPointCount > 0 {
			limitations = append(
				limitations,
				flightfeatures.FeatureLimitation{
					Code:    "trajectory_path_invalid_point_coordinates",
					Message: "One or more trajectory points contain non-finite or out-of-range coordinates and were excluded from path efficiency calculation.",
				},
			)
		}

		return pointCoordinates,
			limitations,
			nil
	}

	limitations := make(
		[]flightfeatures.FeatureLimitation,
		0,
		4,
	)
	if len(item.Points) == 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "trajectory_path_point_evidence_unavailable",
				Message: "No trajectory points were available for path efficiency calculation.",
			},
		)
	} else {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "trajectory_path_point_evidence_unusable",
				Message: "No trajectory point contained a usable geographic coordinate.",
			},
		)
	}

	segments := append(
		[]trajectory.TrajectorySegment(nil),
		item.Segments...,
	)
	sort.SliceStable(
		segments,
		func(left int, right int) bool {
			if segments[left].SequenceNumber !=
				segments[right].SequenceNumber {
				return segments[left].SequenceNumber <
					segments[right].SequenceNumber
			}
			if !segments[left].StartTime.Equal(
				segments[right].StartTime,
			) {
				return segments[left].StartTime.Before(
					segments[right].StartTime,
				)
			}

			return segments[left].ID <
				segments[right].ID
		},
	)

	segmentCoordinates := make(
		[]coordinate,
		0,
		len(segments)*2,
	)
	invalidSegmentCoordinateCount := 0

	appendIfUsable := func(
		latitude float64,
		longitude float64,
	) {
		value, valid := normalizeCoordinate(
			latitude,
			longitude,
		)
		if !valid {
			invalidSegmentCoordinateCount++
			return
		}
		if len(segmentCoordinates) == 0 ||
			!segmentCoordinates[len(segmentCoordinates)-1].equal(value) {
			segmentCoordinates = append(
				segmentCoordinates,
				value,
			)
		}
	}

	for index, segment := range segments {
		if index%1024 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, nil, err
			}
		}
		if segment.Status ==
			trajectory.SegmentStatusInvalid {
			continue
		}

		appendIfUsable(
			segment.StartLatitude,
			segment.StartLongitude,
		)
		appendIfUsable(
			segment.EndLatitude,
			segment.EndLongitude,
		)
	}

	if len(segmentCoordinates) > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "trajectory_path_segment_endpoint_fallback",
				Message: "Path efficiency was approximated from ordered non-invalid trajectory segment endpoints because no usable point coordinate was available.",
			},
		)
		if invalidSegmentCoordinateCount > 0 {
			limitations = append(
				limitations,
				flightfeatures.FeatureLimitation{
					Code:    "trajectory_path_invalid_segment_coordinates",
					Message: "One or more trajectory segment endpoints were invalid and were excluded from path efficiency fallback evidence.",
				},
			)
		}

		return segmentCoordinates,
			limitations,
			nil
	}

	limitations = append(
		limitations,
		flightfeatures.FeatureLimitation{
			Code:    "trajectory_path_coordinates_unavailable",
			Message: "No usable ordered coordinates were available from trajectory points or non-invalid segment endpoints.",
		},
	)

	return nil, limitations, nil
}

func normalizeCoordinate(
	latitude float64,
	longitude float64,
) (coordinate, bool) {
	if math.IsNaN(latitude) ||
		math.IsInf(latitude, 0) ||
		math.IsNaN(longitude) ||
		math.IsInf(longitude, 0) ||
		latitude < -90 ||
		latitude > 90 ||
		longitude < -180 ||
		longitude > 180 {
		return coordinate{}, false
	}

	return coordinate{
		latitude:  latitude,
		longitude: normalizeLongitude(longitude),
	}, true
}

func haversineDistanceKM(
	left coordinate,
	right coordinate,
) float64 {
	leftLatitude := degreesToRadians(left.latitude)
	rightLatitude := degreesToRadians(right.latitude)
	latitudeDifference :=
		rightLatitude - leftLatitude
	longitudeDifference := degreesToRadians(
		normalizeLongitude(
			right.longitude - left.longitude,
		),
	)

	sineLatitude :=
		math.Sin(latitudeDifference / 2)
	sineLongitude :=
		math.Sin(longitudeDifference / 2)
	value := sineLatitude*sineLatitude +
		math.Cos(leftLatitude)*
			math.Cos(rightLatitude)*
			sineLongitude*sineLongitude
	value = math.Min(1, math.Max(0, value))

	return earthMeanRadiusKM *
		2 *
		math.Atan2(
			math.Sqrt(value),
			math.Sqrt(1-value),
		)
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
