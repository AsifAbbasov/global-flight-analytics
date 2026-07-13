package flightcontinuation

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const identityKeyPrefix = "flight-identity-"

var ErrMaxGapInvalid = errors.New(
	"flight identity continuation maximum gap must be non-negative",
)

type Config struct {
	MaxGap time.Duration
}

func (config Config) Validate() error {
	if config.MaxGap < 0 {
		return fmt.Errorf(
			"%w, got %s",
			ErrMaxGapInvalid,
			config.MaxGap,
		)
	}

	return nil
}

func Continue(
	previous trajectory.FlightTrajectory,
	current trajectory.FlightTrajectory,
	config Config,
) (trajectory.FlightTrajectory, bool) {
	if config.MaxGap <= 0 {
		return current, false
	}

	if !sameICAO24(previous.ICAO24, current.ICAO24) {
		return current, false
	}

	if !hasCompleteIdentity(previous) ||
		!hasCompleteIdentity(current) {
		return current, false
	}

	if current.SplitReason !=
		trajectory.FlightSplitReasonInitialObservation {
		return current, false
	}

	if previous.EndTime.IsZero() ||
		current.StartTime.IsZero() ||
		current.StartTime.Before(previous.EndTime) {
		return current, false
	}

	if current.StartTime.Sub(previous.EndTime) >
		config.MaxGap {
		return current, false
	}

	previousSourceFlightID := normalizeUUID(
		previous.FlightID,
	)
	currentSourceFlightID := normalizeUUID(
		current.FlightID,
	)

	if previousSourceFlightID != "" ||
		currentSourceFlightID != "" {
		if previousSourceFlightID == "" ||
			currentSourceFlightID == "" ||
			previousSourceFlightID != currentSourceFlightID {
			return current, false
		}

		return continuedTrajectory(
			previous,
			current,
		), true
	}

	if previous.IdentityBasis !=
		trajectory.FlightIdentityBasisCallsignAndStartTime ||
		current.IdentityBasis !=
			trajectory.FlightIdentityBasisCallsignAndStartTime {
		return current, false
	}

	previousCallsign := normalizeCallsign(
		previous.Callsign,
	)
	currentCallsign := normalizeCallsign(
		current.Callsign,
	)

	if previousCallsign == "" ||
		currentCallsign == "" ||
		previousCallsign != currentCallsign {
		return current, false
	}

	return continuedTrajectory(
		previous,
		current,
	), true
}

func continuedTrajectory(
	previous trajectory.FlightTrajectory,
	current trajectory.FlightTrajectory,
) trajectory.FlightTrajectory {
	result := current
	result.IdentityKey = previous.IdentityKey
	result.IdentityBasis = previous.IdentityBasis
	result.SplitReason =
		trajectory.FlightSplitReasonContinuedFromPreviousBatch

	return result
}

func hasCompleteIdentity(
	item trajectory.FlightTrajectory,
) bool {
	return isIdentityKey(item.IdentityKey) &&
		isKnownIdentityBasis(item.IdentityBasis) &&
		isKnownSplitReason(item.SplitReason)
}

func isIdentityKey(value string) bool {
	if !strings.HasPrefix(
		value,
		identityKeyPrefix,
	) {
		return false
	}

	digest := strings.TrimPrefix(
		value,
		identityKeyPrefix,
	)

	if len(digest) != 64 ||
		digest != strings.ToLower(digest) {
		return false
	}

	decoded, err := hex.DecodeString(digest)

	return err == nil && len(decoded) == 32
}

func isKnownIdentityBasis(
	value trajectory.FlightIdentityBasis,
) bool {
	switch value {
	case trajectory.FlightIdentityBasisSourceFlightID,
		trajectory.FlightIdentityBasisCallsignAndStartTime,
		trajectory.FlightIdentityBasisAircraftAndStartTime:
		return true

	default:
		return false
	}
}

func isKnownSplitReason(
	value trajectory.FlightSplitReason,
) bool {
	switch value {
	case trajectory.FlightSplitReasonInitialObservation,
		trajectory.FlightSplitReasonSourceFlightIDChanged,
		trajectory.FlightSplitReasonCallsignChanged,
		trajectory.FlightSplitReasonGroundCycle,
		trajectory.FlightSplitReasonContinuedFromPreviousBatch:
		return true

	default:
		return false
	}
}

func sameICAO24(
	left string,
	right string,
) bool {
	normalizedLeft := strings.ToUpper(
		strings.TrimSpace(left),
	)
	normalizedRight := strings.ToUpper(
		strings.TrimSpace(right),
	)

	return normalizedLeft != "" &&
		normalizedLeft == normalizedRight
}

func normalizeCallsign(value string) string {
	return strings.ToUpper(
		strings.TrimSpace(value),
	)
}

func normalizeUUID(value string) string {
	normalized := strings.ToLower(
		strings.TrimSpace(value),
	)

	if len(normalized) != 36 {
		return ""
	}

	for index, character := range normalized {
		switch index {
		case 8, 13, 18, 23:
			if character != '-' {
				return ""
			}

		default:
			if character < '0' ||
				character > '9' &&
					(character < 'a' ||
						character > 'f') {
				return ""
			}
		}
	}

	return normalized
}
