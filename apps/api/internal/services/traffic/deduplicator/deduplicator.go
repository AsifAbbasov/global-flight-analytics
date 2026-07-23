package deduplicator

import (
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

type Result struct {
	UniqueStates   []flightstate.FlightState
	DuplicateCount int
}

type strictPointKey struct {
	ICAO24                     string
	Callsign                   string
	ObservedAtUnixNano         int64
	ObservedAtIsZero           bool
	LatitudeBits               uint64
	LongitudeBits              uint64
	BarometricAltitudeBits     uint64
	BarometricAltitudeStatus   flightstate.AltitudeStatus
	GeometricAltitudeBits      uint64
	GeometricAltitudeStatus    flightstate.AltitudeStatus
	VelocityBits               uint64
	VelocityAvailable          bool
	HeadingBits                uint64
	HeadingAvailable           bool
	VerticalRateBits           uint64
	VerticalRateAvailable      bool
	OnGround                   bool
	OnGroundAvailable          bool
	TelemetryAvailabilityKnown bool
	OriginCountry              string
	SquawkCode                 string
	SpecialPurposeIndicator    bool
	PositionSource             flightstate.PositionSource
	AircraftCategory           int
	AircraftCategoryAvailable  bool
	SourceName                 string
}

func RemoveExactDuplicates(
	states []flightstate.FlightState,
) Result {
	uniqueStates := make(
		[]flightstate.FlightState,
		0,
		len(states),
	)

	seen := make(
		map[strictPointKey]struct{},
		len(states),
	)

	duplicateCount := 0

	for _, state := range states {
		key := makeStrictPointKey(
			state,
		)

		if _, exists := seen[key]; exists {
			duplicateCount++

			continue
		}

		seen[key] = struct{}{}

		uniqueStates = append(
			uniqueStates,
			state,
		)
	}

	return Result{
		UniqueStates:   uniqueStates,
		DuplicateCount: duplicateCount,
	}
}

func makeStrictPointKey(
	state flightstate.FlightState,
) strictPointKey {
	observedAtUnixNano := int64(0)
	observedAtIsZero := state.ObservedAt.IsZero()

	if !observedAtIsZero {
		observedAtUnixNano = state.ObservedAt.
			UTC().
			UnixNano()
	}

	return strictPointKey{
		ICAO24:             state.ICAO24,
		Callsign:           state.Callsign,
		ObservedAtUnixNano: observedAtUnixNano,
		ObservedAtIsZero:   observedAtIsZero,
		LatitudeBits: canonicalFloatBits(
			state.Latitude,
		),
		LongitudeBits: canonicalFloatBits(
			state.Longitude,
		),
		BarometricAltitudeBits: canonicalFloatBits(
			state.BarometricAltitudeM,
		),
		BarometricAltitudeStatus: flightstate.ResolveAltitudeStatus(
			state.BarometricAltitudeM,
			state.BarometricAltitudeStatus,
		),
		GeometricAltitudeBits: canonicalFloatBits(
			state.GeometricAltitudeM,
		),
		GeometricAltitudeStatus: flightstate.ResolveAltitudeStatus(
			state.GeometricAltitudeM,
			state.GeometricAltitudeStatus,
		),
		VelocityBits: canonicalFloatBits(
			state.VelocityMPS,
		),
		VelocityAvailable: state.VelocityAvailable,
		HeadingBits: canonicalFloatBits(
			state.HeadingDegrees,
		),
		HeadingAvailable: state.HeadingAvailable,
		VerticalRateBits: canonicalFloatBits(
			state.VerticalRateMPS,
		),
		VerticalRateAvailable:      state.VerticalRateAvailable,
		OnGround:                   state.OnGround,
		OnGroundAvailable:          state.OnGroundAvailable,
		TelemetryAvailabilityKnown: state.TelemetryAvailabilityKnown,
		OriginCountry:              state.OriginCountry,
		SquawkCode:                 state.SquawkCode,
		SpecialPurposeIndicator:    state.SpecialPurposeIndicator,
		PositionSource:             state.PositionSource,
		AircraftCategory:           state.AircraftCategory,
		AircraftCategoryAvailable:  state.AircraftCategoryAvailable,
		SourceName:                 state.SourceName,
	}
}

func canonicalFloatBits(
	value float64,
) uint64 {
	if value == 0 {
		return 0
	}

	return math.Float64bits(
		value,
	)
}
