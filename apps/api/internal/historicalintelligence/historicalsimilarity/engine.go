package historicalsimilarity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const earthRadiusKM = 6371.0088

type Config struct {
	MinimumPointCount int
	SampleCount       int

	MaximumMeanDistanceKM     float64
	MaximumEndpointDistanceKM float64

	GeometryWeight   float64
	EndpointsWeight  float64
	PathLengthWeight float64
	DurationWeight   float64
}

func DefaultConfig() Config {
	return Config{
		MinimumPointCount: DefaultMinimumPointCount,
		SampleCount:       DefaultSampleCount,

		MaximumMeanDistanceKM:     250,
		MaximumEndpointDistanceKM: 100,

		GeometryWeight:   0.55,
		EndpointsWeight:  0.20,
		PathLengthWeight: 0.15,
		DurationWeight:   0.10,
	}
}

func (config Config) Validate() error {
	if config.MinimumPointCount < 2 {
		return ErrMinimumPointCountInvalid
	}
	if config.SampleCount < 2 {
		return ErrSampleCountInvalid
	}
	if !finite(config.MaximumMeanDistanceKM) ||
		config.MaximumMeanDistanceKM <= 0 ||
		!finite(config.MaximumEndpointDistanceKM) ||
		config.MaximumEndpointDistanceKM <= 0 {
		return ErrDistanceThresholdInvalid
	}

	weights := []float64{
		config.GeometryWeight,
		config.EndpointsWeight,
		config.PathLengthWeight,
		config.DurationWeight,
	}
	total := 0.0
	for _, weight := range weights {
		if !finite(weight) || weight < 0 {
			return ErrWeightInvalid
		}
		total += weight
	}
	if math.Abs(total-1) > 1e-9 {
		return ErrWeightInvalid
	}

	return nil
}

type Engine struct {
	config Config
}

func New(config Config) (*Engine, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Engine{config: config}, nil
}

func NewDefault() *Engine {
	engine, err := New(DefaultConfig())
	if err != nil {
		panic(
			fmt.Sprintf(
				"default historical similarity config is invalid: %v",
				err,
			),
		)
	}

	return engine
}

type preparedTrajectory struct {
	id              string
	points          []geoPoint
	samples         []geoPoint
	pathLengthKM    float64
	durationSeconds float64
	limitations     []Notice
}

type geoPoint struct {
	latitude   float64
	longitude  float64
	observedAt time.Time
}

func (engine *Engine) Compare(
	reference trajectory.FlightTrajectory,
	candidate trajectory.FlightTrajectory,
) (Result, error) {
	referenceID := strings.TrimSpace(reference.ID)
	candidateID := strings.TrimSpace(candidate.ID)
	if referenceID != "" &&
		referenceID == candidateID {
		return Result{}, ErrSameTrajectory
	}

	left, err := engine.prepare(reference)
	if err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %v",
				ErrReferenceNotComparable,
				err,
			)
	}
	right, err := engine.prepare(candidate)
	if err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %v",
				ErrCandidateNotComparable,
				err,
			)
	}

	meanDistance, maximumDistance :=
		sampleDistances(
			left.samples,
			right.samples,
		)
	startDistance := haversineKM(
		left.samples[0],
		right.samples[0],
	)
	endDistance := haversineKM(
		left.samples[len(left.samples)-1],
		right.samples[len(right.samples)-1],
	)

	geometryScore := inverseThresholdScore(
		meanDistance,
		engine.config.MaximumMeanDistanceKM,
	)
	endpointsObserved := (startDistance +
		endDistance) / 2
	endpointsScore := inverseThresholdScore(
		endpointsObserved,
		engine.config.MaximumEndpointDistanceKM,
	)
	pathLengthDifference := relativeDifference(
		left.pathLengthKM,
		right.pathLengthKM,
	)
	pathLengthScore := 1 -
		pathLengthDifference
	durationDifference := relativeDifference(
		left.durationSeconds,
		right.durationSeconds,
	)
	durationScore := 1 -
		durationDifference

	score := geometryScore*
		engine.config.GeometryWeight +
		endpointsScore*
			engine.config.EndpointsWeight +
		pathLengthScore*
			engine.config.PathLengthWeight +
		durationScore*
			engine.config.DurationWeight
	score = clampRatio(score)

	limitations := append(
		append(
			[]Notice(nil),
			left.limitations...,
		),
		right.limitations...,
	)
	limitations = normalizeNotices(limitations)

	result := Result{
		Version: Version,

		ReferenceTrajectoryID: left.id,
		CandidateTrajectoryID: right.id,

		Score: score,
		Level: LevelForScore(score),

		ReferencePointCount: len(left.points),
		CandidatePointCount: len(right.points),
		SampleCount:         engine.config.SampleCount,

		MeanDistanceKM:           meanDistance,
		MaximumDistanceKM:        maximumDistance,
		StartEndpointDistanceKM:  startDistance,
		EndEndpointDistanceKM:    endDistance,
		ReferencePathLengthKM:    left.pathLengthKM,
		CandidatePathLengthKM:    right.pathLengthKM,
		ReferenceDurationSeconds: left.durationSeconds,
		CandidateDurationSeconds: right.durationSeconds,

		Components: []Component{
			{
				Name:          ComponentGeometry,
				Score:         geometryScore,
				Weight:        engine.config.GeometryWeight,
				ObservedValue: meanDistance,
				Unit:          "kilometres",
			},
			{
				Name:          ComponentEndpoints,
				Score:         endpointsScore,
				Weight:        engine.config.EndpointsWeight,
				ObservedValue: endpointsObserved,
				Unit:          "kilometres",
			},
			{
				Name:          ComponentPathLength,
				Score:         pathLengthScore,
				Weight:        engine.config.PathLengthWeight,
				ObservedValue: pathLengthDifference,
				Unit:          "ratio",
			},
			{
				Name:          ComponentDuration,
				Score:         durationScore,
				Weight:        engine.config.DurationWeight,
				ObservedValue: durationDifference,
				Unit:          "ratio",
			},
		},
		Reasons: []string{
			"geometry_shape_similarity",
			"endpoint_proximity",
			"path_length_similarity",
			"duration_similarity",
		},
		Limitations: limitations,
		InputFingerprint: comparisonFingerprint(
			reference,
			candidate,
			engine.config,
		),
	}

	if err := result.Validate(); err != nil {
		return Result{}, err
	}

	return result.Clone(), nil
}

