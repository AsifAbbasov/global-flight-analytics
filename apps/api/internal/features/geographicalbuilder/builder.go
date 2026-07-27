package geographicalbuilder

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

var _ extractor.GeographicalBuilder = (*Builder)(nil)

type Builder struct {
	geographicCellPrecision int
}

type Config struct {
	// GeographicCellPrecision is the decimal-degree quantization precision.
	// Zero selects DefaultGeographicCellPrecision; effective precision must be
	// between MinimumGeographicCellPrecision and MaximumGeographicCellPrecision.
	GeographicCellPrecision int
}

type coordinateEvidence struct {
	limitations          []flightfeatures.FeatureLimitation
	pathEdges            []coordinateEdge
	supportingPointCount int
	segmentFallback      bool
}

type pointCollection struct {
	coordinates          []coordinate
	pathEdges            []coordinateEdge
	invalidCoordinateCnt int
	missingTimestampCnt  int
	outsideWindowCnt     int
}

type segmentCollection struct {
	coordinates          []coordinate
	pathEdges            []coordinateEdge
	invalidCoordinateCnt int
	invalidStatusCnt     int
	missingTimestampCnt  int
	outsideWindowCnt     int
	discontinuityCnt     int
	supportingPointCount int
}

type observedCoordinate struct {
	coordinate
	observedAt time.Time
	pointID    string
	inputIndex int
}

func New(config Config) (*Builder, error) {
	precision := config.GeographicCellPrecision
	if precision == 0 {
		precision = DefaultGeographicCellPrecision
	}
	if precision < MinimumGeographicCellPrecision ||
		precision > MaximumGeographicCellPrecision {
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
		return flightfeatures.GeographicalFeatures{}, ErrContextRequired
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.GeographicalFeatures{}, err
	}

	coordinates, evidence, err := collectCoordinates(ctx, item)
	if err != nil {
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

	minimumLatitude, maximumLatitude, err := latitudeBoundsContext(
		ctx,
		coordinates,
	)
	if err != nil {
		return flightfeatures.GeographicalFeatures{}, err
	}
	minimumLongitude, maximumLongitude, longitudeSpan, err :=
		circularLongitudeBoundsContext(ctx, coordinates)
	if err != nil {
		return flightfeatures.GeographicalFeatures{}, err
	}
	observedPathDistance, err := observedEdgeDistanceKMContext(
		ctx,
		evidence.pathEdges,
	)
	if err != nil {
		return flightfeatures.GeographicalFeatures{}, err
	}
	maximumDisplacement, err := maximumDisplacementKMContext(
		ctx,
		coordinates,
	)
	if err != nil {
		return flightfeatures.GeographicalFeatures{}, err
	}
	crossesAntimeridian, err := edgeSetCrossesAntimeridianContext(
		ctx,
		evidence.pathEdges,
	)
	if err != nil {
		return flightfeatures.GeographicalFeatures{}, err
	}
	uniqueCellCount, err := uniqueGeographicCellCountContext(
		ctx,
		coordinates,
		builder.geographicCellPrecision,
	)
	if err != nil {
		return flightfeatures.GeographicalFeatures{}, err
	}

	start := coordinates[0]
	end := coordinates[len(coordinates)-1]
	features := flightfeatures.GeographicalFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Status:               flightfeatures.AvailabilityStatusAvailable,
			AvailableFieldCount:  GeographicalFeatureFieldCount,
			TotalFieldCount:      GeographicalFeatureFieldCount,
			SupportingPointCount: evidence.supportingPointCount,
			Limitations: append(
				[]flightfeatures.FeatureLimitation(nil),
				evidence.limitations...,
			),
		},
		StartLatitude:             start.latitude,
		StartLongitude:            start.longitude,
		EndLatitude:               end.latitude,
		EndLongitude:              end.longitude,
		MinimumLatitude:           minimumLatitude,
		MaximumLatitude:           maximumLatitude,
		MinimumLongitude:          minimumLongitude,
		MaximumLongitude:          maximumLongitude,
		LatitudeSpanDegrees:       maximumLatitude - minimumLatitude,
		LongitudeSpanDegrees:      longitudeSpan,
		GreatCircleDistanceKM:     haversineDistanceKM(start, end),
		ObservedPathDistanceKM:    observedPathDistance,
		MaximumDisplacementKM:     maximumDisplacement,
		CrossesAntimeridian:       crossesAntimeridian,
		UniqueGeographicCellCount: uniqueCellCount,
		GeographicCellPrecision:   builder.geographicCellPrecision,
	}

	if len(coordinates) == 1 {
		features.Evidence.Limitations = append(
			features.Evidence.Limitations,
			flightfeatures.FeatureLimitation{
				Code:    flightfeatures.GeographicalLimitationSingleCoordinate,
				Message: "Only one usable coordinate supports the geographical feature group, so movement distances and spans are zero.",
			},
		)
	}

	if err := ctx.Err(); err != nil {
		return flightfeatures.GeographicalFeatures{}, err
	}

	return cloneFeatures(features), nil
}

