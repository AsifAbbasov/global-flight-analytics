package postgres

import (
	"errors"
	"testing"
)

func TestTrajectoryUUIDArgumentsPreserveNativeUUIDType(t *testing.T) {
	arguments, err := trajectoryUUIDArguments([]string{
		" 11111111-1111-1111-1111-111111111111 ",
	})
	if err != nil {
		t.Fatalf("trajectory UUID arguments error = %v", err)
	}
	if len(arguments) != 1 || !arguments[0].Valid {
		t.Fatalf("trajectory UUID arguments = %#v", arguments)
	}
}

func TestTrajectoryUUIDArgumentsRejectInvalidIdentifier(t *testing.T) {
	_, err := trajectoryUUIDArguments([]string{"not-a-uuid"})
	if !errors.Is(err, ErrTrajectoryIdentifierInvalid) {
		t.Fatalf("invalid trajectory identifier error = %v", err)
	}
}
