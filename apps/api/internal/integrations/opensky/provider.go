package opensky

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

const (
	sourceName                        = "opensky"
	maximumRegionalRadiusNauticalMile = 250
	earthRadiusNauticalMile           = 3440.065
	maximumRegionalStateCreditCost    = 3
)

var (
	ErrStatesClientRequired = errors.New(
		"OpenSky states client is required",
	)
	ErrRegionalRadiusInvalid = errors.New(
		"OpenSky regional radius must be between one and 250 nautical miles",
	)
	ErrRegionalLatitudeInvalid = errors.New(
		"OpenSky regional latitude is invalid",
	)
	ErrRegionalLongitudeInvalid = errors.New(
		"OpenSky regional longitude is invalid",
	)
	ErrRegionalBoundingBoxCrossesDateline = errors.New(
		"OpenSky regional bounding box crosses the international date line",
	)
	ErrRegionalBoundingBoxCrossesPole = errors.New(
		"OpenSky regional bounding box crosses a geographic pole",
	)
	ErrRegionalBoundingBoxTooExpensive = errors.New(
		"OpenSky regional bounding box exceeds the configured free-credit boundary",
	)
)

type StatesClient interface {
	GetStates(
		ctx context.Context,
		input StatesRequest,
	) (StatesResult, error)
}

type Provider struct {
	client StatesClient
}

func NewProvider(client StatesClient) (*Provider, error) {
	if client == nil {
		return nil, ErrStatesClientRequired
	}

	return &Provider{client: client}, nil
}

func (provider *Provider) SourceName() string {
	return sourceName
}

func (provider *Provider) LoadByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) ([]flightstate.FlightState, error) {
	box, err := RegionalBoundingBox(
		latitude,
		longitude,
		radius,
	)
	if err != nil {
		return nil, err
	}

	result, err := provider.client.GetStates(
		ctx,
		StatesRequest{BoundingBox: &box, Extended: true},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"load OpenSky regional states: %w",
			err,
		)
	}

	states := make(
		[]flightstate.FlightState,
		0,
		len(result.States),
	)

	for index := range result.States {
		mapped, usable, mapErr := MapStateVector(
			result.States[index],
		)
		if mapErr != nil {
			return nil, fmt.Errorf(
				"map OpenSky state vector %d: %w",
				index,
				mapErr,
			)
		}
		if !usable {
			continue
		}

		states = append(states, mapped)
	}

	return states, nil
}

func RegionalBoundingBox(
	latitude float64,
	longitude float64,
	radiusNauticalMiles int,
) (BoundingBox, error) {
	if math.IsNaN(latitude) || math.IsInf(latitude, 0) ||
		latitude < -90 || latitude > 90 {
		return BoundingBox{}, ErrRegionalLatitudeInvalid
	}
	if math.IsNaN(longitude) || math.IsInf(longitude, 0) ||
		longitude < -180 || longitude > 180 {
		return BoundingBox{}, ErrRegionalLongitudeInvalid
	}
	if radiusNauticalMiles <= 0 ||
		radiusNauticalMiles > maximumRegionalRadiusNauticalMile {
		return BoundingBox{}, ErrRegionalRadiusInvalid
	}

	angularDistance := float64(radiusNauticalMiles) /
		earthRadiusNauticalMile
	latitudeRadians := latitude * math.Pi / 180
	latitudeDelta := angularDistance * 180 / math.Pi

	minimumLatitude := latitude - latitudeDelta
	maximumLatitude := latitude + latitudeDelta
	if minimumLatitude <= -90 || maximumLatitude >= 90 {
		return BoundingBox{}, ErrRegionalBoundingBoxCrossesPole
	}

	longitudeDenominator := math.Cos(latitudeRadians)
	if math.Abs(longitudeDenominator) < 1e-12 {
		return BoundingBox{}, ErrRegionalBoundingBoxCrossesPole
	}

	ratio := math.Sin(angularDistance) / longitudeDenominator
	if ratio < -1 || ratio > 1 {
		return BoundingBox{}, ErrRegionalBoundingBoxCrossesPole
	}
	longitudeDelta := math.Asin(ratio) * 180 / math.Pi
	minimumLongitude := longitude - longitudeDelta
	maximumLongitude := longitude + longitudeDelta
	if minimumLongitude < -180 || maximumLongitude > 180 {
		return BoundingBox{}, ErrRegionalBoundingBoxCrossesDateline
	}

	box := BoundingBox{
		MinimumLatitude:  minimumLatitude,
		MaximumLatitude:  maximumLatitude,
		MinimumLongitude: minimumLongitude,
		MaximumLongitude: maximumLongitude,
	}
	if err := box.Validate(); err != nil {
		return BoundingBox{}, err
	}

	creditCost, err := box.EstimatedStateCreditCost()
	if err != nil {
		return BoundingBox{}, err
	}
	if creditCost > maximumRegionalStateCreditCost {
		return BoundingBox{}, fmt.Errorf(
			"%w: estimated_credit_cost=%d",
			ErrRegionalBoundingBoxTooExpensive,
			creditCost,
		)
	}

	return box, nil
}

