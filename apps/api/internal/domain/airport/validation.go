package airport

import (
	"errors"
	"math"
	"strings"
)

var (
	ErrAirportICAORequired       = errors.New("airport ICAO code is required")
	ErrAirportCoordinatesInvalid = errors.New("airport coordinates are invalid")
	ErrAirportElevationInvalid   = errors.New("airport elevation is invalid")
)

func (value Airport) Validate() error {
	if strings.TrimSpace(value.ICAOCode) == "" {
		return ErrAirportICAORequired
	}
	if !isFiniteAirportValue(value.Latitude) || value.Latitude < -90 || value.Latitude > 90 ||
		!isFiniteAirportValue(value.Longitude) || value.Longitude < -180 || value.Longitude > 180 {
		return ErrAirportCoordinatesInvalid
	}
	if value.ElevationAvailable && !isFiniteAirportValue(value.ElevationM) {
		return ErrAirportElevationInvalid
	}
	return nil
}

func isFiniteAirportValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
