package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
)

func TestNewProviderBudgetStoreValidatesDependencies(t *testing.T) {
	_, err := NewProviderBudgetStore(nil, time.Second)
	if !errors.Is(err, ErrProviderBudgetStorePoolRequired) {
		t.Fatalf(
			"expected ErrProviderBudgetStorePoolRequired, got %v",
			err,
		)
	}
}

func TestNormalizeFixedWindowReservationsSortsAndRejectsDuplicates(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		23,
		16,
		0,
		0,
		0,
		time.UTC,
	)
	normalized, err := normalizeFixedWindowReservations(
		[]providerbudget.FixedWindowReservation{
			{
				LimitIndex:  2,
				WindowStart: now,
				WindowEnd:   now.Add(time.Hour),
				MaxRequests: 10,
			},
			{
				LimitIndex:  0,
				WindowStart: now,
				WindowEnd:   now.Add(time.Second),
				MaxRequests: 1,
			},
		},
	)
	if err != nil {
		t.Fatalf("normalize reservations: %v", err)
	}
	if normalized[0].LimitIndex != 0 ||
		normalized[1].LimitIndex != 2 {
		t.Fatalf("reservations were not sorted: %+v", normalized)
	}

	_, err = normalizeFixedWindowReservations(
		[]providerbudget.FixedWindowReservation{
			{
				LimitIndex:  0,
				WindowStart: now,
				WindowEnd:   now.Add(time.Second),
				MaxRequests: 1,
			},
			{
				LimitIndex:  0,
				WindowStart: now,
				WindowEnd:   now.Add(time.Second),
				MaxRequests: 1,
			},
		},
	)
	if !errors.Is(err, ErrProviderBudgetReservationInvalid) {
		t.Fatalf(
			"expected ErrProviderBudgetReservationInvalid, got %v",
			err,
		)
	}
}