func MapStateVector(
	state StateVector,
) (flightstate.FlightState, bool, error) {
	validity, err := EvaluateStateVectorValidity(state)
	if err != nil {
		return flightstate.FlightState{}, false, err
	}
	if !validity.PositionUsable ||
		!validity.LastContactWithinWindow ||
		state.Latitude == nil ||
		state.Longitude == nil ||
		state.TimePosition == nil {
		return flightstate.FlightState{}, false, nil
	}

	barometricAltitude, barometricStatus := altitudeReading(
		state.BaroAltitudeM,
		state.OnGround,
	)
	geometricAltitude, geometricStatus := altitudeReading(
		state.GeoAltitudeM,
		false,
	)
	velocity, velocityAvailable :=
		optionalFiniteFloat64(
			state.VelocityMPS,
		)
	heading, headingAvailable :=
		optionalFiniteFloat64(
			state.TrueTrack,
		)
	verticalRate, verticalRateAvailable :=
		optionalFiniteFloat64(
			state.VerticalRateMPS,
		)

	mapped := flightstate.FlightState{
		ICAO24:                     strings.ToUpper(state.ICAO24),
		Latitude:                   *state.Latitude,
		Longitude:                  *state.Longitude,
		BarometricAltitudeM:        barometricAltitude,
		BarometricAltitudeStatus:   barometricStatus,
		GeometricAltitudeM:         geometricAltitude,
		GeometricAltitudeStatus:    geometricStatus,
		VelocityMPS:                velocity,
		VelocityAvailable:          velocityAvailable,
		HeadingDegrees:             heading,
		HeadingAvailable:           headingAvailable,
		VerticalRateMPS:            verticalRate,
		VerticalRateAvailable:      verticalRateAvailable,
		OnGround:                   state.OnGround,
		OnGroundAvailable:          true,
		TelemetryAvailabilityKnown: true,
		OriginCountry:              strings.TrimSpace(state.OriginCountry),
		SquawkCode:                 optionalTrimmedStringValue(state.Squawk),
		SpecialPurposeIndicator:    state.SPI,
		PositionSource:             canonicalPositionSource(state.PositionSource),
		AircraftCategory:           int(state.Category),
		AircraftCategoryAvailable:  state.CategoryAvailable,
		ObservedAt:                 state.TimePosition.UTC(),
		SourceName:                 sourceName,
	}
	if state.Callsign != nil {
		mapped.Callsign = strings.TrimSpace(*state.Callsign)
	}

	return mapped, true, nil
}

func altitudeReading(
	value *float64,
	onGround bool,
) (float64, flightstate.AltitudeStatus) {
	if onGround {
		return 0, flightstate.AltitudeStatusGround
	}
	if value == nil {
		return 0, flightstate.AltitudeStatusUnavailable
	}
	if math.IsNaN(*value) || math.IsInf(*value, 0) {
		return 0, flightstate.AltitudeStatusInvalid
	}

	return *value, flightstate.AltitudeStatusObserved
}

func optionalFiniteFloat64(
	value *float64,
) (float64, bool) {
	if value == nil ||
		math.IsNaN(*value) ||
		math.IsInf(*value, 0) {
		return 0, false
	}

	return *value, true
}

func optionalTrimmedStringValue(
	value *string,
) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func canonicalPositionSource(
	value PositionSource,
) flightstate.PositionSource {
	switch value {
	case PositionSourceADSB:
		return flightstate.PositionSourceADSB
	case PositionSourceASTERIX:
		return flightstate.PositionSourceASTERIX
	case PositionSourceMLAT:
		return flightstate.PositionSourceMLAT
	case PositionSourceFLARM:
		return flightstate.PositionSourceFLARM
	default:
		return flightstate.PositionSourceUnknown
	}
}

// OPEN-AVIATION-RESEARCH-EVIDENCE-V1
