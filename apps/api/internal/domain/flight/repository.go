package flight

import "context"

type Repository interface {
	List(ctx context.Context) ([]Flight, error)
	GetByID(ctx context.Context, id string) (Flight, error)
}
