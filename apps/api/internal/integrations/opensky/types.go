package opensky

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrStateVectorFieldCount = errors.New("OpenSky state vector has fewer than 17 fields")
	ErrStateVectorICAO24     = errors.New("OpenSky state vector ICAO24 is missing")
	ErrStateVectorType       = errors.New("OpenSky state vector field type is invalid")
)

type PositionSource int

const (
	PositionSourceADSB    PositionSource = 0
	PositionSourceASTERIX PositionSource = 1
	PositionSourceMLAT    PositionSource = 2
	PositionSourceFLARM   PositionSource = 3
)

type AircraftCategory int

const (
	AircraftCategoryNoInformation AircraftCategory = iota
	AircraftCategoryNoEmitterInformation
	AircraftCategoryLight
	AircraftCategorySmall
	AircraftCategoryLarge
	AircraftCategoryHighVortexLarge
	AircraftCategoryHeavy
	AircraftCategoryHighPerformance
	AircraftCategoryRotorcraft
	AircraftCategoryGlider
	AircraftCategoryLighterThanAir
	AircraftCategoryParachutist
	AircraftCategoryUltralight
	AircraftCategoryReserved
	AircraftCategoryUnmannedAerialVehicle
	AircraftCategorySpaceVehicle
	AircraftCategoryEmergencySurfaceVehicle
	AircraftCategoryServiceSurfaceVehicle
	AircraftCategoryPointObstacle
	AircraftCategoryClusterObstacle
	AircraftCategoryLineObstacle
)

type StateResponse struct {
	Time   int64             `json:"time"`
	States []json.RawMessage `json:"states"`
}

type StateVector struct {
	SnapshotTime      time.Time
	ICAO24            string
	Callsign          *string
	OriginCountry     string
	TimePosition      *time.Time
	LastContact       time.Time
	Longitude         *float64
	Latitude          *float64
	BaroAltitudeM     *float64
	OnGround          bool
	VelocityMPS       *float64
	TrueTrack         *float64
	VerticalRateMPS   *float64
	SensorSerials     []int64
	GeoAltitudeM      *float64
	Squawk            *string
	SPI               bool
	PositionSource    PositionSource
	Category          AircraftCategory
	CategoryAvailable bool
}

func ParseStateVector(raw json.RawMessage) (StateVector, error) {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return StateVector{}, fmt.Errorf("decode OpenSky state vector array: %w", err)
	}
	if len(values) < 17 {
		return StateVector{}, fmt.Errorf("%w: got %d", ErrStateVectorFieldCount, len(values))
	}

	icao24, err := requiredString(values[0], "icao24")
	if err != nil {
		return StateVector{}, err
	}
	icao24 = strings.ToLower(strings.TrimSpace(icao24))
	if icao24 == "" {
		return StateVector{}, ErrStateVectorICAO24
	}

	callsign, err := optionalTrimmedString(values[1], "callsign")
	if err != nil {
		return StateVector{}, err
	}
	originCountry, err := requiredString(values[2], "origin_country")
	if err != nil {
		return StateVector{}, err
	}
	timePositionUnix, err := optionalInt64(values[3], "time_position")
	if err != nil {
		return StateVector{}, err
	}
	lastContactUnix, err := requiredInt64(values[4], "last_contact")
	if err != nil {
		return StateVector{}, err
	}
	longitude, err := optionalFloat64(values[5], "longitude")
	if err != nil {
		return StateVector{}, err
	}
	latitude, err := optionalFloat64(values[6], "latitude")
	if err != nil {
		return StateVector{}, err
	}
	baroAltitude, err := optionalFloat64(values[7], "baro_altitude")
	if err != nil {
		return StateVector{}, err
	}
	onGround, err := requiredBool(values[8], "on_ground")
	if err != nil {
		return StateVector{}, err
	}
	velocity, err := optionalFloat64(values[9], "velocity")
	if err != nil {
		return StateVector{}, err
	}
	trueTrack, err := optionalFloat64(values[10], "true_track")
	if err != nil {
		return StateVector{}, err
	}
	verticalRate, err := optionalFloat64(values[11], "vertical_rate")
	if err != nil {
		return StateVector{}, err
	}
	sensors, err := optionalInt64Slice(values[12], "sensors")
	if err != nil {
		return StateVector{}, err
	}
	geoAltitude, err := optionalFloat64(values[13], "geo_altitude")
	if err != nil {
		return StateVector{}, err
	}
	squawk, err := optionalTrimmedString(values[14], "squawk")
	if err != nil {
		return StateVector{}, err
	}
	spi, err := requiredBool(values[15], "spi")
	if err != nil {
		return StateVector{}, err
	}
	positionSource, err := requiredInt64(values[16], "position_source")
	if err != nil {
		return StateVector{}, err
	}
	category := int64(AircraftCategoryNoInformation)
	categoryAvailable := false
	if len(values) >= 18 {
		category, err = requiredInt64(values[17], "category")
		if err != nil {
			return StateVector{}, err
		}
		categoryAvailable = true
	}

	var timePosition *time.Time
	if timePositionUnix != nil {
		value := time.Unix(*timePositionUnix, 0).UTC()
		timePosition = &value
	}

	return StateVector{
		ICAO24:            icao24,
		Callsign:          callsign,
		OriginCountry:     strings.TrimSpace(originCountry),
		TimePosition:      timePosition,
		LastContact:       time.Unix(lastContactUnix, 0).UTC(),
		Longitude:         longitude,
		Latitude:          latitude,
		BaroAltitudeM:     baroAltitude,
		OnGround:          onGround,
		VelocityMPS:       velocity,
		TrueTrack:         trueTrack,
		VerticalRateMPS:   verticalRate,
		SensorSerials:     sensors,
		GeoAltitudeM:      geoAltitude,
		Squawk:            squawk,
		SPI:               spi,
		PositionSource:    PositionSource(positionSource),
		Category:          AircraftCategory(category),
		CategoryAvailable: categoryAvailable,
	}, nil
}

