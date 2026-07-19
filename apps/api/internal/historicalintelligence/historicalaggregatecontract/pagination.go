package historicalaggregatecontract

import (
	"strings"
	"time"
)

const MaximumListCursorIdentifierLength = 256

type ListCursor struct {
	WindowEnd   time.Time
	WindowStart time.Time
	AsOfTime    time.Time
	ID          string
}

func (cursor ListCursor) Clone() ListCursor {
	return ListCursor{
		WindowEnd:   cursor.WindowEnd,
		WindowStart: cursor.WindowStart,
		AsOfTime:    cursor.AsOfTime,
		ID:          cursor.ID,
	}
}

func (cursor ListCursor) IsZero() bool {
	return cursor.WindowEnd.IsZero() &&
		cursor.WindowStart.IsZero() &&
		cursor.AsOfTime.IsZero() &&
		strings.TrimSpace(cursor.ID) == ""
}

func NormalizeListCursor(
	cursor ListCursor,
) (ListCursor, error) {
	identifier := strings.TrimSpace(cursor.ID)
	if cursor.WindowEnd.IsZero() ||
		cursor.WindowStart.IsZero() ||
		cursor.AsOfTime.IsZero() ||
		identifier == "" ||
		len(identifier) >
			MaximumListCursorIdentifierLength {
		return ListCursor{},
			ErrInvalidListCursor
	}

	windowEnd := cursor.WindowEnd.UTC()
	windowStart := cursor.WindowStart.UTC()
	asOfTime := cursor.AsOfTime.UTC()
	if !windowStart.Before(windowEnd) ||
		asOfTime.Before(windowEnd) {
		return ListCursor{},
			ErrInvalidListCursor
	}

	return ListCursor{
		WindowEnd:   windowEnd,
		WindowStart: windowStart,
		AsOfTime:    asOfTime,
		ID:          identifier,
	}, nil
}