func collectCoordinates(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) ([]coordinate, coordinateEvidence, error) {
	startTime, endTime, windowAvailable, err := resolveTrajectoryWindow(item)
	if err != nil {
		return nil, coordinateEvidence{}, err
	}

	points, err := collectPointCoordinates(
		ctx,
		item.Points,
		startTime,
		endTime,
		windowAvailable,
	)
	if err != nil {
		return nil, coordinateEvidence{}, err
	}
	segments, err := collectSegmentCoordinates(
		ctx,
		item.Segments,
		startTime,
		endTime,
		windowAvailable,
	)
	if err != nil {
		return nil, coordinateEvidence{}, err
	}

	limitations := make([]flightfeatures.FeatureLimitation, 0, 16)
	if !windowAvailable && (len(item.Points) > 0 || len(item.Segments) > 0) {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    flightfeatures.GeographicalLimitationTemporalWindowUnavailable,
				Message: "Trajectory start and end timestamps were unavailable, so geographical evidence retained legacy input ordering without temporal-window filtering.",
			},
		)
	}
	appendPointLimitations(&limitations, item, points)
	appendPointCountMismatch(&limitations, item)

	if len(points.coordinates) >= 2 {
		return points.coordinates, coordinateEvidence{
			limitations:          limitations,
			pathEdges:            points.pathEdges,
			supportingPointCount: len(points.coordinates),
		}, nil
	}
	if len(item.Points) > 0 && len(points.coordinates) == 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: flightfeatures.GeographicalLimitationPointEvidenceUnusable,
				Message: fmt.Sprintf(
					"None of the %d trajectory point records supplied a temporally eligible, finite, in-range coordinate.",
					len(item.Points),
				),
			},
		)
	}

	appendSegmentLimitations(&limitations, segments)
	if len(segments.coordinates) > 0 && len(points.coordinates) <= 1 {
		if len(points.coordinates) == 1 {
			limitations = append(
				limitations,
				flightfeatures.FeatureLimitation{
					Code:    flightfeatures.GeographicalLimitationSinglePointSegmentFallback,
					Message: "Only one usable point coordinate was available, so the more complete ordered segment endpoint evidence was selected.",
				},
			)
		}
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    flightfeatures.GeographicalLimitationSegmentEndpointFallback,
				Message: "Geographical features were reconstructed from ordered non-invalid trajectory segment endpoints; observed path distance includes only movement inside each usable segment and excludes discontinuities between segments.",
			},
		)

		supportingPointCount := segments.supportingPointCount
		if supportingPointCount <= 0 {
			supportingPointCount = item.PointCount
		}
		if supportingPointCount <= 0 {
			limitations = append(
				limitations,
				flightfeatures.FeatureLimitation{
					Code:    flightfeatures.GeographicalLimitationSegmentSupportingPointCountUnavailable,
					Message: "Segment fallback produced geographical coordinates, but no authoritative observation-point count was available; supporting point count remains zero.",
				},
			)
			supportingPointCount = 0
		}

		return segments.coordinates, coordinateEvidence{
			limitations:          limitations,
			pathEdges:            segments.pathEdges,
			supportingPointCount: supportingPointCount,
		}, nil
	}

	if len(points.coordinates) == 1 {
		return points.coordinates, coordinateEvidence{
			limitations:          limitations,
			pathEdges:            nil,
			supportingPointCount: 1,
		}, nil
	}

	if len(item.Segments) > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code: flightfeatures.GeographicalLimitationSegmentEvidenceUnusable,
				Message: fmt.Sprintf(
					"None of the %d trajectory segments supplied a temporally eligible, non-invalid pair of finite in-range endpoint coordinates.",
					len(item.Segments),
				),
			},
		)
	}
	limitations = append(
		limitations,
		flightfeatures.FeatureLimitation{
			Code:    flightfeatures.GeographicalLimitationCoordinatesUnavailable,
			Message: "No usable geographic coordinate was available from trajectory points or segment endpoints.",
		},
	)

	return nil, coordinateEvidence{
		limitations: limitations,
	}, nil
}

