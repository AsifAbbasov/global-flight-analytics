package live

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

var icao24Pattern = regexp.MustCompile(`^[0-9a-f]{6}$`)

var (
	ErrInvalidBounds   = errors.New("live traffic bounds are invalid")
	ErrInvalidLimit    = errors.New("live traffic snapshot limit is invalid")
	ErrTooManySelected = errors.New("too many selected aircraft")
)

type Aircraft struct {
	ICAO24          string
	Callsign        string
	Latitude        float64
	Longitude       float64
	AltitudeM       *float64
	VelocityMPS     *float64
	HeadingDegrees  *float64
	VerticalRateMPS *float64
	OnGround        *bool
	ObservedAt      time.Time
	ReceivedAt      time.Time
	Source          string
}

type Bounds struct {
	MinLatitude  float64
	MinLongitude float64
	MaxLatitude  float64
	MaxLongitude float64
}

func (b Bounds) Validate() error {
	values := []float64{
		b.MinLatitude,
		b.MinLongitude,
		b.MaxLatitude,
		b.MaxLongitude,
	}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return ErrInvalidBounds
		}
	}
	if b.MinLatitude < -90 || b.MaxLatitude > 90 ||
		b.MinLongitude < -180 || b.MaxLongitude > 180 ||
		b.MinLatitude > b.MaxLatitude ||
		b.MinLongitude > b.MaxLongitude {
		return ErrInvalidBounds
	}
	return nil
}

func (b Bounds) Contains(latitude, longitude float64) bool {
	return latitude >= b.MinLatitude && latitude <= b.MaxLatitude &&
		longitude >= b.MinLongitude && longitude <= b.MaxLongitude
}

type SnapshotQuery struct {
	Bounds         *Bounds
	SelectedICAO24 []string
	Limit          int
}

type Snapshot struct {
	ServerTime  time.Time
	Sequence    uint64
	Aircraft    []Aircraft
	TotalActive int
	Matching    int
	Truncated   bool
}

func NormalizeICAO24(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if !icao24Pattern.MatchString(normalized) {
		return "", fmt.Errorf("invalid ICAO24 %q", value)
	}
	return normalized, nil
}
