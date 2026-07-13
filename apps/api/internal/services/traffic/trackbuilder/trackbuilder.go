package trackbuilder

import (
	"fmt"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/flightsplitter"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/gapdetector"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/trajectoryquality"
)

type Config struct {
	GapDetectorConfig gapdetector.Config
}

type Builder struct {
	config Config
}

func NewBuilder(
	config Config,
) (*Builder, error) {
	if err := config.GapDetectorConfig.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate gap detector config: %w",
			err,
		)
	}

	return &Builder{
		config: config,
	}, nil
}

func (builder *Builder) BuildMany(
	inputs []InputState,
) map[string]trajectory.FlightTrajectory {
	groups := flightsplitter.Split(
		inputs,
	)
	groupsPerAircraft := make(map[string]int)
	for _, group := range groups {
		groupsPerAircraft[group.ICAO24]++
	}

	result := make(
		map[string]trajectory.FlightTrajectory,
		len(groups),
	)

	for _, group := range groups {
		item := builder.build(group)
		collectionKey := group.IdentityKey
		if groupsPerAircraft[group.ICAO24] == 1 {
			// Preserve the existing unambiguous lookup contract while the
			// processing result still uses a map. Multiple flights of the same
			// aircraft are always keyed by their distinct logical identities.
			collectionKey = group.ICAO24
		}
		result[collectionKey] = item
	}

	return result
}

