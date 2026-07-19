package historicalcursor

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregatecontract"
)

func TestCursorCodecRoundTrip(
	t *testing.T,
) {
	cursor := historicalaggregatecontract.ListCursor{
		WindowEnd: time.Date(
			2026,
			time.July,
			19,
			12,
			0,
			0,
			123,
			time.UTC,
		),
		WindowStart: time.Date(
			2026,
			time.July,
			19,
			11,
			0,
			0,
			123,
			time.UTC,
		),
		AsOfTime: time.Date(
			2026,
			time.July,
			19,
			12,
			5,
			0,
			123,
			time.UTC,
		),
		ID: "record-a",
	}

	encoded, err := Encode(cursor)
	if err != nil {
		t.Fatalf(
			"Encode() error = %v",
			err,
		)
	}
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf(
			"Decode() error = %v",
			err,
		)
	}
	if decoded == nil ||
		!decoded.WindowEnd.Equal(
			cursor.WindowEnd,
		) ||
		!decoded.WindowStart.Equal(
			cursor.WindowStart,
		) ||
		!decoded.AsOfTime.Equal(
			cursor.AsOfTime,
		) ||
		decoded.ID != cursor.ID {
		t.Fatalf(
			"cursor round trip changed data: encoded=%q decoded=%#v",
			encoded,
			decoded,
		)
	}
}

func TestCursorCodecTreatsEmptyValueAsFirstPage(
	t *testing.T,
) {
	cursor, err := Decode("   ")
	if err != nil {
		t.Fatalf(
			"Decode() error = %v",
			err,
		)
	}
	if cursor != nil {
		t.Fatalf(
			"empty cursor decoded as %#v",
			cursor,
		)
	}
}

func TestCursorCodecRejectsMalformedAndOversizedValues(
	t *testing.T,
) {
	for _, value := range []string{
		"not-base64",
		strings.Repeat(
			"a",
			MaximumEncodedLength+1,
		),
	} {
		_, err := Decode(value)
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf(
				"Decode(%q) error = %v, want ErrInvalid",
				value,
				err,
			)
		}
	}
}
