package geographicalbuilder

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

var _ extractor.GeographicalBuilder = (*Builder)(nil)

type Builder struct {
	geographicCellPrecision int
}

type Config struct {
	GeographicCellPrecision int
}

func New(config Config) (*Builder, error) {
	precision := config.GeographicCellPrecision
	if precision == 0 {
		precision = DefaultGeographicCellPrecision
	}
	if precision < 0 || precision > 6 {
		return nil, ErrInvalidGeographicCellPrecision
	}

	return &Builder{
		geographicCellPrecision: precision,
	}, nil
}

func (builder *Builder) Build(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (flightfeatures.GeographicalFeatures, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.GeographicalFeatures{}, err
	}

	coordinates, evidence := collectCoordinates(ctx, item)
	if err := ctx.Err(); err != nil {
		return flightfeatures.GeographicalFeatures{}, err
	}
	if len(coordinates) == 0 {
		return flightfeatures.GeographicalFeatures{
			Evidence: flightfeatures.GroupEvidence{
				Status:          flightfeatures.AvailabilityStatusUnavailable,
				TotalFieldCount: GeographicalFeatureFieldCount,
				Limitations: append(
					[]flightfeatures.FeatureLimitation(nil),
					evidence.limitations...,
				),
			},
			GeographicCellPrecision: builder.geographicCellPrecision,
		}, nil
	}

	start := coordinates[0]
	end := coordinates[len(coordinates)-1]
	minimumLatitude, maximumLatitude :=
		latitudeBounds(coordinates)
	minimumLongitude, maximumLongitude, longitudeSpan :=
		circularLongitudeBounds(coordinates)

	features := flightfeatures.GeographicalFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Status:               flightfeatures.AvailabilityStatusAvailable,
			AvailableFieldCount:  GeographicalFeatureFieldCount,
			TotalFieldCount:      GeographicalFeatureFieldCount,
			SupportingPointCount: len(coordinates),
			Limitations: append(
				[]flightfeatures.FeatureLimitation(nil),
				evidence.limitations...,
			),
		},
		StartLatitude:          start.latitude,
		StartLongitude:         start.longitude,
		EndLatitude:            end.latitude,
		EndLongitude:           end.longitude,
		MinimumLatitude:        minimumLatitude,
		MaximumLatitude:        maximumLatitude,
		MinimumLongitude:       minimumLongitude,
		MaximumLongitude:       maximumLongitude,
		LatitudeSpanDegrees:    maximumLatitude - minimumLatitude,
		LongitudeSpanDegrees:   longitudeSpan,
		GreatCircleDistanceKM:  haversineDistanceKM(start, end),
		ObservedPathDistanceKM: observedPathDistanceKM(coordinates),
		MaximumDisplacementKM:  maximumDisplacementKM(coordinates),
		CrossesAntimeridian:    pathCrossesAntimeridian(coordinates),
		UniqueGeographicCellCount: uniqueGeographicCellCount(
			coordinates,
			builder.geographicCellPrecision,
		),
		GeographicCellPrecision: builder.geographicCellPrecision,
	}

	if len(coordinates) == 1 {
		features.Evidence.Limitations = append(
			features.Evidence.Limitations,
			flightfeatures.FeatureLimitation{
				Code:    "geographical_single_coordinate",
				Message: "Only one usable coordinate supports the geographical feature group, so movement distances and spans are zero.",
			},
		)
	}

	if item.PointCount > 0 &&
		item.PointCount != len(item.Points) {
		features.Evidence.Limitations = append(
			features.Evidence.Limitations,
			flightfeatures.FeatureLimitation{
				Code: "trajectory_point_count_metadata_mismatch",
				Message: fmt.Sprintf(
					"Trajectory point-count metadata reports %d points while %d point records are present.",
					item.PointCount,
					len(item.Points),
				),
			},
		)
	}

	if err := ctx.Err(); err != nil {
		return flightfeatures.GeographicalFeatures{}, err
	}

	return cloneFeatures(features), nil
}

type coordinateEvidence struct {
	limitations []flightfeatures.FeatureLimitation
}