func (builder *Builder) build(
	group flightsplitter.Group,
) trajectory.FlightTrajectory {
	inputs := group.Observations
	if len(inputs) == 0 {
		return trajectory.FlightTrajectory{}
	}

	sortedInputs := copyAndSortInputs(
		inputs,
	)

	points := make(
		[]trajectory.TrackPoint4D,
		0,
		len(sortedInputs),
	)

	segments := make(
		[]trajectory.TrajectorySegment,
		0,
	)

	coverageGaps := make(
		[]trajectory.CoverageGap,
		0,
	)

	for _, input := range sortedInputs {
		points = append(
			points,
			toTrackPoint4D(
				input.State,
			),
		)
	}

	segmentStartIndex := 0

	for index := 1; index < len(sortedInputs); index++ {
		previous := sortedInputs[index-1].State
		current := sortedInputs[index].State

		gap := gapdetector.Detect(
			previous,
			current,
			builder.config.GapDetectorConfig,
		)

		if !gap.HasGap {
			continue
		}

		segments = append(
			segments,
			builder.buildSegment(
				sortedInputs[segmentStartIndex:index],
				len(segments)+1,
			),
		)

		coverageGaps = append(
			coverageGaps,
			trajectory.CoverageGap{
				ICAO24:          current.ICAO24,
				StartTime:       previous.ObservedAt,
				EndTime:         current.ObservedAt,
				DurationSeconds: int64(gap.Duration.Seconds()),
				DistanceKm:      gap.DistanceKm,
				Reason:          gap.Reason,
				CreatedAt:       time.Now().UTC(),
			},
		)

		segmentStartIndex = index
	}

	segments = append(
		segments,
		builder.buildSegment(
			sortedInputs[segmentStartIndex:],
			len(segments)+1,
		),
	)

	firstState := sortedInputs[0].State
	lastState := sortedInputs[len(sortedInputs)-1].State

	return trajectory.FlightTrajectory{
		IdentityKey:   group.IdentityKey,
		IdentityBasis: group.IdentityBasis,
		SplitReason:   group.SplitReason,
		FlightID:      firstNonEmptyFlightID(sortedInputs),
		AircraftID:    firstState.AircraftID,
		ICAO24:        firstState.ICAO24,
		Callsign: firstNonEmptyCallsign(
			sortedInputs,
		),
		StartTime: firstState.ObservedAt,
		EndTime:   lastState.ObservedAt,
		DurationSeconds: int64(
			lastState.ObservedAt.
				Sub(firstState.ObservedAt).
				Seconds(),
		),
		SegmentCount:     len(segments),
		PointCount:       len(points),
		CoverageGapCount: len(coverageGaps),
		QualityScore: trajectoryquality.TrajectoryScore(
			segments,
		),
		SourceName: dominantSourceName(
			sortedInputs,
		),
		Points:       points,
		Segments:     segments,
		CoverageGaps: coverageGaps,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}

func (builder *Builder) buildSegment(
	inputs []InputState,
	sequenceNumber int,
) trajectory.TrajectorySegment {
	firstState := inputs[0].State
	lastState := inputs[len(inputs)-1].State

	return trajectory.TrajectorySegment{
		FlightID:       firstNonEmptyFlightID(inputs),
		AircraftID:     firstState.AircraftID,
		ICAO24:         firstState.ICAO24,
		Callsign:       firstNonEmptyCallsign(inputs),
		SequenceNumber: sequenceNumber,
		Status:         trajectory.SegmentStatusObserved,
		QualityScore: segmentQualityScore(
			inputs,
		),
		StartTime: firstState.ObservedAt,
		EndTime:   lastState.ObservedAt,
		DurationSeconds: int64(
			lastState.ObservedAt.
				Sub(firstState.ObservedAt).
				Seconds(),
		),
		StartLatitude:  firstState.Latitude,
		StartLongitude: firstState.Longitude,
		EndLatitude:    lastState.Latitude,
		EndLongitude:   lastState.Longitude,
		PointCount:     len(inputs),
		SourceName: dominantSourceName(
			inputs,
		),
		CreatedAt: time.Now().UTC(),
	}
}

func segmentQualityScore(
	inputs []InputState,
) float64 {
	totalScore := 0.0

	for _, input := range inputs {
		totalScore += input.QualityScore
	}

	return trajectoryquality.SegmentScoreFromAggregate(
		totalScore,
		len(inputs),
	)
}

func copyAndSortInputs(
	inputs []InputState,
) []InputState {
	result := make(
		[]InputState,
		len(inputs),
	)

	copy(
		result,
		inputs,
	)

	sort.SliceStable(
		result,
		func(left int, right int) bool {
			leftState := result[left].State
			rightState := result[right].State

			if leftState.ICAO24 == rightState.ICAO24 {
				return leftState.ObservedAt.Before(
					rightState.ObservedAt,
				)
			}

			return leftState.ICAO24 <
				rightState.ICAO24
		},
	)

	return result
}

func groupInputsByAircraft(
	inputs []InputState,
) map[string][]InputState {
	groupedInputs := make(
		map[string][]InputState,
	)

	for _, input := range inputs {
		if input.State.ICAO24 == "" {
			continue
		}

		groupedInputs[input.State.ICAO24] = append(
			groupedInputs[input.State.ICAO24],
			input,
		)
	}

	return groupedInputs
}

func toTrackPoint4D(
	state flightstate.FlightState,
) trajectory.TrackPoint4D {
	return trajectory.TrackPoint4D{
		FlightStateID:       state.ID,
		FlightID:            state.FlightID,
		AircraftID:          state.AircraftID,
		ICAO24:              state.ICAO24,
		Callsign:            state.Callsign,
		Latitude:            state.Latitude,
		Longitude:           state.Longitude,
		BarometricAltitudeM: state.BarometricAltitudeM,
		BarometricAltitudeStatus: flightstate.ResolveAltitudeStatus(
			state.BarometricAltitudeM,
			state.BarometricAltitudeStatus,
		),
		GeometricAltitudeM: state.GeometricAltitudeM,
		GeometricAltitudeStatus: flightstate.ResolveAltitudeStatus(
			state.GeometricAltitudeM,
			state.GeometricAltitudeStatus,
		),
		VelocityMPS:     state.VelocityMPS,
		HeadingDegrees:  state.HeadingDegrees,
		VerticalRateMPS: state.VerticalRateMPS,
		OnGround:        state.OnGround,
		OriginCountry:   state.OriginCountry,
		ObservedAt:      state.ObservedAt,
		SourceName:      state.SourceName,
	}
}

func firstNonEmptyFlightID(
	inputs []InputState,
) string {
	for _, input := range inputs {
		if input.State.FlightID != "" {
			return input.State.FlightID
		}
	}

	return ""
}

func firstNonEmptyCallsign(
	inputs []InputState,
) string {
	for _, input := range inputs {
		if input.State.Callsign != "" {
			return input.State.Callsign
		}
	}

	return ""
}

func dominantSourceName(
	inputs []InputState,
) string {
	if len(inputs) == 0 {
		return ""
	}

	sourceName := inputs[0].State.SourceName

	for _, input := range inputs {
		if input.State.SourceName != sourceName {
			return "mixed"
		}
	}

	return sourceName
}