func resolveTrajectoryWindow(
	item trajectory.FlightTrajectory,
) (time.Time, time.Time, bool, error) {
	if item.StartTime.IsZero() && item.EndTime.IsZero() {
		return time.Time{}, time.Time{}, false, nil
	}
	if item.StartTime.IsZero() {
		return time.Time{}, time.Time{}, false, ErrTrajectoryStartTimeRequired
	}
	if item.EndTime.IsZero() {
		return time.Time{}, time.Time{}, false, ErrTrajectoryEndTimeRequired
	}
	if item.EndTime.Before(item.StartTime) {
		return time.Time{}, time.Time{}, false, ErrInvalidTrajectoryWindow
	}

	return item.StartTime.UTC(), item.EndTime.UTC(), true, nil
}

func collectPointCoordinates(
	ctx context.Context,
	points []trajectory.TrackPoint4D,
	startTime time.Time,
	endTime time.Time,
	windowAvailable bool,
) (pointCollection, error) {
	observed := make([]observedCoordinate, 0, len(points))
	result := pointCollection{}

	for index, point := range points {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return pointCollection{}, err
			}
		}
		if windowAvailable {
			if point.ObservedAt.IsZero() {
				result.missingTimestampCnt++
				continue
			}
			observedAt := point.ObservedAt.UTC()
			if observedAt.Before(startTime) || observedAt.After(endTime) {
				result.outsideWindowCnt++
				continue
			}
		}

		value, valid := normalizeCoordinate(point.Latitude, point.Longitude)
		if !valid {
			result.invalidCoordinateCnt++
			continue
		}
		observed = append(observed, observedCoordinate{
			coordinate: value,
			observedAt: point.ObservedAt.UTC(),
			pointID:    point.ID,
			inputIndex: index,
		})
	}

	if windowAvailable {
		sort.SliceStable(observed, func(left int, right int) bool {
			if !observed[left].observedAt.Equal(observed[right].observedAt) {
				return observed[left].observedAt.Before(observed[right].observedAt)
			}
			if observed[left].pointID != observed[right].pointID {
				return observed[left].pointID < observed[right].pointID
			}
			if observed[left].latitude != observed[right].latitude {
				return observed[left].latitude < observed[right].latitude
			}
			if observed[left].longitude != observed[right].longitude {
				return observed[left].longitude < observed[right].longitude
			}
			return observed[left].inputIndex < observed[right].inputIndex
		})
	}
	if err := ctx.Err(); err != nil {
		return pointCollection{}, err
	}

	result.coordinates = make([]coordinate, 0, len(observed))
	for _, value := range observed {
		result.coordinates = append(result.coordinates, value.coordinate)
	}
	result.pathEdges = consecutiveEdges(result.coordinates)
	return result, nil
}

func collectSegmentCoordinates(
	ctx context.Context,
	segments []trajectory.TrajectorySegment,
	startTime time.Time,
	endTime time.Time,
	windowAvailable bool,
) (segmentCollection, error) {
	ordered := append([]trajectory.TrajectorySegment(nil), segments...)
	sort.SliceStable(ordered, func(left int, right int) bool {
		if ordered[left].SequenceNumber != ordered[right].SequenceNumber {
			return ordered[left].SequenceNumber < ordered[right].SequenceNumber
		}
		if !ordered[left].StartTime.Equal(ordered[right].StartTime) {
			return ordered[left].StartTime.Before(ordered[right].StartTime)
		}
		return ordered[left].ID < ordered[right].ID
	})
	if err := ctx.Err(); err != nil {
		return segmentCollection{}, err
	}

	result := segmentCollection{
		coordinates: make([]coordinate, 0, len(ordered)*2),
		pathEdges:   make([]coordinateEdge, 0, len(ordered)),
	}
	var previousEnd coordinate
	hasPreviousEnd := false

	for index, segment := range ordered {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return segmentCollection{}, err
			}
		}
		switch segment.Status {
		case trajectory.SegmentStatusObserved,
			trajectory.SegmentStatusInterpolated,
			trajectory.SegmentStatusEstimated:
		default:
			result.invalidStatusCnt++
			continue
		}
		if windowAvailable {
			if segment.StartTime.IsZero() || segment.EndTime.IsZero() {
				result.missingTimestampCnt++
				continue
			}
			segmentStart := segment.StartTime.UTC()
			segmentEnd := segment.EndTime.UTC()
			if segmentEnd.Before(segmentStart) ||
				segmentStart.Before(startTime) ||
				segmentEnd.After(endTime) {
				result.outsideWindowCnt++
				continue
			}
		}

		start, startValid := normalizeCoordinate(
			segment.StartLatitude,
			segment.StartLongitude,
		)
		end, endValid := normalizeCoordinate(
			segment.EndLatitude,
			segment.EndLongitude,
		)
		if !startValid || !endValid {
			result.invalidCoordinateCnt++
			continue
		}

		if hasPreviousEnd && !previousEnd.equal(start) {
			result.discontinuityCnt++
		}
		appendCoordinateWithoutAdjacentDuplicate(&result.coordinates, start)
		appendCoordinateWithoutAdjacentDuplicate(&result.coordinates, end)
		result.pathEdges = append(result.pathEdges, coordinateEdge{
			start: start,
			end:   end,
		})
		if segment.PointCount > 0 {
			result.supportingPointCount += segment.PointCount
		}
		previousEnd = end
		hasPreviousEnd = true
	}

	return result, nil
}

