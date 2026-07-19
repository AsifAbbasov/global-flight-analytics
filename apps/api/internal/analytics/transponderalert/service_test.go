package transponderalert

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

type latestStateReaderStub struct {
	state flightstate.FlightState
	err   error
}

func (stub latestStateReaderStub) GetLatestByICAO24(
	_ context.Context,
	_ string,
) (flightstate.FlightState, error) {
	return stub.state, stub.err
}

func TestServiceReturnsEvidenceWithoutConfirmingEmergency(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		19,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	service, err := NewService(
		ServiceConfig{
			LatestStateReader: latestStateReaderStub{
				state: flightstate.FlightState{
					ICAO24:     "4a001a",
					Callsign:   "AHY101",
					SquawkCode: "7700",
					ObservedAt: now.Add(-30 * time.Second),
					SourceName: "opensky",
				},
			},
			Now: func() time.Time {
				return now
			},
		},
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	result, err := service.GetLatest(
		context.Background(),
		"4a001a",
	)
	if err != nil {
		t.Fatalf("get latest evidence: %v", err)
	}

	if result.Evidence.ICAO24 != "4A001A" {
		t.Fatalf(
			"ICAO24 = %q",
			result.Evidence.ICAO24,
		)
	}
	if result.Evidence.SquawkCode != "7700" {
		t.Fatalf(
			"squawk = %q",
			result.Evidence.SquawkCode,
		)
	}
	if result.FreshnessStatus != FreshnessRecent {
		t.Fatalf(
			"freshness = %q",
			result.FreshnessStatus,
		)
	}
	if result.Confidence.Level != ConfidenceLimited {
		t.Fatalf(
			"confidence = %q",
			result.Confidence.Level,
		)
	}
	if !result.EvidenceOnly {
		t.Fatal("evidence-only flag is false")
	}
	if result.ConfirmedEmergency {
		t.Fatal("emergency was incorrectly confirmed")
	}
	if result.OperationalAlert {
		t.Fatal("operational alert was incorrectly produced")
	}
}

func TestServiceMarksOldEvidenceStale(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		19,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	service, err := NewService(
		ServiceConfig{
			LatestStateReader: latestStateReaderStub{
				state: flightstate.FlightState{
					ICAO24:     "4A001A",
					SquawkCode: "7600",
					ObservedAt: now.Add(-10 * time.Minute),
					SourceName: "opensky",
				},
			},
			MaximumFreshAge: 5 * time.Minute,
			Now: func() time.Time {
				return now
			},
		},
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	result, err := service.GetLatest(
		context.Background(),
		"4A001A",
	)
	if err != nil {
		t.Fatalf("get latest evidence: %v", err)
	}
	if result.FreshnessStatus != FreshnessStale {
		t.Fatalf(
			"freshness = %q, want stale",
			result.FreshnessStatus,
		)
	}
	if result.Confidence.Level != ConfidenceDegraded {
		t.Fatalf(
			"confidence = %q, want degraded",
			result.Confidence.Level,
		)
	}
	if len(result.Evidence.Limitations) < 5 {
		t.Fatalf(
			"stale limitations = %v",
			result.Evidence.Limitations,
		)
	}
}

func TestServiceReturnsNotFoundForOrdinaryCode(
	t *testing.T,
) {
	now := time.Now().UTC()
	service, err := NewService(
		ServiceConfig{
			LatestStateReader: latestStateReaderStub{
				state: flightstate.FlightState{
					ICAO24:     "4A001A",
					SquawkCode: "1200",
					ObservedAt: now,
				},
			},
			Now: func() time.Time {
				return now
			},
		},
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	_, err = service.GetLatest(
		context.Background(),
		"4A001A",
	)
	if !errors.Is(err, ErrEvidenceNotFound) {
		t.Fatalf(
			"error = %v, want evidence not found",
			err,
		)
	}
}

func TestServiceRejectsInvalidICAO24(
	t *testing.T,
) {
	service, err := NewService(
		ServiceConfig{
			LatestStateReader: latestStateReaderStub{},
		},
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	_, err = service.GetLatest(
		context.Background(),
		"invalid",
	)
	if !errors.Is(err, ErrICAO24Invalid) {
		t.Fatalf(
			"error = %v, want invalid ICAO24",
			err,
		)
	}
}

func TestServicePreservesFlightStateNotFound(
	t *testing.T,
) {
	service, err := NewService(
		ServiceConfig{
			LatestStateReader: latestStateReaderStub{
				err: flightstate.ErrNotFound,
			},
		},
	)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	_, err = service.GetLatest(
		context.Background(),
		"4A001A",
	)
	if !errors.Is(err, flightstate.ErrNotFound) {
		t.Fatalf(
			"error = %v, want flight state not found",
			err,
		)
	}
}

func TestNewServiceRequiresReader(
	t *testing.T,
) {
	_, err := NewService(ServiceConfig{})
	if !errors.Is(
		err,
		ErrLatestStateReaderRequired,
	) {
		t.Fatalf(
			"error = %v, want reader required",
			err,
		)
	}
}
