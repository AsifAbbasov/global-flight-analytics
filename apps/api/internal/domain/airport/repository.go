package airport

import "context"

type Repository interface {
	List(ctx context.Context) ([]Airport, error)
	GetByICAO(ctx context.Context, icao string) (Airport, error)
}

type PagedRepository interface {
	ListPage(ctx context.Context, request ListRequest) (ListPage, error)
}
