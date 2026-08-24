package airplaneslive

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/readsbcompat"

type StateResponse struct {
	Now      float64        `json:"now"`
	Messages int            `json:"messages"`
	Total    int            `json:"total"`
	Aircraft []AircraftItem `json:"ac"`
}

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
