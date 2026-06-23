package postgres

import (
	"context"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AirportRepository struct {
	pool *pgxpool.Pool
}

func NewAirportRepository(pool *pgxpool.Pool) *AirportRepository {
	return &AirportRepository{
		pool: pool,
	}
}

func (r *AirportRepository) List(ctx context.Context) ([]airport.Airport, error) {
	return []airport.Airport{}, nil
}

func (r *AirportRepository) GetByICAO(ctx context.Context, icao string) (airport.Airport, error) {
	return airport.Airport{}, nil
}
