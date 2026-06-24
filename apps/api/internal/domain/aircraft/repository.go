package aircraft

import "context"

type Repository interface {
	List(ctx context.Context) ([]Aircraft, error)
	GetByICAO24(ctx context.Context, icao24 string) (Aircraft, error)
}
