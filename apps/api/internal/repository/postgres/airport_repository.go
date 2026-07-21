package postgres

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const legacyAirportListPageSize = airport.MaximumListPageSize

type AirportRepository struct {
	pool *pgxpool.Pool
}

var (
	_ airport.Repository      = (*AirportRepository)(nil)
	_ airport.PagedRepository = (*AirportRepository)(nil)
)

func NewAirportRepository(pool *pgxpool.Pool) *AirportRepository {
	return &AirportRepository{pool: pool}
}

func (repository *AirportRepository) List(
	ctx context.Context,
) ([]airport.Airport, error) {
	request := airport.ListRequest{Limit: legacyAirportListPageSize}
	items := make([]airport.Airport, 0, legacyAirportListPageSize)

	for {
		page, err := repository.ListPage(ctx, request)
		if err != nil {
			return nil, err
		}
		items = append(items, page.Items...)
		if page.NextCursor == nil {
			return items, nil
		}
		request.Cursor = page.NextCursor
	}
}

func (repository *AirportRepository) GetByICAO(
	ctx context.Context,
	icao string,
) (airport.Airport, error) {
	if err := requireRepositoryContext(ctx); err != nil {
		return airport.Airport{}, err
	}

	record, err := scanAirportRecord(
		repository.pool.QueryRow(ctx, airportByICAOQuery, icao),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return airport.Airport{}, airport.ErrNotFound
		}
		return airport.Airport{}, err
	}

	return record.Item, nil
}
