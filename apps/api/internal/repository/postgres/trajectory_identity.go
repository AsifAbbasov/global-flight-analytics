package postgres

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

const persistedFlightIdentityKeyPrefix = "flight-identity-"

var (
	errFlightIdentityMetadataIncomplete = errors.New("flight identity metadata must be either entirely empty or entirely populated")
	errFlightIdentityKeyInvalid         = errors.New("flight identity key is invalid")
	errFlightIdentityBasisInvalid       = errors.New("flight identity basis is invalid")
	errFlightSplitReasonInvalid         = errors.New("flight split reason is invalid")
)

func validatePersistedFlightIdentity(
	item trajectory.FlightTrajectory,
) error {
	identityKey := strings.TrimSpace(item.IdentityKey)
	identityBasis := strings.TrimSpace(string(item.IdentityBasis))
	splitReason := strings.TrimSpace(string(item.SplitReason))

	if identityKey == "" && identityBasis == "" && splitReason == "" {
		return nil
	}

	if identityKey == "" || identityBasis == "" || splitReason == "" {
		return errFlightIdentityMetadataIncomplete
	}

	if identityKey != item.IdentityKey || !isPersistedFlightIdentityKey(identityKey) {
		return fmt.Errorf(
			"%w: %q",
			errFlightIdentityKeyInvalid,
			item.IdentityKey,
		)
	}

	if !isKnownFlightIdentityBasis(item.IdentityBasis) {
		return fmt.Errorf(
			"%w: %q",
			errFlightIdentityBasisInvalid,
			item.IdentityBasis,
		)
	}

	if !isKnownFlightSplitReason(item.SplitReason) {
		return fmt.Errorf(
			"%w: %q",
			errFlightSplitReasonInvalid,
			item.SplitReason,
		)
	}

	return nil
}

func isPersistedFlightIdentityKey(value string) bool {
	if !strings.HasPrefix(value, persistedFlightIdentityKeyPrefix) {
		return false
	}

	digest := strings.TrimPrefix(value, persistedFlightIdentityKeyPrefix)
	if len(digest) != 64 || digest != strings.ToLower(digest) {
		return false
	}

	decoded, err := hex.DecodeString(digest)
	return err == nil && len(decoded) == 32
}

func isKnownFlightIdentityBasis(
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

func isKnownFlightSplitReason(
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
