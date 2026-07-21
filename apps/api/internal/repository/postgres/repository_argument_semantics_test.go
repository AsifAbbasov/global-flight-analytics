package postgres

import (
	"errors"
	"testing"
)

func TestNullableUUIDArgumentRepresentsAbsentValueWithoutTypedNil(t *testing.T) {
	t.Parallel()

	argument := nullableUUID("   ")
	value, err := argument.Value()
	if err != nil {
		t.Fatalf("nullable UUID value: %v", err)
	}
	if value != nil {
		t.Fatalf("nullable UUID value = %#v, want nil", value)
	}
}

func TestNullableUUIDArgumentRejectsMalformedIdentifier(t *testing.T) {
	t.Parallel()

	_, err := nullableUUID("not-a-uuid").Value()
	if !errors.Is(err, ErrRepositoryUUIDArgumentInvalid) {
		t.Fatalf("expected UUID argument error, got %v", err)
	}
}

func TestNullableUUIDArgumentNormalizesValidIdentifier(t *testing.T) {
	t.Parallel()

	value, err := nullableUUID(
		" 11111111-1111-1111-1111-111111111111 ",
	).Value()
	if err != nil {
		t.Fatalf("nullable UUID value: %v", err)
	}
	if value != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("UUID value = %#v", value)
	}
}

func TestNullableTextArgumentPreservesNullAndNormalizedText(t *testing.T) {
	t.Parallel()

	value, err := nullableText("   ").Value()
	if err != nil {
		t.Fatalf("nullable text value: %v", err)
	}
	if value != nil {
		t.Fatalf("nullable text value = %#v, want nil", value)
	}

	value, err = nullableText("  callsign  ").Value()
	if err != nil {
		t.Fatalf("nullable text value: %v", err)
	}
	if value != "callsign" {
		t.Fatalf("nullable text value = %#v", value)
	}
}

func TestRequiredSourceNameArgumentRejectsMissingEvidence(t *testing.T) {
	t.Parallel()

	_, err := requiredSourceNameValue("   ").Value()
	if !errors.Is(err, ErrRepositorySourceNameRequired) {
		t.Fatalf("expected source-name error, got %v", err)
	}
}

func TestRequiredSourceNameArgumentNormalizesEvidence(t *testing.T) {
	t.Parallel()

	value, err := requiredSourceNameValue("  opensky  ").Value()
	if err != nil {
		t.Fatalf("required source name value: %v", err)
	}
	if value != "opensky" {
		t.Fatalf("source name value = %#v", value)
	}
}
