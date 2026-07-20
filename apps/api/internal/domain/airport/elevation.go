package airport

import "math"

type ElevationStatus string

const (
	ElevationStatusObserved ElevationStatus = "observed"
	ElevationStatusUnknown  ElevationStatus = "unknown"
	ElevationStatusInvalid  ElevationStatus = "invalid"
)

func ResolveElevation(
	value float64,
	available bool,
) (
	float64,
	ElevationStatus,
	bool,
) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, ElevationStatusInvalid, false
	}

	// Preserve compatibility with legacy in-memory fixtures that supplied a
	// non-zero value before availability became explicit. A real zero requires
	// explicit availability so unknown elevation is never inferred as sea level.
	if available || value != 0 {
		if value == 0 {
			value = 0
		}

		return value, ElevationStatusObserved, true
	}

	return 0, ElevationStatusUnknown, false
}
