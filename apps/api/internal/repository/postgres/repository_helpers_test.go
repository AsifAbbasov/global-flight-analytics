package postgres

import "testing"

func TestNullableUUIDReturnsNilForEmptyValue(
	t *testing.T,
) {
	result := nullableUUID(
		"",
	)

	if result != nil {
		t.Fatalf(
			"expected nil, got %v",
			result,
		)
	}
}

func TestNullableUUIDReturnsTrimmedValue(
	t *testing.T,
) {
	result := nullableUUID(
		"  11111111-1111-1111-1111-111111111111  ",
	)

	if result == nil {
		t.Fatal(
			"expected non-nil uuid value",
		)
	}

	if *result !=
		"11111111-1111-1111-1111-111111111111" {
		t.Fatalf(
			"expected trimmed uuid, got %q",
			*result,
		)
	}
}

func TestNullableTextReturnsNilForEmptyValue(
	t *testing.T,
) {
	result := nullableText(
		"   ",
	)

	if result != nil {
		t.Fatalf(
			"expected nil, got %v",
			result,
		)
	}
}

func TestNullableTextReturnsTrimmedValue(
	t *testing.T,
) {
	result := nullableText(
		"  AHY101  ",
	)

	if result == nil {
		t.Fatal(
			"expected non-nil text value",
		)
	}

	if *result != "AHY101" {
		t.Fatalf(
			"expected trimmed text, got %q",
			*result,
		)
	}
}

func TestSourceNameOrUnknown(
	t *testing.T,
) {
	if sourceNameOrUnknown(
		"",
	) != "unknown" {
		t.Fatal(
			"expected unknown source name",
		)
	}

	if sourceNameOrUnknown(
		"  test  ",
	) != "test" {
		t.Fatal(
			"expected trimmed source name",
		)
	}
}

func TestInferredPreviousSegmentID(
	t *testing.T,
) {
	segmentIDs := []string{
		"segment-1",
		"segment-2",
	}

	if inferredPreviousSegmentID(
		0,
		segmentIDs,
		"",
	) != "segment-1" {
		t.Fatal(
			"expected first segment id",
		)
	}

	if inferredPreviousSegmentID(
		0,
		segmentIDs,
		"explicit",
	) != "explicit" {
		t.Fatal(
			"expected explicit segment id",
		)
	}
}

func TestInferredNextSegmentID(
	t *testing.T,
) {
	segmentIDs := []string{
		"segment-1",
		"segment-2",
	}

	if inferredNextSegmentID(
		0,
		segmentIDs,
		"",
	) != "segment-2" {
		t.Fatal(
			"expected second segment id",
		)
	}

	if inferredNextSegmentID(
		0,
		segmentIDs,
		"explicit",
	) != "explicit" {
		t.Fatal(
			"expected explicit segment id",
		)
	}
}
