package transponderalert

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

const DefaultMaximumFreshAge = 5 * time.Minute

var (
	ErrLatestStateReaderRequired = errors.New(
		"transponder evidence latest flight state reader is required",
	)
	ErrICAO24Invalid = errors.New(
		"transponder evidence ICAO24 is invalid",
	)
	ErrEvidenceNotFound = errors.New(
		"transponder evidence was not found",
	)
)

var productionICAO24Pattern = regexp.MustCompile(
	`^[A-F0-9]{6}$`,
)

type FreshnessStatus string

const (
	FreshnessRecent FreshnessStatus = "recent"
	FreshnessStale  FreshnessStatus = "stale"
)

type ConfidenceLevel string

const (
	ConfidenceLimited  ConfidenceLevel = "limited"
	ConfidenceDegraded ConfidenceLevel = "degraded"
)

type Confidence struct {
	Level   ConfidenceLevel
	Reasons []string
}

type LatestEvidence struct {
	Evidence Evidence

	FreshnessStatus FreshnessStatus
	Age             time.Duration
	MaximumFreshAge time.Duration
	Confidence      Confidence

	EvidenceOnly       bool
	ConfirmedEmergency bool
	OperationalAlert   bool
}

type LatestStateReader interface {
	GetLatestByICAO24(
		ctx context.Context,
		icao24 string,
	) (flightstate.FlightState, error)
}

type ServiceConfig struct {
	LatestStateReader LatestStateReader
	MaximumFreshAge   time.Duration
	Now               func() time.Time
}

type Service struct {
	latestStateReader LatestStateReader
	maximumFreshAge   time.Duration
	now               func() time.Time
}

func NewService(
	config ServiceConfig,
) (*Service, error) {
	if config.LatestStateReader == nil {
		return nil, ErrLatestStateReaderRequired
	}

	maximumFreshAge := config.MaximumFreshAge
	if maximumFreshAge <= 0 {
		maximumFreshAge = DefaultMaximumFreshAge
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Service{
		latestStateReader: config.LatestStateReader,
		maximumFreshAge:   maximumFreshAge,
		now:               now,
	}, nil
}

func (service *Service) GetLatest(
	ctx context.Context,
	icao24 string,
) (LatestEvidence, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return LatestEvidence{}, err
	}

	normalizedICAO24 := strings.ToUpper(
		strings.TrimSpace(icao24),
	)
	if !productionICAO24Pattern.MatchString(
		normalizedICAO24,
	) {
		return LatestEvidence{}, ErrICAO24Invalid
	}

	state, err := service.latestStateReader.
		GetLatestByICAO24(
			ctx,
			normalizedICAO24,
		)
	if err != nil {
		return LatestEvidence{}, fmt.Errorf(
			"read latest flight state for ICAO24 %s: %w",
			normalizedICAO24,
			err,
		)
	}

	state.ICAO24 = normalizedICAO24
	if strings.TrimSpace(state.SquawkCode) == "" {
		return LatestEvidence{}, ErrEvidenceNotFound
	}

	asOfTime := service.now().UTC()
	evidenceItems, err := Build(
		[]flightstate.FlightState{state},
		asOfTime,
	)
	if err != nil {
		return LatestEvidence{}, fmt.Errorf(
			"build latest transponder evidence for ICAO24 %s: %w",
			normalizedICAO24,
			err,
		)
	}
	if len(evidenceItems) == 0 {
		return LatestEvidence{}, ErrEvidenceNotFound
	}

	evidence := evidenceItems[0]
	age := asOfTime.Sub(evidence.LastObservedAt)
	freshnessStatus := FreshnessRecent
	confidenceLevel := ConfidenceLimited
	confidenceReasons := []string{
		"The assessment is based on the latest persisted external observation only.",
		"The observed code is evidence of a transmitted value and does not confirm an emergency or incident cause.",
	}

	if age > service.maximumFreshAge {
		freshnessStatus = FreshnessStale
		confidenceLevel = ConfidenceDegraded
		confidenceReasons = append(
			confidenceReasons,
			"The latest persisted observation is older than the configured freshness threshold.",
		)
		evidence.Limitations = append(
			evidence.Limitations,
			"The latest persisted transponder observation is stale and may no longer describe the aircraft's current transmitted code.",
		)
	}

	return LatestEvidence{
		Evidence:        evidence,
		FreshnessStatus: freshnessStatus,
		Age:             age,
		MaximumFreshAge: service.maximumFreshAge,
		Confidence: Confidence{
			Level: confidenceLevel,
			Reasons: append(
				[]string(nil),
				confidenceReasons...,
			),
		},
		EvidenceOnly:       true,
		ConfirmedEmergency: false,
		OperationalAlert:   false,
	}, nil
}
