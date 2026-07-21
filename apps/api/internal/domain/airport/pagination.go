package airport

import (
	"errors"
	"fmt"
	"strings"
)

const (
	DefaultListPageSize       = 200
	MaximumListPageSize       = 1000
	MaximumListCursorNameSize = 512
	MaximumListCursorIDSize   = 64
)

var (
	ErrListPageSizeInvalid = errors.New("airport list page size is invalid")
	ErrListCursorInvalid   = errors.New("airport list cursor is invalid")
)

type ListCursor struct {
	Name string
	ID   string
}

type ListRequest struct {
	Limit  int
	Cursor *ListCursor
}

type ListPage struct {
	Items      []Airport
	NextCursor *ListCursor
}

func NormalizeListRequest(request ListRequest) (ListRequest, error) {
	normalized := request
	if normalized.Limit == 0 {
		normalized.Limit = DefaultListPageSize
	}
	if normalized.Limit < 1 || normalized.Limit > MaximumListPageSize {
		return ListRequest{}, fmt.Errorf(
			"%w: got %d, allowed range is 1..%d",
			ErrListPageSizeInvalid,
			normalized.Limit,
			MaximumListPageSize,
		)
	}

	if normalized.Cursor == nil {
		return normalized, nil
	}

	name := normalized.Cursor.Name
	identifier := strings.TrimSpace(normalized.Cursor.ID)
	if strings.TrimSpace(name) == "" || identifier == "" {
		return ListRequest{}, fmt.Errorf(
			"%w: both name and id are required",
			ErrListCursorInvalid,
		)
	}
	if len(name) > MaximumListCursorNameSize ||
		len(identifier) > MaximumListCursorIDSize {
		return ListRequest{}, fmt.Errorf(
			"%w: cursor component exceeds the maximum length",
			ErrListCursorInvalid,
		)
	}

	normalized.Cursor = &ListCursor{
		Name: name,
		ID:   identifier,
	}
	return normalized, nil
}
