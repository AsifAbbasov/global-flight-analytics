package postgres

import (
	"errors"
	"testing"
)

func TestNullableUUIDReturnsDatabaseNullForEmptyValue(t *testing.T) {
	value, err := nullableUUID("").Value()
	if err != nil {
		t.Fatalf("nullable UUID value: %v", err)
	}
	if value != nil {
		t.Fatalf("expected database null, got %#v", value)
	}
}

func TestNullableUUIDReturnsCanonicalValue(t *testing.T) {
	value, err := nullableUUID(
		"  11111111-1111-1111-1111-111111111111  ",
	).Value()
	if err != nil {
		t.Fatalf("nullable UUID value: %v", err)
	}
	if value != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected canonical UUID, got %#v", value)
	}
}

func TestNullableUUIDRejectsMalformedValue(t *testing.T) {
	_, err := nullableUUID("not-a-uuid").Value()
	if !errors.Is(err, ErrRepositoryUUIDArgumentInvalid) {
		t.Fatalf("expected UUID argument error, got %v", err)
	}
}

func TestNullableTextReturnsDatabaseNullForEmptyValue(t *testing.T) {
	value, err := nullableText("   ").Value()
	if err != nil {
		t.Fatalf("nullable text value: %v", err)
	}
	if value != nil {
		t.Fatalf("expected database null, got %#v", value)
	}
}

func TestNullableTextReturnsNormalizedValue(t *testing.T) {
	value, err := nullableText("  AHY101  ").Value()
	if err != nil {
		t.Fatalf("nullable text value: %v", err)
	}
	if value != "AHY101" {
		t.Fatalf("expected normalized text, got %#v", value)
	}
}

func TestRequiredSourceNameRejectsMissingEvidence(t *testing.T) {
	_, err := requiredSourceNameValue("   ").Value()
	if !errors.Is(err, ErrRepositorySourceNameRequired) {
		t.Fatalf("expected source-name error, got %v", err)
	}
}

func TestRequiredSourceNameReturnsNormalizedEvidence(t *testing.T) {
	value, err := requiredSourceNameValue("  test  ").Value()
	if err != nil {
		t.Fatalf("required source name value: %v", err)
	}
	if value != "test" {
		t.Fatalf("expected normalized source name, got %#v", value)
	}
}

func TestInferredPreviousSegmentID(t *testing.T) {
	segmentIDs := []string{"segment-1", "segment-2"}

	if inferredPreviousSegmentID(0, segmentIDs, "") != "segment-1" {
		t.Fatal("expected first segment id")
	}
	if inferredPreviousSegmentID(0, segmentIDs, "explicit") != "explicit" {
		t.Fatal("expected explicit segment id")
	}
}

func TestInferredNextSegmentID(t *testing.T) {
	segmentIDs := []string{"segment-1", "segment-2"}

	if inferredNextSegmentID(0, segmentIDs, "") != "segment-2" {
		t.Fatal("expected second segment id")
	}
	if inferredNextSegmentID(0, segmentIDs, "explicit") != "explicit" {
		t.Fatal("expected explicit segment id")
	}
}
