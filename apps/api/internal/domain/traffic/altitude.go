package traffic

import (
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

type AltitudeSource string

const (
	AltitudeSourceGeometric  AltitudeSource = "geometric"
	AltitudeSourceBarometric AltitudeSource = "barometric"
	AltitudeSourceGround     AltitudeSource = "ground"
	AltitudeSourceNone       AltitudeSource = "none"
)

// ResolveCurrentAltitude selects the best altitude evidence for the current
// traffic view without using zero as a missing-value sentinel.
//
// Selection order:
//  1. on-ground state, represented explicitly as ground at zero metres;
//  2. observed geometric altitude, including a legitimate observed zero;
//  3. observed barometric altitude;
//  4. no numeric value, with the strongest absence or invalidity status.
func ResolveCurrentAltitude(
	onGround bool,
	geometricValue *float64,
	geometricStatus flightstate.AltitudeStatus,
	barometricValue *float64,
	barometricStatus flightstate.AltitudeStatus,
) (
	*float64,
	flightstate.AltitudeStatus,
	AltitudeSource,
) {
	if onGround {
		value := 0.0

		return &value,
			flightstate.AltitudeStatusGround,
			AltitudeSourceGround
	}

	normalizedGeometricStatus := normalizeCurrentAltitudeCandidate(
		geometricValue,
		geometricStatus,
	)
	if normalizedGeometricStatus == flightstate.AltitudeStatusObserved {
		return copyAltitudeValue(geometricValue),
			flightstate.AltitudeStatusObserved,
			AltitudeSourceGeometric
	}

	normalizedBarometricStatus := normalizeCurrentAltitudeCandidate(
		barometricValue,
		barometricStatus,
	)
	if normalizedBarometricStatus == flightstate.AltitudeStatusObserved {
		return copyAltitudeValue(barometricValue),
			flightstate.AltitudeStatusObserved,
			AltitudeSourceBarometric
	}

	return nil,
		mergeUnavailableAltitudeStatuses(
			normalizedGeometricStatus,
			normalizedBarometricStatus,
		),
		AltitudeSourceNone
}

func normalizeCurrentAltitudeCandidate(
	value *float64,
	status flightstate.AltitudeStatus,
) flightstate.AltitudeStatus {
	effectiveStatus := status
	if effectiveStatus == "" {
		if value == nil {
			return flightstate.AltitudeStatusUnavailable
		}

		effectiveStatus = flightstate.ResolveAltitudeStatus(
			*value,
			status,
		)
	}

	if !flightstate.IsKnownAltitudeStatus(effectiveStatus) {
		return flightstate.AltitudeStatusInvalid
	}

	switch effectiveStatus {
	case flightstate.AltitudeStatusObserved:
		if value == nil ||
			math.IsNaN(*value) ||
			math.IsInf(*value, 0) {
			return flightstate.AltitudeStatusInvalid
		}

		return flightstate.AltitudeStatusObserved

	case flightstate.AltitudeStatusGround:
		// Ground is valid only when the independent on_ground evidence is true.
		// ResolveCurrentAltitude handles that state before candidate selection.
		return flightstate.AltitudeStatusInvalid

	case flightstate.AltitudeStatusUnknown,
		flightstate.AltitudeStatusUnavailable,
		flightstate.AltitudeStatusInvalid:
		if value != nil {
			return flightstate.AltitudeStatusInvalid
		}

		return effectiveStatus

	default:
		return flightstate.AltitudeStatusInvalid
	}
}

func mergeUnavailableAltitudeStatuses(
	first flightstate.AltitudeStatus,
	second flightstate.AltitudeStatus,
) flightstate.AltitudeStatus {
	for _, status := range []flightstate.AltitudeStatus{
		first,
		second,
	} {
		if status == flightstate.AltitudeStatusInvalid {
			return flightstate.AltitudeStatusInvalid
		}
	}

	for _, status := range []flightstate.AltitudeStatus{
		first,
		second,
	} {
		if status == flightstate.AltitudeStatusUnknown {
			return flightstate.AltitudeStatusUnknown
		}
	}

	return flightstate.AltitudeStatusUnavailable
}

func copyAltitudeValue(value *float64) *float64 {
	if value == nil {
		return nil
	}

	result := *value

	return &result
}
