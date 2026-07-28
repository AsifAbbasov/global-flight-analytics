package historicalsimilarity

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type preparedTrajectory struct {
	id string

	points  []geoPoint
	samples []geoPoint

	pathLengthKM    float64
	durationSeconds float64

	quality     EvidenceQuality
	limitations []Notice
}

type geoPoint struct {
	sourceID  string
	latitude  float64
	longitude float64

	observedAt time.Time
}

func (engine *Engine) prepare(
	item trajectory.FlightTrajectory,
) (preparedTrajectory, error) {
	id := strings.TrimSpace(item.ID)
	if id == "" {
		return preparedTrajectory{},
			fmt.Errorf(
				"trajectory identifier is required",
			)
	}
	if len(item.Points) >
		MaximumInputPointCount {
		return preparedTrajectory{},
			fmt.Errorf(
				"%w: points=%d maximum=%d",
				ErrTrajectoryPointLimitExceeded,
				len(item.Points),
				MaximumInputPointCount,
			)
	}

	points, excludedCount,
		equalTimestampCount :=
		canonicalPoints(item.Points)
	if len(points) <
		engine.config.MinimumPointCount {
		return preparedTrajectory{},
			fmt.Errorf(
				"usable points=%d minimum=%d",
				len(points),
				engine.config.MinimumPointCount,
			)
	}

	duration := points[len(points)-1].
		observedAt.Sub(
		points[0].observedAt,
	).Seconds()
	if !finite(duration) || duration <= 0 {
		return preparedTrajectory{},
			fmt.Errorf(
				"trajectory observation duration must be finite and positive",
			)
	}

	pathLength := trajectoryLengthKM(points)
	if !finite(pathLength) || pathLength < 0 {
		return preparedTrajectory{},
			fmt.Errorf(
				"trajectory path length is invalid",
			)
	}

	resampled := resample(
		points,
		engine.config.SampleCount,
		pathLength,
	)

	quality, qualityLimitations, err :=
		assessEvidenceQuality(
			item,
			points,
			excludedCount,
			equalTimestampCount,
		)
	if err != nil {
		return preparedTrajectory{}, err
	}

	limitations := make(
		[]Notice,
		0,
		len(qualityLimitations)+4,
	)
	if excludedCount > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "trajectory_points_excluded",
				Message: fmt.Sprintf(
					"%d trajectory points without usable time or coordinates were excluded.",
					excludedCount,
				),
			},
		)
	}
	if equalTimestampCount > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "equal_timestamp_points_canonicalized",
				Message: fmt.Sprintf(
					"%d usable points shared an observation timestamp and were ordered canonically by coordinates and identifier.",
					equalTimestampCount,
				),
			},
		)
	}
	if resampled.usedIndexFallback {
		limitations = append(
			limitations,
			Notice{
				Code:    "zero_length_path_index_resampling",
				Message: "The trajectory had zero geographic path length, so normalized point-index resampling was used.",
			},
		)
	}
	if resampled.usedAntipodalFallback {
		limitations = append(
			limitations,
			Notice{
				Code:    "near_antipodal_interpolation_fallback",
				Message: "A near-antipodal segment used deterministic longitude-aware coordinate interpolation because the great-circle interpolation axis was numerically ambiguous.",
			},
		)
	}
	limitations = append(
		limitations,
		qualityLimitations...,
	)

	return preparedTrajectory{
		id:              id,
		points:          points,
		samples:         resampled.points,
		pathLengthKM:    pathLength,
		durationSeconds: duration,
		quality:         quality,
		limitations: normalizeNotices(
			limitations,
		),
	}, nil
}

func canonicalPoints(
	values []trajectory.TrackPoint4D,
) ([]geoPoint, int, int) {
	points := make(
		[]geoPoint,
		0,
		len(values),
	)
	excludedCount := 0

	for _, point := range values {
		if point.ObservedAt.IsZero() ||
			!validLatitude(point.Latitude) ||
			!validLongitude(point.Longitude) {
			excludedCount++
			continue
		}

		points = append(
			points,
			geoPoint{
				sourceID: strings.TrimSpace(
					point.ID,
				),
				latitude: point.Latitude,
				longitude: normalizeLongitude(
					point.Longitude,
				),
				observedAt: point.
					ObservedAt.UTC(),
			},
		)
	}

	sort.Slice(
		points,
		func(left int, right int) bool {
			if !points[left].observedAt.Equal(
				points[right].observedAt,
			) {
				return points[left].observedAt.
					Before(
						points[right].
							observedAt,
					)
			}
			if points[left].latitude !=
				points[right].latitude {
				return points[left].latitude <
					points[right].latitude
			}
			if points[left].longitude !=
				points[right].longitude {
				return points[left].longitude <
					points[right].longitude
			}
			return points[left].sourceID <
				points[right].sourceID
		},
	)

	equalTimestampCount := 0
	for index := 1; index < len(points); index++ {
		if points[index].observedAt.Equal(
			points[index-1].observedAt,
		) {
			equalTimestampCount++
		}
	}

	return points,
		excludedCount,
		equalTimestampCount
}
