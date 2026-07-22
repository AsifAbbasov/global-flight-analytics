package flightstate

import (
	"errors"
	"math"
)

var (
	ErrAltitudeStatusInvalid = errors.New("flight state altitude status is invalid")
	ErrAltitudeValueInvalid  = errors.New("flight state altitude value is invalid")
	ErrAltitudeStateConflict = errors.New("flight state altitude value conflicts with its status")
)

type Altitude struct {
	meters float64
	status AltitudeStatus
}

func NewAltitude(
	meters float64,
	status AltitudeStatus,
) (Altitude, error) {
	effectiveStatus := status
	if effectiveStatus == "" {
		switch {
		case math.IsNaN(meters) || math.IsInf(meters, 0):
			return Altitude{status: AltitudeStatusInvalid}, ErrAltitudeValueInvalid
		case meters == 0:
			effectiveStatus = AltitudeStatusUnavailable
		default:
			effectiveStatus = AltitudeStatusObserved
		}
	}

	if !IsKnownAltitudeStatus(effectiveStatus) {
		return Altitude{status: AltitudeStatusInvalid}, ErrAltitudeStatusInvalid
	}

	switch effectiveStatus {
	case AltitudeStatusObserved:
		if math.IsNaN(meters) || math.IsInf(meters, 0) {
			return Altitude{status: AltitudeStatusInvalid}, ErrAltitudeValueInvalid
		}
		return Altitude{
			meters: meters,
			status: AltitudeStatusObserved,
		}, nil
	case AltitudeStatusGround:
		return Altitude{
			meters: 0,
			status: AltitudeStatusGround,
		}, nil
	case AltitudeStatusUnknown,
		AltitudeStatusUnavailable:
		if !math.IsNaN(meters) && !math.IsInf(meters, 0) && meters != 0 {
			return Altitude{status: AltitudeStatusInvalid}, ErrAltitudeStateConflict
		}
		return Altitude{
			meters: 0,
			status: effectiveStatus,
		}, nil
	case AltitudeStatusInvalid:
		return Altitude{status: AltitudeStatusInvalid}, ErrAltitudeValueInvalid
	default:
		return Altitude{status: AltitudeStatusInvalid}, ErrAltitudeStatusInvalid
	}
}

func (value Altitude) Meters() float64 {
	return value.meters
}

func (value Altitude) Status() AltitudeStatus {
	return value.status
}

func (value Altitude) Available() bool {
	return value.status == AltitudeStatusObserved ||
		value.status == AltitudeStatusGround
}

func (value Altitude) Validate() error {
	if !IsKnownAltitudeStatus(value.status) {
		return ErrAltitudeStatusInvalid
	}
	if value.status == AltitudeStatusInvalid {
		return ErrAltitudeValueInvalid
	}
	if value.status == AltitudeStatusObserved &&
		(math.IsNaN(value.meters) || math.IsInf(value.meters, 0)) {
		return ErrAltitudeValueInvalid
	}
	if value.status != AltitudeStatusObserved &&
		value.status != AltitudeStatusGround &&
		value.meters != 0 {
		return ErrAltitudeValueInvalid
	}
	return nil
}

func (state FlightState) BarometricAltitude() (Altitude, error) {
	return NewAltitude(
		state.BarometricAltitudeM,
		state.BarometricAltitudeStatus,
	)
}

func (state FlightState) GeometricAltitude() (Altitude, error) {
	return NewAltitude(
		state.GeometricAltitudeM,
		state.GeometricAltitudeStatus,
	)
}
