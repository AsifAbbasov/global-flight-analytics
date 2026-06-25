package traffic

import "context"

type Bounds struct {
	MinLatitude  float64
	MaxLatitude  float64
	MinLongitude float64
	MaxLongitude float64
}

type Repository interface {
	GetCurrent(ctx context.Context) ([]CurrentTrafficItem, error)
	GetCurrentByBounds(ctx context.Context, bounds Bounds) ([]CurrentTrafficItem, error)
}