func (response StateResponse) ParseStates() ([]StateVector, error) {
	states := make([]StateVector, 0, len(response.States))
	snapshotTime := time.Unix(response.Time, 0).UTC()
	for index, raw := range response.States {
		state, err := ParseStateVector(raw)
		if err != nil {
			return nil, fmt.Errorf("parse OpenSky state vector %d: %w", index, err)
		}
		state.SnapshotTime = snapshotTime
		states = append(states, state)
	}
	return states, nil
}

func isNull(raw json.RawMessage) bool {
	return strings.TrimSpace(string(raw)) == "null"
}

func requiredString(raw json.RawMessage, field string) (string, error) {
	if isNull(raw) {
		return "", fmt.Errorf("%w: %s is null", ErrStateVectorType, field)
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("%w: %s: %v", ErrStateVectorType, field, err)
	}
	return value, nil
}

func optionalTrimmedString(raw json.RawMessage, field string) (*string, error) {
	if isNull(raw) {
		return nil, nil
	}
	value, err := requiredString(raw, field)
	if err != nil {
		return nil, err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	return &value, nil
}

func requiredInt64(raw json.RawMessage, field string) (int64, error) {
	if isNull(raw) {
		return 0, fmt.Errorf("%w: %s is null", ErrStateVectorType, field)
	}
	var value int64
	if err := json.Unmarshal(raw, &value); err != nil {
		return 0, fmt.Errorf("%w: %s: %v", ErrStateVectorType, field, err)
	}
	return value, nil
}

func optionalInt64(raw json.RawMessage, field string) (*int64, error) {
	if isNull(raw) {
		return nil, nil
	}
	value, err := requiredInt64(raw, field)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func optionalFloat64(raw json.RawMessage, field string) (*float64, error) {
	if isNull(raw) {
		return nil, nil
	}
	var value float64
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrStateVectorType, field, err)
	}
	return &value, nil
}

func requiredBool(raw json.RawMessage, field string) (bool, error) {
	if isNull(raw) {
		return false, fmt.Errorf("%w: %s is null", ErrStateVectorType, field)
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return false, fmt.Errorf("%w: %s: %v", ErrStateVectorType, field, err)
	}
	return value, nil
}

func optionalInt64Slice(raw json.RawMessage, field string) ([]int64, error) {
	if isNull(raw) {
		return nil, nil
	}
	var value []int64
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrStateVectorType, field, err)
	}
	return value, nil
}
