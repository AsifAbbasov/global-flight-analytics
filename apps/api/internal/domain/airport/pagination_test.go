package airport

import (
	"errors"
	"testing"
)

func TestNormalizeListRequestAppliesDefaultAndClonesCursor(t *testing.T) {
	t.Parallel()

	cursor := &ListCursor{
		Name: "Baku Heydar Aliyev International Airport",
		ID:   " 11111111-1111-1111-1111-111111111111 ",
	}
	normalized, err := NormalizeListRequest(ListRequest{Cursor: cursor})
	if err != nil {
		t.Fatalf("normalize list request: %v", err)
	}
	if normalized.Limit != DefaultListPageSize {
		t.Fatalf("limit = %d, want %d", normalized.Limit, DefaultListPageSize)
	}
	if normalized.Cursor == cursor {
		t.Fatal("normalized cursor aliases caller-owned cursor")
	}
	if normalized.Cursor.Name != cursor.Name {
		t.Fatalf("cursor name = %q, want %q", normalized.Cursor.Name, cursor.Name)
	}
	if normalized.Cursor.ID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("cursor id = %q", normalized.Cursor.ID)
	}
}

func TestNormalizeListRequestRejectsInvalidLimits(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{-1, MaximumListPageSize + 1} {
		_, err := NormalizeListRequest(ListRequest{Limit: limit})
		if !errors.Is(err, ErrListPageSizeInvalid) {
			t.Fatalf("limit %d: expected page-size error, got %v", limit, err)
		}
	}
}

func TestNormalizeListRequestRejectsIncompleteCursor(t *testing.T) {
	t.Parallel()

	for _, cursor := range []*ListCursor{
		{Name: "Airport"},
		{ID: "11111111-1111-1111-1111-111111111111"},
		{Name: "   ", ID: "11111111-1111-1111-1111-111111111111"},
	} {
		_, err := NormalizeListRequest(ListRequest{Cursor: cursor})
		if !errors.Is(err, ErrListCursorInvalid) {
			t.Fatalf("cursor %#v: expected cursor error, got %v", cursor, err)
		}
	}
}
