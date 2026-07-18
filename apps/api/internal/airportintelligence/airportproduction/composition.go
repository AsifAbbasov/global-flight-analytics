package airportproduction

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(pool *pgxpool.Pool) (*Service, error) {
	if pool == nil {
		return nil, ErrPostgresPoolRequired
	}
	reader, err := NewPostgresObservationReader(pool)
	if err != nil {
		return nil, err
	}
	return New(Config{
		AirportRepository: postgres.NewAirportRepository(pool),
		ObservationReader: reader,
		MaximumDataAge:    48 * time.Hour,
	})
}
