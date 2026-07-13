package postgres

import (
	"errors"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestValidatePersistedFlightIdentityAllowsLegacyEmptyMetadata(
	t *testing.T,
) {
	if err := validatePersistedFlightIdentity(
		trajectory.FlightTrajectory{},
	); err != nil {
		t.Fatalf("expected legacy empty identity metadata to be accepted, got %v", err)
	}
}

func TestValidatePersistedFlightIdentityAcceptsCompleteMetadata(
	t *testing.T,
) {
	item := validPersistedIdentityTrajectory()

	if err := validatePersistedFlightIdentity(item); err != nil {
		t.Fatalf("expected complete identity metadata to be accepted, got %v", err)
	}
}

func TestValidatePersistedFlightIdentityRejectsIncompleteMetadata(
	t *testing.T,
) {
	item := validPersistedIdentityTrajectory()
	item.SplitReason = ""

	err := validatePersistedFlightIdentity(item)
	if !errors.Is(err, errFlightIdentityMetadataIncomplete) {
		t.Fatalf("expected incomplete metadata error, got %v", err)
	}
}

func TestValidatePersistedFlightIdentityRejectsInvalidKey(
	t *testing.T,
) {
	item := validPersistedIdentityTrajectory()
	item.IdentityKey = "flight-identity-not-a-sha256-digest"

	err := validatePersistedFlightIdentity(item)
	if !errors.Is(err, errFlightIdentityKeyInvalid) {
		t.Fatalf("expected invalid identity key error, got %v", err)
	}
}

func TestValidatePersistedFlightIdentityRejectsUnknownBasis(
	t *testing.T,
) {
	item := validPersistedIdentityTrajectory()
	item.IdentityBasis = trajectory.FlightIdentityBasis("unknown")

	err := validatePersistedFlightIdentity(item)
	if !errors.Is(err, errFlightIdentityBasisInvalid) {
		t.Fatalf("expected invalid identity basis error, got %v", err)
	}
}

func TestValidatePersistedFlightIdentityRejectsUnknownSplitReason(
	t *testing.T,
) {
	item := validPersistedIdentityTrajectory()
	item.SplitReason = trajectory.FlightSplitReason("unknown")

	err := validatePersistedFlightIdentity(item)
	if !errors.Is(err, errFlightSplitReasonInvalid) {
		t.Fatalf("expected invalid split reason error, got %v", err)
	}
}

func validPersistedIdentityTrajectory() trajectory.FlightTrajectory {
	return trajectory.FlightTrajectory{
		IdentityKey: persistedFlightIdentityKeyPrefix + strings.Repeat("a", 64),
		IdentityBasis: trajectory.
			FlightIdentityBasisCallsignAndStartTime,
		SplitReason: trajectory.FlightSplitReasonInitialObservation,
	}
}
