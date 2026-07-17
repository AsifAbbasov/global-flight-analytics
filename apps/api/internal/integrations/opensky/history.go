package opensky

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrICAO24Required        = errors.New("OpenSky ICAO24 is required")
	ErrAirportICAORequired   = errors.New("OpenSky airport ICAO code is required")
	ErrTimeRangeInvalid      = errors.New("OpenSky time range is invalid")
	ErrFlightRangeTooLarge   = errors.New("OpenSky flight interval exceeds the supported two-hour range")
	ErrAircraftRangeTooLarge = errors.New("OpenSky aircraft flight interval exceeds the supported two-day range")
	ErrTrackTooOld           = errors.New("OpenSky track request exceeds the supported thirty-day retention")
)

type FlightData struct {
	ICAO24                          string  `json:"icao24"`
	FirstSeen                       int64   `json:"firstSeen"`
	EstimatedDepartureAirport       *string `json:"estDepartureAirport"`
	LastSeen                        int64   `json:"lastSeen"`
	EstimatedArrivalAirport         *string `json:"estArrivalAirport"`
	Callsign                        *string `json:"callsign"`
	EstimatedDepartureHorizontalM   *int64  `json:"estDepartureAirportHorizDistance"`
	EstimatedDepartureVerticalM     *int64  `json:"estDepartureAirportVertDistance"`
	EstimatedArrivalHorizontalM     *int64  `json:"estArrivalAirportHorizDistance"`
	EstimatedArrivalVerticalM       *int64  `json:"estArrivalAirportVertDistance"`
	DepartureAirportCandidatesCount *int64  `json:"departureAirportCandidatesCount"`
	ArrivalAirportCandidatesCount   *int64  `json:"arrivalAirportCandidatesCount"`
}

type TimeRange struct {
	Begin time.Time
	End   time.Time
}

func (value TimeRange) Validate() error {
	if value.Begin.IsZero() || value.End.IsZero() || !value.End.After(value.Begin) {
		return ErrTimeRangeInvalid
	}
	return nil
}

func ValidateAllFlightsRange(value TimeRange) error {
	if err := value.Validate(); err != nil {
		return err
	}
	if value.End.Sub(value.Begin) > 2*time.Hour {
		return ErrFlightRangeTooLarge
	}
	return nil
}

func ValidateAircraftFlightsRange(value TimeRange) error {
	if err := value.Validate(); err != nil {
		return err
	}
	if value.End.Sub(value.Begin) > 48*time.Hour {
		return ErrAircraftRangeTooLarge
	}
	return nil
}

func NormalizeICAO24(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", ErrICAO24Required
	}
	return value, nil
}

func NormalizeAirportICAO(value string) (string, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "", ErrAirportICAORequired
	}
	return value, nil
}

func ValidateTrackTime(value time.Time, now time.Time) error {
	if value.IsZero() {
		return nil
	}
	if value.After(now.UTC()) {
		return ErrRequestTimeInFuture
	}
	if value.Before(now.UTC().Add(-30 * 24 * time.Hour)) {
		return ErrTrackTooOld
	}
	return nil
}

func EstimatedAirportDisclosure(flight FlightData) []string {
	labels := []string{
		"OpenSky airport fields are estimates, not official airport operations data.",
		"Candidate counts and provider distances must remain visible when available.",
	}
	if flight.EstimatedDepartureAirport == nil && flight.EstimatedArrivalAirport == nil {
		labels = append(labels, "No estimated airport was reported.")
	}
	return labels
}

func FlightWindowDescription(value TimeRange) string {
	return fmt.Sprintf(
		"%s/%s",
		value.Begin.UTC().Format(time.RFC3339),
		value.End.UTC().Format(time.RFC3339),
	)
}
