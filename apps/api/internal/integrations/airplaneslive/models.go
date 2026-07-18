package airplaneslive

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
)

type StateResponse struct {
	Now      float64        `json:"now"`
	Messages int            `json:"messages"`
	Total    int            `json:"total"`
	Aircraft []AircraftItem `json:"ac"`
}

type BarometricAltitudeKind string

const (
	BarometricAltitudeKindObserved    BarometricAltitudeKind = "observed"
	BarometricAltitudeKindGround      BarometricAltitudeKind = "ground"
	BarometricAltitudeKindUnknown     BarometricAltitudeKind = "unknown"
	BarometricAltitudeKindUnavailable BarometricAltitudeKind = "unavailable"
	BarometricAltitudeKindInvalid     BarometricAltitudeKind = "invalid"
)

type BarometricAltitude struct {
	Feet float64
	Kind BarometricAltitudeKind
}

func (value *BarometricAltitude) UnmarshalJSON(
	data []byte,
) error {
	trimmed := bytes.TrimSpace(data)

	*value = BarometricAltitude{
		Kind: BarometricAltitudeKindInvalid,
	}

	if len(trimmed) == 0 {
		return nil
	}

	if bytes.Equal(
		trimmed,
		[]byte("null"),
	) {
		value.Kind = BarometricAltitudeKindUnavailable

		return nil
	}

	if trimmed[0] == '"' {
		var text string

		if err := json.Unmarshal(
			trimmed,
			&text,
		); err != nil {
			return err
		}

		switch strings.ToLower(
			strings.TrimSpace(text),
		) {
		case "ground":
			value.Kind = BarometricAltitudeKindGround

		case "", "unknown":
			value.Kind = BarometricAltitudeKindUnknown

		default:
			value.Kind = BarometricAltitudeKindInvalid
		}

		return nil
	}

	var feet float64

	if err := json.Unmarshal(
		trimmed,
		&feet,
	); err != nil {
		value.Kind = BarometricAltitudeKindInvalid

		return nil
	}

	if math.IsNaN(feet) ||
		math.IsInf(feet, 0) {
		value.Kind = BarometricAltitudeKindInvalid

		return nil
	}

	value.Feet = feet
	value.Kind = BarometricAltitudeKindObserved

	return nil
}

type AircraftItem struct {
	Hex          string             `json:"hex"`
	Flight       string             `json:"flight"`
	Latitude     float64            `json:"lat"`
	Longitude    float64            `json:"lon"`
	AltBaro      BarometricAltitude `json:"alt_baro"`
	AltGeom      *float64           `json:"alt_geom"`
	GroundSpeed  float64            `json:"gs"`
	Track        float64            `json:"track"`
	BaroRate     float64            `json:"baro_rate"`
	Seen         float64            `json:"seen"`
	Type         string             `json:"type"`
	Registration string             `json:"r"`
	AircraftType string             `json:"t"`
	Squawk       string             `json:"squawk"`
}
