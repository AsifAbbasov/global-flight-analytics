package sharedsnapshot

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
)

func TestStoreRejectsOlderCycleThatFinishesLater(
	t *testing.T,
) {
	store := NewStore()

	olderCycleStartedAt := time.Date(
		2026,
		time.July,
		5,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	newerCycleStartedAt := olderCycleStartedAt.Add(
		time.Minute,
	)

	newerSnapshot, err := FromEnvelopeForCycle(
		newerCycleStartedAt,
		newerCycleStartedAt.Add(
			time.Minute,
		),
		providerfanin.Envelope[Payload]{
			Status: providerfanin.BatchStatusSucceeded,

			TotalCount:   1,
			SuccessCount: 1,

			Successes: []providerfanin.Success[Payload]{
				{
					TaskID:     TaskIDRegionalTraffic,
					RequestKey: "newer-cycle",
					Value: NewRegionalTrafficPayload(
						[]flightstate.FlightState{},
					),
				},
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create newer cycle snapshot: %v",
			err,
		)
	}

	olderSnapshot, err := FromEnvelopeForCycle(
		olderCycleStartedAt,
		newerCycleStartedAt.Add(
			2*time.Minute,
		),
		providerfanin.Envelope[Payload]{
			Status: providerfanin.BatchStatusSucceeded,

			TotalCount:   1,
			SuccessCount: 1,

			Successes: []providerfanin.Success[Payload]{
				{
					TaskID:     TaskIDRegionalTraffic,
					RequestKey: "older-cycle",
					Value: NewRegionalTrafficPayload(
						[]flightstate.FlightState{},
					),
				},
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create older cycle snapshot: %v",
			err,
		)
	}

	if err := store.Publish(
		newerSnapshot,
	); err != nil {
		t.Fatalf(
			"publish newer cycle snapshot: %v",
			err,
		)
	}

	err = store.Publish(
		olderSnapshot,
	)
	if err == nil {
		t.Fatal(
			"expected older cycle snapshot to be rejected",
		)
	}

	if !errors.Is(
		err,
		ErrSnapshotOlderThanCurrent,
	) {
		t.Fatalf(
			"expected ErrSnapshotOlderThanCurrent, got %v",
			err,
		)
	}

	current, exists := store.Current()
	if !exists {
		t.Fatal(
			"expected current snapshot",
		)
	}

	if len(current.Successes) != 1 {
		t.Fatalf(
			"unexpected current success count: got %d, want 1",
			len(current.Successes),
		)
	}

	if current.Successes[0].TaskID !=
		TaskIDRegionalTraffic {
		t.Fatalf(
			"unexpected current snapshot task identifier: %q",
			current.Successes[0].TaskID,
		)
	}

	if current.Successes[0].RequestKey !=
		"newer-cycle" {
		t.Fatalf(
			"unexpected current snapshot request key: got %q, want %q",
			current.Successes[0].RequestKey,
			"newer-cycle",
		)
	}
}
