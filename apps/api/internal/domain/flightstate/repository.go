package flightstate

import "context"

type Repository interface {
	ListByFlightID(ctx context.Context, flightID string) ([]FlightState, error)
	GetLatestByICAO24(ctx context.Context, icao24 string) (FlightState, error)
}
