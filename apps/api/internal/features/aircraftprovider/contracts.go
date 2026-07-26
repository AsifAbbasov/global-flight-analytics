package aircraftprovider

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
)

const Version = "aircraft-feature-provider-v4"

const DefaultNotFoundPolicyVersion = "aircraft-domain-not-found-v2"
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
	DefaultMaxCacheEntries  = 4096
	DefaultLookupTimeout    = 15 * time.Second
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
