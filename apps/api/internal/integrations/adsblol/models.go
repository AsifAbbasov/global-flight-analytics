package adsblol

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/readsbcompat"

type StateResponse = readsbcompat.StateResponse

type BarometricAltitudeKind = readsbcompat.BarometricAltitudeKind

const (
	BarometricAltitudeKindObserved    = readsbcompat.BarometricAltitudeKindObserved
	BarometricAltitudeKindGround      = readsbcompat.BarometricAltitudeKindGround
	BarometricAltitudeKindUnknown     = readsbcompat.BarometricAltitudeKindUnknown
	BarometricAltitudeKindUnavailable = readsbcompat.BarometricAltitudeKindUnavailable
	BarometricAltitudeKindInvalid     = readsbcompat.BarometricAltitudeKindInvalid
)

type BarometricAltitude = readsbcompat.BarometricAltitude

type OptionalFloat64 = readsbcompat.OptionalFloat64

type AircraftItem = readsbcompat.AircraftItem
