package historicalaggregatecontract

import (
	"errors"
	"testing"
	"time"
)

func TestNormalizeListCursorCanonicalizesCompleteCursor(
	t *testing.T,
) {
	location := time.FixedZone(
		"cursor-test",
		4*60*60,
	)
	cursor := ListCursor{
		WindowEnd: time.Date(
			2026,
			time.July,
			19,
			16,
			0,
			0,
			123,
			location,
		),
		WindowStart: time.Date(
			2026,
			time.July,
			19,
			15,
			0,
			0,
			123,
			location,
		),
		AsOfTime: time.Date(
			2026,
			time.July,
			19,
			16,
			5,
			0,
			123,
			location,
		),
		ID: "  record-a  ",
	}

	normalized, err := NormalizeListCursor(cursor)
	if err != nil {
		t.Fatalf(
			"NormalizeListCursor() error = %v",
			err,
		)
	}
	if normalized.ID != "record-a" ||
		normalized.WindowEnd.Location() != time.UTC ||
		normalized.WindowStart.Location() != time.UTC ||
		normalized.AsOfTime.Location() != time.UTC {
		t.Fatalf(
			"unexpected normalized cursor: %#v",
			normalized,
		)
	}
}

func TestNormalizeListCursorRejectsPartialAndInvalidOrdering(
	t *testing.T,
) {
	end := time.Date(
		2026,
		time.July,
		19,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	testCases := []ListCursor{
		{
			WindowEnd: end,
		},
		{
			WindowEnd:   end,
			WindowStart: end,
			AsOfTime:    end,
			ID:          "record-a",
		},
		{
			WindowEnd:   end,
			WindowStart: end.Add(-time.Hour),
			AsOfTime:    end.Add(-time.Second),
			ID:          "record-a",
		},
	}

	for _, cursor := range testCases {
		_, err := NormalizeListCursor(cursor)
		if !errors.Is(
			err,
			ErrInvalidListCursor,
		) {
			t.Fatalf(
				"cursor %#v error = %v, want ErrInvalidListCursor",
				cursor,
				err,
			)
		}
	}
}

func TestPageCloneSeparatesNextCursor(
	t *testing.T,
) {
	cursor := &ListCursor{
		WindowEnd: time.Date(
			2026,
			time.July,
			19,
			12,
			0,
			0,
			0,
			time.UTC,
		),
		WindowStart: time.Date(
			2026,
			time.July,
			19,
			11,
			0,
			0,
			0,
			time.UTC,
		),
		AsOfTime: time.Date(
			2026,
			time.July,
			19,
			12,
			5,
			0,
			0,
			time.UTC,
		),
		ID: "record-a",
	}
	page := Page{
		HasMore:    true,
		NextCursor: cursor,
	}

	cloned := page.Clone()
	cloned.NextCursor.ID = "record-b"

	if page.NextCursor.ID != "record-a" {
		t.Fatal(
			"page clone shares the pagination cursor",
		)
	}
}

func TestListCursorZeroRequiresEveryComponentToBeAbsent(
	t *testing.T,
) {
	if !(ListCursor{}).IsZero() {
		t.Fatal(
			"empty cursor is not zero",
		)
	}
	if (ListCursor{
		ID: "record-a",
	}).IsZero() {
		t.Fatal(
			"partially populated cursor was treated as zero",
		)
	}
}
