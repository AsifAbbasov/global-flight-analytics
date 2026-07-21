package postgres

import (
	"context"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (repository *AirportRepository) ListPage(
	ctx context.Context,
	request airport.ListRequest,
) (airport.ListPage, error) {
	normalized, err := airport.NormalizeListRequest(request)
	if err != nil {
		return airport.ListPage{}, err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	queryLimit := normalized.Limit + 1
	var rows pgx.Rows
	if normalized.Cursor == nil {
		rows, err = repository.pool.Query(
			ctx,
			airportListFirstPageQuery,
			queryLimit,
		)
	} else {
		cursorID, parseErr := parseAirportCursorID(normalized.Cursor.ID)
		if parseErr != nil {
			return airport.ListPage{}, parseErr
		}
		rows, err = repository.pool.Query(
			ctx,
			airportListAfterCursorQuery,
			normalized.Cursor.Name,
			cursorID,
			queryLimit,
		)
	}
	if err != nil {
		return airport.ListPage{}, err
	}
	defer rows.Close()

	records := make([]airportRecord, 0, queryLimit)
	for rows.Next() {
		record, scanErr := scanAirportRecord(rows)
		if scanErr != nil {
			return airport.ListPage{}, scanErr
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return airport.ListPage{}, err
	}

	return buildAirportPage(records, normalized.Limit), nil
}

func parseAirportCursorID(value string) (pgtype.UUID, error) {
	var identifier pgtype.UUID
	if err := identifier.Scan(value); err != nil || !identifier.Valid {
		return pgtype.UUID{}, fmt.Errorf(
			"%w: id %q is not a valid UUID",
			airport.ErrListCursorInvalid,
			value,
		)
	}
	return identifier, nil
}

func buildAirportPage(
	records []airportRecord,
	limit int,
) airport.ListPage {
	returnedCount := len(records)
	if returnedCount > limit {
		returnedCount = limit
	}

	page := airport.ListPage{
		Items: make([]airport.Airport, returnedCount),
	}
	for index := 0; index < returnedCount; index++ {
		page.Items[index] = records[index].Item
	}

	if len(records) > limit && limit > 0 {
		lastReturned := records[limit-1]
		page.NextCursor = &airport.ListCursor{
			Name: lastReturned.Item.Name,
			ID:   lastReturned.ID,
		}
	}
	return page
}
