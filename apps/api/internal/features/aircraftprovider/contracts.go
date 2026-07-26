package aircraftprovider

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
)

const Version = "aircraft-feature-provider-v3"

const DefaultNotFoundPolicyVersion = "aircraft-not-found-pgx-no-rows-v1"
const MetadataSourceName = "aircraft-reference-lookup"

type CacheMode string

const (
	CacheModeDefault  CacheMode = ""
	CacheModeEnabled  CacheMode = "enabled"
	CacheModeDisabled CacheMode = "disabled"
)

const (
	DefaultPositiveCacheTTL = 30 * time.Minute
	DefaultNegativeCacheTTL = 5 * time.Minute
)

type AircraftLookup interface {
	GetByICAO24(
		ctx context.Context,
		icao24 string,
	) (aircraft.Aircraft, error)
}

type Config struct {
	Lookup           AircraftLookup
	CacheMode        CacheMode
	PositiveCacheTTL time.Duration
	NegativeCacheTTL time.Duration
	Now              func() time.Time
	IsNotFound       func(error) bool
}
