package flightstate

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrSquawkCodeInvalid = errors.New(
		"flight state squawk code must be empty or contain exactly four octal digits",
	)
	ErrPositionSourceInvalid = errors.New(
		"flight state position source is invalid",
	)
	ErrAircraftCategoryInvalid = errors.New(
		"flight state aircraft category is invalid",
	)
)

type PositionSource string

const (
	PositionSourceUnknown PositionSource = ""
	PositionSourceADSB    PositionSource = "adsb"
	PositionSourceASTERIX PositionSource = "asterix"
	PositionSourceMLAT    PositionSource = "mlat"
	PositionSourceFLARM   PositionSource = "flarm"
)

const (
	MinimumAircraftCategory = 0
	MaximumAircraftCategory = 20
)

func NormalizeSquawkCode(
	value string,
) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", nil
	}
	if len(normalized) != 4 {
		return "", fmt.Errorf(
			"%w: %q",
			ErrSquawkCodeInvalid,
			value,
		)
	}
	for _, character := range normalized {
		if character < '0' || character > '7' {
			return "", fmt.Errorf(
				"%w: %q",
				ErrSquawkCodeInvalid,
				value,
			)
		}
	}
	return normalized, nil
}

func IsSpecialTransponderCode(
	value string,
) bool {
	normalized, err := NormalizeSquawkCode(value)
	if err != nil {
		return false
	}
	switch normalized {
	case "7500", "7600", "7700":
		return true
	default:
		return false
	}
}

func NormalizePositionSource(
	value PositionSource,
) (PositionSource, error) {
	normalized := PositionSource(
		strings.ToLower(
			strings.TrimSpace(
				string(value),
			),
		),
	)

	switch normalized {
	case PositionSourceUnknown,
		PositionSourceADSB,
		PositionSourceASTERIX,
		PositionSourceMLAT,
		PositionSourceFLARM:
		return normalized, nil
	default:
		return PositionSourceUnknown, fmt.Errorf(
			"%w: %q",
			ErrPositionSourceInvalid,
			value,
		)
	}
}

func ValidateObservationMetadata(
	state FlightState,
) error {
	if _, err := NormalizeSquawkCode(
		state.SquawkCode,
	); err != nil {
		return err
	}
	if _, err := NormalizePositionSource(
		state.PositionSource,
	); err != nil {
		return err
	}
	_, err := state.ResolveAircraftCategory()
	return err
}
