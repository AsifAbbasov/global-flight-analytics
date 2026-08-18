package adsblol

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
)

type BarometricAltitudeKind string

const (
	BarometricAltitudeObserved    BarometricAltitudeKind = "observed"
	BarometricAltitudeGround      BarometricAltitudeKind = "ground"
	BarometricAltitudeUnavailable BarometricAltitudeKind = "unavailable"
	BarometricAltitudeInvalid     BarometricAltitudeKind = "invalid"
)

type BarometricAltitude struct {
	Feet float64
	Kind BarometricAltitudeKind
}

func (value *BarometricAltitude) UnmarshalJSON(data []byte) error {
	*value = BarometricAltitude{Kind: BarometricAltitudeInvalid}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		value.Kind = BarometricAltitudeUnavailable
		return nil
	}
	if trimmed[0] == '"' {
		var text string
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return err
		}
		if strings.EqualFold(strings.TrimSpace(text), "ground") {
			value.Kind = BarometricAltitudeGround
		}
		return nil
	}

	var feet float64
	if err := json.Unmarshal(trimmed, &feet); err != nil {
		return nil
	}
	if math.IsNaN(feet) || math.IsInf(feet, 0) {
		return nil
	}
	value.Feet = feet
	value.Kind = BarometricAltitudeObserved
	return nil
}

type StateResponse struct {
	Aircraft []AircraftItem `json:"ac"`
	Now      int64          `json:"now"`
	Total    int            `json:"total"`
}

type AircraftItem struct {
	Hex         string             `json:"hex"`
	Flight      *string            `json:"flight"`
	Latitude    *float64           `json:"lat"`
	Longitude   *float64           `json:"lon"`
	AltBaro     BarometricAltitude `json:"alt_baro"`
	AltGeom     *float64           `json:"alt_geom"`
	GroundSpeed *float64           `json:"gs"`
	Track       *float64           `json:"track"`
	BaroRate    *float64           `json:"baro_rate"`
	Seen        *float64           `json:"seen"`
	Squawk      *string            `json:"squawk"`
}
