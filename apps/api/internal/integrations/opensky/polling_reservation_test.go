package opensky

import (
	"errors"
	"testing"
	"time"
)

func TestPollingReservationCanBeReleasedBeforeTransportAttempt(
	t *testing.T,
) {
	client := &Client{
		config: Config{
			PollingInterval: 10 * time.Second,
		},
	}

	reservedAt, err := client.reservePollSlot()
	if err != nil {
		t.Fatalf(
			"reserve first polling slot: %v",
			err,
		)
	}
	client.releasePollSlot(
		reservedAt,
	)

	if _, err := client.reservePollSlot(); err != nil {
		t.Fatalf(
			"reserve polling slot after release: %v",
			err,
		)
	}
}

func TestPollingTooSoonErrorCarriesRetryEvidence(
	t *testing.T,
) {
	client := &Client{
		config: Config{
			PollingInterval: time.Minute,
		},
	}

	firstAt, err := client.reservePollSlot()
	if err != nil {
		t.Fatalf(
			"reserve first polling slot: %v",
			err,
		)
	}
	_, err = client.reservePollSlot()
	if err == nil {
		t.Fatal(
			"expected polling cooldown error",
		)
	}
	if !errors.Is(
		err,
		ErrPollingTooSoon,
	) {
		t.Fatalf(
			"expected ErrPollingTooSoon, got %v",
			err,
		)
	}

	var cooldown *PollingTooSoonError
	if !errors.As(
		err,
		&cooldown,
	) {
		t.Fatalf(
			"expected PollingTooSoonError, got %v",
			err,
		)
	}
	wantRetryAt := firstAt.Add(
		time.Minute,
	)
	if !cooldown.RetryAtTime().Equal(
		wantRetryAt,
	) {
		t.Fatalf(
			"retry at = %s, want %s",
			cooldown.RetryAtTime(),
			wantRetryAt,
		)
	}
	if cooldown.ExternalRequestAttempted() {
		t.Fatal(
			"polling cooldown must not report an HTTP attempt",
		)
	}
}