func (engine *Engine) Rank(
	reference trajectory.FlightTrajectory,
	candidates []trajectory.FlightTrajectory,
	limit int,
) ([]Result, error) {
	if limit < 1 || limit > MaximumRankLimit {
		return nil, ErrRankLimitInvalid
	}
	if _, err := engine.prepare(reference); err != nil {
		return nil,
			fmt.Errorf(
				"%w: %v",
				ErrReferenceNotComparable,
				err,
			)
	}

	results := make([]Result, 0, len(candidates))
	for _, candidate := range candidates {
		result, err := engine.Compare(
			reference,
			candidate,
		)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	sort.SliceStable(
		results,
		func(left int, right int) bool {
			if results[left].Score !=
				results[right].Score {
				return results[left].Score >
					results[right].Score
			}
			return results[left].
				CandidateTrajectoryID <
				results[right].
					CandidateTrajectoryID
		},
	)

	if len(results) > limit {
		results = results[:limit]
	}

	cloned := make([]Result, len(results))
	for index, result := range results {
		cloned[index] = result.Clone()
	}

	return cloned, nil
}

func (engine *Engine) prepare(
	item trajectory.FlightTrajectory,
) (preparedTrajectory, error) {
	id := strings.TrimSpace(item.ID)
	if id == "" {
		return preparedTrajectory{},
			fmt.Errorf("trajectory identifier is required")
	}

	type indexedPoint struct {
		point geoPoint
		index int
	}
	valid := make(
		[]indexedPoint,
		0,
		len(item.Points),
	)
	excludedCount := 0

	for index, point := range item.Points {
		if point.ObservedAt.IsZero() ||
			!validLatitude(point.Latitude) ||
			!validLongitude(point.Longitude) {
			excludedCount++
			continue
		}

		valid = append(
			valid,
			indexedPoint{
				point: geoPoint{
					latitude:   point.Latitude,
					longitude:  normalizeLongitude(point.Longitude),
					observedAt: point.ObservedAt.UTC(),
				},
				index: index,
			},
		)
	}

	sort.SliceStable(
		valid,
		func(left int, right int) bool {
			if valid[left].point.observedAt.Equal(
				valid[right].point.observedAt,
			) {
				return valid[left].index <
					valid[right].index
			}
			return valid[left].point.observedAt.Before(
				valid[right].point.observedAt,
			)
		},
	)

	if len(valid) < engine.config.MinimumPointCount {
		return preparedTrajectory{},
			fmt.Errorf(
				"usable points=%d minimum=%d",
				len(valid),
				engine.config.MinimumPointCount,
			)
	}

	points := make([]geoPoint, len(valid))
	for index, point := range valid {
		points[index] = point.point
	}

	duration := points[len(points)-1].
		observedAt.Sub(
		points[0].observedAt,
	).Seconds()
	if duration <= 0 {
		return preparedTrajectory{},
			fmt.Errorf(
				"trajectory observation duration must be positive",
			)
	}

	pathLength := trajectoryLengthKM(points)
	samples, usedIndexFallback := resample(
		points,
		engine.config.SampleCount,
		pathLength,
	)

	limitations := make([]Notice, 0, 2)
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
	if usedIndexFallback {
		limitations = append(
			limitations,
			Notice{
				Code:    "zero_length_path_index_resampling",
				Message: "The trajectory had zero geographic path length, so normalized point-index resampling was used.",
			},
		)
	}

	return preparedTrajectory{
		id:              id,
		points:          points,
		samples:         samples,
		pathLengthKM:    pathLength,
		durationSeconds: duration,
		limitations:     limitations,
	}, nil
}

func resample(
	points []geoPoint,
	sampleCount int,
	pathLengthKM float64,
) ([]geoPoint, bool) {
	if pathLengthKM <= 1e-9 {
		return resampleByIndex(
			points,
			sampleCount,
		), true
	}

	cumulative := make([]float64, len(points))
	for index := 1; index < len(points); index++ {
		cumulative[index] =
			cumulative[index-1] +
				haversineKM(
					points[index-1],
					points[index],
				)
	}

	result := make([]geoPoint, 0, sampleCount)
	for sampleIndex := 0; sampleIndex < sampleCount; sampleIndex++ {
		target := pathLengthKM *
			float64(sampleIndex) /
			float64(sampleCount-1)
		segment := sort.Search(
			len(cumulative),
			func(index int) bool {
				return cumulative[index] >= target
			},
		)
		switch {
		case segment <= 0:
			result = append(result, points[0])
		case segment >= len(points):
			result = append(
				result,
				points[len(points)-1],
			)
		default:
			startDistance := cumulative[segment-1]
			endDistance := cumulative[segment]
			fraction := 0.0
			if endDistance > startDistance {
				fraction = (target - startDistance) /
					(endDistance - startDistance)
			}
			result = append(
				result,
				interpolateGeoPoint(
					points[segment-1],
					points[segment],
					fraction,
				),
			)
		}
	}

	return result, false
}

func resampleByIndex(
	points []geoPoint,
	sampleCount int,
) []geoPoint {
	result := make([]geoPoint, 0, sampleCount)
	maxIndex := float64(len(points) - 1)

	for sampleIndex := 0; sampleIndex < sampleCount; sampleIndex++ {
		position := maxIndex *
			float64(sampleIndex) /
			float64(sampleCount-1)
		left := int(math.Floor(position))
		right := int(math.Ceil(position))
		if right >= len(points) {
			right = len(points) - 1
		}
		fraction := position - float64(left)
		result = append(
			result,
			interpolateGeoPoint(
				points[left],
				points[right],
				fraction,
			),
		)
	}

	return result
}

func interpolateGeoPoint(
	left geoPoint,
	right geoPoint,
	fraction float64,
) geoPoint {
	fraction = clampRatio(fraction)
	longitudeDelta := right.longitude -
		left.longitude
	if longitudeDelta > 180 {
		longitudeDelta -= 360
	} else if longitudeDelta < -180 {
		longitudeDelta += 360
	}

	duration := right.observedAt.Sub(
		left.observedAt,
	)
	return geoPoint{
		latitude: left.latitude +
			(right.latitude-left.latitude)*
				fraction,
		longitude: normalizeLongitude(
			left.longitude +
				longitudeDelta*fraction,
		),
		observedAt: left.observedAt.Add(
			time.Duration(
				float64(duration) * fraction,
			),
		),
	}
}

func sampleDistances(
	left []geoPoint,
	right []geoPoint,
) (float64, float64) {
	total := 0.0
	maximum := 0.0

	for index := range left {
		distance := haversineKM(
			left[index],
			right[index],
		)
		total += distance
		if distance > maximum {
			maximum = distance
		}
	}

	return total / float64(len(left)),
		maximum
}

func trajectoryLengthKM(
	points []geoPoint,
) float64 {
	total := 0.0
	for index := 1; index < len(points); index++ {
		total += haversineKM(
			points[index-1],
			points[index],
		)
	}
	return total
}

func haversineKM(
	left geoPoint,
	right geoPoint,
) float64 {
	leftLatitude := degreesToRadians(
		left.latitude,
	)
	rightLatitude := degreesToRadians(
		right.latitude,
	)
	latitudeDelta := rightLatitude -
		leftLatitude
	longitudeDelta := degreesToRadians(
		right.longitude - left.longitude,
	)

	sineLatitude := math.Sin(
		latitudeDelta / 2,
	)
	sineLongitude := math.Sin(
		longitudeDelta / 2,
	)
	value := sineLatitude*sineLatitude +
		math.Cos(leftLatitude)*
			math.Cos(rightLatitude)*
			sineLongitude*sineLongitude
	value = math.Min(
		1,
		math.Max(0, value),
	)

	return earthRadiusKM *
		2 *
		math.Atan2(
			math.Sqrt(value),
			math.Sqrt(1-value),
		)
}

func inverseThresholdScore(
	value float64,
	threshold float64,
) float64 {
	return clampRatio(
		1 - value/threshold,
	)
}

func relativeDifference(
	left float64,
	right float64,
) float64 {
	scale := math.Max(
		math.Max(left, right),
		1,
	)
	return clampRatio(
		math.Abs(left-right) / scale,
	)
}

func clampRatio(value float64) float64 {
	switch {
	case value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}

func validLatitude(value float64) bool {
	return finite(value) &&
		value >= -90 &&
		value <= 90
}

func validLongitude(value float64) bool {
	return finite(value) &&
		value >= -180 &&
		value <= 180
}

func normalizeLongitude(
	value float64,
) float64 {
	for value > 180 {
		value -= 360
	}
	for value < -180 {
		value += 360
	}
	return value
}

func degreesToRadians(
	value float64,
) float64 {
	return value * math.Pi / 180
}

func normalizeNotices(
	values []Notice,
) []Notice {
	seen := make(map[string]struct{})
	result := make([]Notice, 0, len(values))

	for _, value := range values {
		value.Code = strings.TrimSpace(value.Code)
		value.Message = strings.TrimSpace(
			value.Message,
		)
		if value.Code == "" ||
			value.Message == "" {
			continue
		}
		key := value.Code + "\x00" +
			value.Message
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}

	sort.SliceStable(
		result,
		func(left int, right int) bool {
			if result[left].Code !=
				result[right].Code {
				return result[left].Code <
					result[right].Code
			}
			return result[left].Message <
				result[right].Message
		},
	)

	return result
}

func comparisonFingerprint(
	reference trajectory.FlightTrajectory,
	candidate trajectory.FlightTrajectory,
	config Config,
) string {
	records := []string{
		FingerprintVersion,
		fmt.Sprintf(
			"config|%d|%d|%.12f|%.12f|%.12f|%.12f|%.12f|%.12f",
			config.MinimumPointCount,
			config.SampleCount,
			config.MaximumMeanDistanceKM,
			config.MaximumEndpointDistanceKM,
			config.GeometryWeight,
			config.EndpointsWeight,
			config.PathLengthWeight,
			config.DurationWeight,
		),
	}
	records = append(
		records,
		trajectoryFingerprintRecords(
			"reference",
			reference,
		)...,
	)
	records = append(
		records,
		trajectoryFingerprintRecords(
			"candidate",
			candidate,
		)...,
	)

	sort.Strings(records)
	sum := sha256.Sum256(
		[]byte(strings.Join(records, "\n")),
	)
	return "sha256:" +
		hex.EncodeToString(sum[:])
}

func trajectoryFingerprintRecords(
	role string,
	item trajectory.FlightTrajectory,
) []string {
	records := []string{
		fmt.Sprintf(
			"%s|trajectory|%s|%s|%s|%s",
			role,
			strings.TrimSpace(item.ID),
			strings.TrimSpace(item.IdentityKey),
			item.StartTime.UTC().
				Format(time.RFC3339Nano),
			item.EndTime.UTC().
				Format(time.RFC3339Nano),
		),
	}

	for _, point := range item.Points {
		records = append(
			records,
			fmt.Sprintf(
				"%s|point|%s|%.12f|%.12f|%s",
				role,
				strings.TrimSpace(point.ID),
				point.Latitude,
				point.Longitude,
				point.ObservedAt.UTC().
					Format(time.RFC3339Nano),
			),
		)
	}

	return records
}
