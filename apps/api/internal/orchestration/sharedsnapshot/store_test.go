package sharedsnapshot

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
)

func TestStoreCurrentReturnsFalseBeforeFirstPublication(
	t *testing.T,
) {
	store := NewStore()

	_, exists := store.Current()
	if exists {
		t.Fatal(
			"expected empty store before first publication",
		)
	}
}

func TestNilStoreCurrentReturnsFalse(
	t *testing.T,
) {
	var store *Store

	_, exists := store.Current()
	if exists {
		t.Fatal(
			"expected nil store to report no current snapshot",
		)
	}
}

func TestStorePublishProtectsCurrentSnapshotFromCallerMutation(
	t *testing.T,
) {
	store := NewStore()

	snapshot := Snapshot{
		AssembledAt: time.Date(
			2026,
			time.July,
			5,
			12,
			0,
			0,
			0,
			time.UTC,
		),

		Status: providerfanin.BatchStatusSucceeded,

		TotalCount:   1,
		SuccessCount: 1,

		Successes: []providerfanin.Success{
			{
				TaskID:     "traffic",
				RequestKey: "regional-traffic",
				Value:      "traffic-value",
				Shared:     true,
			},
		},
	}

	if err := store.Publish(snapshot); err != nil {
		t.Fatalf(
			"publish shared snapshot: %v",
			err,
		)
	}

	snapshot.Successes[0].TaskID = "caller-mutated"

	current, exists := store.Current()
	if !exists {
		t.Fatal(
			"expected current snapshot after publication",
		)
	}

	if current.Successes[0].TaskID != "traffic" {
		t.Fatal(
			"expected stored snapshot to be protected from caller mutation",
		)
	}

	current.Successes[0].TaskID = "reader-mutated"

	currentAgain, exists := store.Current()
	if !exists {
		t.Fatal(
			"expected current snapshot after reader mutation",
		)
	}

	if currentAgain.Successes[0].TaskID != "traffic" {
		t.Fatal(
			"expected stored snapshot to be protected from reader mutation",
		)
	}
}

func TestStorePublishRejectsSnapshotOlderThanCurrent(
	t *testing.T,
) {
	store := NewStore()

	currentAssembledAt := time.Date(
		2026,
		time.July,
		5,
		12,
		30,
		0,
		0,
		time.UTC,
	)

	olderAssembledAt := currentAssembledAt.Add(
		-time.Minute,
	)

	currentSnapshot := Snapshot{
		AssembledAt: currentAssembledAt,
		Status:      providerfanin.BatchStatusSucceeded,
		Successes: []providerfanin.Success{
			{
				TaskID: "current",
			},
		},
	}

	olderSnapshot := Snapshot{
		AssembledAt: olderAssembledAt,
		Status:      providerfanin.BatchStatusSucceeded,
		Successes: []providerfanin.Success{
			{
				TaskID: "older",
			},
		},
	}

	if err := store.Publish(currentSnapshot); err != nil {
		t.Fatalf(
			"publish current shared snapshot: %v",
			err,
		)
	}

	err := store.Publish(olderSnapshot)
	if err == nil {
		t.Fatal(
			"expected older shared snapshot to be rejected",
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
			"expected current snapshot to remain published",
		)
	}

	if !current.AssembledAt.Equal(
		currentAssembledAt,
	) {
		t.Fatalf(
			"unexpected current assembled time: got %s, want %s",
			current.AssembledAt,
			currentAssembledAt,
		)
	}

	if current.Successes[0].TaskID != "current" {
		t.Fatalf(
			"unexpected current task identifier: %q",
			current.Successes[0].TaskID,
		)
	}
}

func TestStorePublishAcceptsNewerSnapshot(
	t *testing.T,
) {
	store := NewStore()

	firstAssembledAt := time.Date(
		2026,
		time.July,
		5,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	newerAssembledAt := firstAssembledAt.Add(
		time.Minute,
	)

	firstSnapshot := Snapshot{
		AssembledAt: firstAssembledAt,
		Status:      providerfanin.BatchStatusSucceeded,
		Successes: []providerfanin.Success{
			{
				TaskID: "first",
			},
		},
	}

	newerSnapshot := Snapshot{
		AssembledAt: newerAssembledAt,
		Status:      providerfanin.BatchStatusSucceeded,
		Successes: []providerfanin.Success{
			{
				TaskID: "newer",
			},
		},
	}

	if err := store.Publish(firstSnapshot); err != nil {
		t.Fatalf(
			"publish first shared snapshot: %v",
			err,
		)
	}

	if err := store.Publish(newerSnapshot); err != nil {
		t.Fatalf(
			"publish newer shared snapshot: %v",
			err,
		)
	}

	current, exists := store.Current()
	if !exists {
		t.Fatal(
			"expected newer current snapshot",
		)
	}

	if !current.AssembledAt.Equal(
		newerAssembledAt,
	) {
		t.Fatalf(
			"unexpected current assembled time: got %s, want %s",
			current.AssembledAt,
			newerAssembledAt,
		)
	}

	if current.Successes[0].TaskID != "newer" {
		t.Fatalf(
			"unexpected current task identifier: %q",
			current.Successes[0].TaskID,
		)
	}
}

func TestStorePublishRejectsZeroAssembledAt(
	t *testing.T,
) {
	store := NewStore()

	err := store.Publish(
		Snapshot{},
	)
	if err == nil {
		t.Fatal(
			"expected zero assembled time to be rejected",
		)
	}

	if !errors.Is(
		err,
		ErrAssembledAtRequired,
	) {
		t.Fatalf(
			"expected ErrAssembledAtRequired, got %v",
			err,
		)
	}
}