func collectCoordinates(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) ([]coordinate, coordinateEvidence) {
	pointCoordinates, invalidPointCount :=
		usablePointCoordinates(ctx, item.Points)
	if len(pointCoordinates) > 0 {
		limitations := make(
			[]flightfeatures.FeatureLimitation,
			0,
			2,
		)
		if invalidPointCount > 0 {
			limitations = append(
				limitations,
				flightfeatures.FeatureLimitation{
					Code:    "geographical_invalid_point_coordinates",
					Message: "One or more trajectory points contain non-finite or out-of-range coordinates and were excluded.",
				},
			)
		}

		return pointCoordinates, coordinateEvidence{
			limitations: limitations,
		}
	}

	segmentCoordinates, invalidSegmentCount :=
		usableSegmentCoordinates(ctx, item.Segments)
	limitations := make(
		[]flightfeatures.FeatureLimitation,
		0,
		4,
	)

	if len(item.Points) == 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "geographical_point_evidence_unavailable",
				Message: "No trajectory points were available for geographical feature extraction.",
			},
		)
	} else if invalidPointCount > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "geographical_point_evidence_unusable",
				Message: "No trajectory point contained a usable geographic coordinate.",
			},
		)
	}

	if len(segmentCoordinates) > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "geographical_segment_endpoint_fallback",
				Message: "Geographical features were approximated from ordered trajectory segment endpoints because no usable point coordinate was available.",
			},
		)
		if invalidSegmentCount > 0 {
			limitations = append(
				limitations,
				flightfeatures.FeatureLimitation{
					Code:    "geographical_invalid_segment_coordinates",
					Message: "One or more trajectory segment endpoints were invalid and were excluded from fallback evidence.",
				},
			)
		}

		return segmentCoordinates, coordinateEvidence{
			limitations: limitations,
		}
	}

	if invalidSegmentCount > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "geographical_segment_evidence_unusable",
				Message: "No non-invalid trajectory segment contained usable endpoint coordinates.",
			},
		)
	}
	limitations = append(
		limitations,
		flightfeatures.FeatureLimitation{
			Code:    "geographical_coordinates_unavailable",
			Message: "No usable geographic coordinate was available from trajectory points or segment endpoints.",
		},
	)

	return nil, coordinateEvidence{
		limitations: limitations,
	}
}

func usablePointCoordinates(
	ctx context.Context,
	points []trajectory.TrackPoint4D,
) ([]coordinate, int) {
	result := make([]coordinate, 0, len(points))
	invalidCount := 0

	for index, point := range points {
		if index%1024 == 0 && ctx.Err() != nil {
			return nil, invalidCount
		}

		value, valid := normalizeCoordinate(
			point.Latitude,
			point.Longitude,
		)
		if !valid {
			invalidCount++
			continue
		}
		result = append(result, value)
	}

	return result, invalidCount
}

func usableSegmentCoordinates(
	ctx context.Context,
	segments []trajectory.TrajectorySegment,
) ([]coordinate, int) {
	ordered := append(
		[]trajectory.TrajectorySegment(nil),
		segments...,
	)
	sort.SliceStable(
		ordered,
		func(left int, right int) bool {
			if ordered[left].SequenceNumber !=
				ordered[right].SequenceNumber {
				return ordered[left].SequenceNumber <
					ordered[right].SequenceNumber
			}
			if !ordered[left].StartTime.Equal(
				ordered[right].StartTime,
			) {
				return ordered[left].StartTime.Before(
					ordered[right].StartTime,
				)
			}

			return ordered[left].ID < ordered[right].ID
		},
	)

	result := make([]coordinate, 0, len(ordered)*2)
	invalidCount := 0

	appendIfUsable := func(
		latitude float64,
		longitude float64,
	) {
		value, valid := normalizeCoordinate(
			latitude,
			longitude,
		)
		if !valid {
			invalidCount++
			return
		}
		if len(result) == 0 ||
			!result[len(result)-1].equal(value) {
			result = append(result, value)
		}
	}

	for index, segment := range ordered {
		if index%1024 == 0 && ctx.Err() != nil {
			return nil, invalidCount
		}
		if segment.Status == trajectory.SegmentStatusInvalid {
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

	return result, invalidCount
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

func cloneFeatures(
	features flightfeatures.GeographicalFeatures,
) flightfeatures.GeographicalFeatures {
	cloned := features
	cloned.Evidence.Limitations = append(
		[]flightfeatures.FeatureLimitation(nil),
		features.Evidence.Limitations...,
	)

	return cloned
}