func appendPointLimitations(
	limitations *[]flightfeatures.FeatureLimitation,
	item trajectory.FlightTrajectory,
	points pointCollection,
) {
	if points.invalidCoordinateCnt > 0 {
		*limitations = append(*limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.GeographicalLimitationInvalidPointCoordinates,
			Message: fmt.Sprintf(
				"%d trajectory point coordinates were non-finite or outside valid latitude/longitude ranges and were excluded.",
				points.invalidCoordinateCnt,
			),
		})
	}
	if points.missingTimestampCnt > 0 {
		*limitations = append(*limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.GeographicalLimitationPointTimestampMissing,
			Message: fmt.Sprintf(
				"%d trajectory point timestamps were missing and their coordinates were excluded because chronological ordering could not be proven.",
				points.missingTimestampCnt,
			),
		})
	}
	if points.outsideWindowCnt > 0 {
		*limitations = append(*limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.GeographicalLimitationPointOutsideWindow,
			Message: fmt.Sprintf(
				"%d trajectory point timestamps were outside the authoritative trajectory window and their coordinates were excluded.",
				points.outsideWindowCnt,
			),
		})
	}
	if len(item.Points) == 0 {
		*limitations = append(*limitations, flightfeatures.FeatureLimitation{
			Code:    flightfeatures.GeographicalLimitationPointEvidenceUnavailable,
			Message: "No trajectory point records were available for geographical feature extraction.",
		})
	}
}

func appendPointCountMismatch(
	limitations *[]flightfeatures.FeatureLimitation,
	item trajectory.FlightTrajectory,
) {
	if len(item.Points) == 0 || item.PointCount == len(item.Points) {
		return
	}
	*limitations = append(*limitations, flightfeatures.FeatureLimitation{
		Code: flightfeatures.LimitationTrajectoryPointCountMetadataMismatch,
		Message: fmt.Sprintf(
			"Trajectory point-count metadata reports %d points while %d point records are present.",
			item.PointCount,
			len(item.Points),
		),
	})
}

func appendSegmentLimitations(
	limitations *[]flightfeatures.FeatureLimitation,
	segments segmentCollection,
) {
	if segments.invalidStatusCnt > 0 {
		*limitations = append(*limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.GeographicalLimitationInvalidSegmentStatus,
			Message: fmt.Sprintf(
				"%d invalid-status trajectory segments were excluded from geographical fallback evidence.",
				segments.invalidStatusCnt,
			),
		})
	}
	if segments.invalidCoordinateCnt > 0 {
		*limitations = append(*limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.GeographicalLimitationInvalidSegmentCoordinates,
			Message: fmt.Sprintf(
				"%d trajectory segments contained a non-finite or out-of-range endpoint coordinate and were excluded from geographical fallback evidence.",
				segments.invalidCoordinateCnt,
			),
		})
	}
	if segments.missingTimestampCnt > 0 {
		*limitations = append(*limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.GeographicalLimitationSegmentTimestampMissing,
			Message: fmt.Sprintf(
				"%d trajectory segments had a missing boundary timestamp and were excluded because temporal eligibility could not be proven.",
				segments.missingTimestampCnt,
			),
		})
	}
	if segments.outsideWindowCnt > 0 {
		*limitations = append(*limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.GeographicalLimitationSegmentOutsideWindow,
			Message: fmt.Sprintf(
				"%d trajectory segments were outside or inconsistent with the authoritative trajectory window and were excluded.",
				segments.outsideWindowCnt,
			),
		})
	}
	if segments.discontinuityCnt > 0 {
		*limitations = append(*limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.GeographicalLimitationSegmentDiscontinuityExcluded,
			Message: fmt.Sprintf(
				"%d discontinuities between ordered trajectory segments were excluded from observed path distance and antimeridian path-crossing calculations.",
				segments.discontinuityCnt,
			),
		})
	}
}

func appendCoordinateWithoutAdjacentDuplicate(
	coordinates *[]coordinate,
	value coordinate,
) {
	if len(*coordinates) == 0 || !(*coordinates)[len(*coordinates)-1].equal(value) {
		*coordinates = append(*coordinates, value)
	}
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
