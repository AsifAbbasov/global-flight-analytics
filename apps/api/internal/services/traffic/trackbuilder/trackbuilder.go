package trackbuilder

import (
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/gapdetector"
)

type Config struct {
	GapDetectorConfig    gapdetector.Config
	MinimumSegmentPoints int
}

type Builder struct {
	config Config
}

func DefaultConfig() Config {
	return Config{
		GapDetectorConfig:    gapdetector.DefaultConfig(),
		MinimumSegmentPoints: 2,
	}
}

func NewBuilder(config Config) *Builder {
	defaultConfig := DefaultConfig()

	if config.GapDetectorConfig.MaxTimeGap <= 0 {
		config.GapDetectorConfig.MaxTimeGap = defaultConfig.GapDetectorConfig.MaxTimeGap
	}

	if config.GapDetectorConfig.MaxGroundSpeedMPS <= 0 {
		config.GapDetectorConfig.MaxGroundSpeedMPS = defaultConfig.GapDetectorConfig.MaxGroundSpeedMPS
	}

	if config.MinimumSegmentPoints <= 0 {
		config.MinimumSegmentPoints = defaultConfig.MinimumSegmentPoints
	}

	return &Builder{config: config}
}

func (builder *Builder) Build(states []flightstate.FlightState) trajectory.FlightTrajectory {
	if len(states) == 0 {
		return trajectory.FlightTrajectory{}
	}

	sortedStates := copyAndSortStates(states)
	points := make([]trajectory.TrackPoint4D, 0, len(sortedStates))
	segments := make([]trajectory.TrajectorySegment, 0)
	coverageGaps := make([]trajectory.CoverageGap, 0)

	for _, state := range sortedStates {
		points = append(points, toTrackPoint4D(state))
	}

	segmentStartIndex := 0

	for index := 1; index < len(sortedStates); index++ {
		previous := sortedStates[index-1]
		current := sortedStates[index]

		gap := gapdetector.Detect(previous, current, builder.config.GapDetectorConfig)

		if !gap.HasGap {
			continue
		}

		segments = append(
			segments,
			builder.buildSegment(sortedStates[segmentStartIndex:index], len(segments)+1),
		)

		coverageGaps = append(coverageGaps, trajectory.CoverageGap{
			ICAO24:          current.ICAO24,
			StartTime:       previous.ObservedAt,
			EndTime:         current.ObservedAt,
			DurationSeconds: int64(gap.Duration.Seconds()),
			DistanceKm:      gap.DistanceKm,
			Reason:          gap.Reason,
			CreatedAt:       time.Now().UTC(),
		})

		segmentStartIndex = index
	}

	segments = append(
		segments,
		builder.buildSegment(sortedStates[segmentStartIndex:], len(segments)+1),
	)

	firstState := sortedStates[0]
	lastState := sortedStates[len(sortedStates)-1]

	return trajectory.FlightTrajectory{
		FlightID:         firstState.FlightID,
		AircraftID:       firstState.AircraftID,
		ICAO24:           firstState.ICAO24,
		Callsign:         firstNonEmptyCallsign(sortedStates),
		StartTime:        firstState.ObservedAt,
		EndTime:          lastState.ObservedAt,
		DurationSeconds:  int64(lastState.ObservedAt.Sub(firstState.ObservedAt).Seconds()),
		SegmentCount:     len(segments),
		PointCount:       len(points),
		CoverageGapCount: len(coverageGaps),
		QualityScore:     calculateTrajectoryQuality(segments),
		SourceName:       dominantSourceName(sortedStates),
		Points:           points,
		Segments:         segments,
		CoverageGaps:     coverageGaps,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
}

func (builder *Builder) BuildMany(states []flightstate.FlightState) map[string]trajectory.FlightTrajectory {
	groupedStates := make(map[string][]flightstate.FlightState)

	for _, state := range states {
		if state.ICAO24 == "" {
			continue
		}

		groupedStates[state.ICAO24] = append(groupedStates[state.ICAO24], state)
	}

	result := make(map[string]trajectory.FlightTrajectory, len(groupedStates))

	for icao24, group := range groupedStates {
		result[icao24] = builder.Build(group)
	}

	return result
}

func (builder *Builder) buildSegment(states []flightstate.FlightState, sequenceNumber int) trajectory.TrajectorySegment {
	firstState := states[0]
	lastState := states[len(states)-1]

	return trajectory.TrajectorySegment{
		FlightID:        firstState.FlightID,
		AircraftID:      firstState.AircraftID,
		ICAO24:          firstState.ICAO24,
		Callsign:        firstNonEmptyCallsign(states),
		SequenceNumber:  sequenceNumber,
		Status:          trajectory.SegmentStatusObserved,
		QualityScore:    builder.calculateSegmentQuality(states),
		StartTime:       firstState.ObservedAt,
		EndTime:         lastState.ObservedAt,
		DurationSeconds: int64(lastState.ObservedAt.Sub(firstState.ObservedAt).Seconds()),
		StartLatitude:   firstState.Latitude,
		StartLongitude:  firstState.Longitude,
		EndLatitude:     lastState.Latitude,
		EndLongitude:    lastState.Longitude,
		PointCount:      len(states),
		SourceName:      dominantSourceName(states),
		CreatedAt:       time.Now().UTC(),
	}
}

func (builder *Builder) calculateSegmentQuality(states []flightstate.FlightState) float64 {
	if len(states) == 0 {
		return 0
	}

	if len(states) < builder.config.MinimumSegmentPoints {
		return 0.45
	}

	firstState := states[0]
	lastState := states[len(states)-1]
	duration := lastState.ObservedAt.Sub(firstState.ObservedAt)

	if duration <= 0 {
		return 0.35
	}

	averageInterval := duration.Seconds() / float64(len(states)-1)
	maxAllowedInterval := builder.config.GapDetectorConfig.MaxTimeGap.Seconds()

	score := 1.0

	if averageInterval > maxAllowedInterval/2 {
		score -= 0.25
	}

	if len(states) < 4 {
		score -= 0.15
	}

	return clampScore(score)
}

func copyAndSortStates(states []flightstate.FlightState) []flightstate.FlightState {
	result := make([]flightstate.FlightState, len(states))
	copy(result, states)

	sort.SliceStable(result, func(left int, right int) bool {
		if result[left].ICAO24 == result[right].ICAO24 {
			return result[left].ObservedAt.Before(result[right].ObservedAt)
		}

		return result[left].ICAO24 < result[right].ICAO24
	})

	return result
}

func toTrackPoint4D(state flightstate.FlightState) trajectory.TrackPoint4D {
	return trajectory.TrackPoint4D{
		FlightStateID:       state.ID,
		FlightID:            state.FlightID,
		AircraftID:          state.AircraftID,
		ICAO24:              state.ICAO24,
		Callsign:            state.Callsign,
		Latitude:            state.Latitude,
		Longitude:           state.Longitude,
		BarometricAltitudeM: state.BarometricAltitudeM,
		GeometricAltitudeM:  state.GeometricAltitudeM,
		VelocityMPS:         state.VelocityMPS,
		HeadingDegrees:      state.HeadingDegrees,
		VerticalRateMPS:     state.VerticalRateMPS,
		OnGround:            state.OnGround,
		OriginCountry:       state.OriginCountry,
		ObservedAt:          state.ObservedAt,
		SourceName:          state.SourceName,
	}
}

func calculateTrajectoryQuality(segments []trajectory.TrajectorySegment) float64 {
	if len(segments) == 0 {
		return 0
	}

	var weightedScore float64
	var totalPoints int

	for _, segment := range segments {
		weightedScore += segment.QualityScore * float64(segment.PointCount)
		totalPoints += segment.PointCount
	}

	if totalPoints == 0 {
		return 0
	}

	return clampScore(weightedScore / float64(totalPoints))
}

func firstNonEmptyCallsign(states []flightstate.FlightState) string {
	for _, state := range states {
		if state.Callsign != "" {
			return state.Callsign
		}
	}

	return ""
}

func dominantSourceName(states []flightstate.FlightState) string {
	if len(states) == 0 {
		return ""
	}

	sourceName := states[0].SourceName

	for _, state := range states {
		if state.SourceName != sourceName {
			return "mixed"
		}
	}

	return sourceName
}

func clampScore(score float64) float64 {
	if score < 0 {
		return 0
	}

	if score > 1 {
		return 1
	}

	return score
}
