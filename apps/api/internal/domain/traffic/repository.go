package traffic

import "context"

type Repository interface {
	GetCurrent(ctx context.Context) ([]CurrentTrafficItem, error)
}
