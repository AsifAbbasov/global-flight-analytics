package snapshot

import (
	"testing"
	"time"
)

func TestStoreSaveAndLoad(t *testing.T) {
	store := NewStore()

	expected := Snapshot{
		Time:                 time.Now().UTC(),
		ActiveAircraft:       153,
		AreaSquareKilometers: 86_600,
		ObservedSamples:      75,
		ExpectedSamples:      100,
	}

	store.Save(expected)

	actual := store.Load()

	if !actual.Time.Equal(expected.Time) {
		t.Fatal("unexpected snapshot time")
	}

	if actual.ActiveAircraft != expected.ActiveAircraft {
		t.Fatal("unexpected active aircraft count")
	}

	if actual.AreaSquareKilometers != expected.AreaSquareKilometers {
		t.Fatal("unexpected area")
	}

	if actual.ObservedSamples != expected.ObservedSamples {
		t.Fatal("unexpected observed sample count")
	}

	if actual.ExpectedSamples != expected.ExpectedSamples {
		t.Fatal("unexpected expected sample count")
	}
}

func TestNewStoreStartsWithZeroSnapshot(t *testing.T) {
	store := NewStore()
	actual := store.Load()

	if !actual.Time.IsZero() {
		t.Fatal("new store must start with zero snapshot time")
	}

	if actual.ActiveAircraft != 0 {
		t.Fatal("new store must start with zero active aircraft count")
	}

	if actual.AreaSquareKilometers != 0 {
		t.Fatal("new store must start with zero area")
	}

	if actual.ObservedSamples != 0 {
		t.Fatal("new store must start with zero observed sample count")
	}

	if actual.ExpectedSamples != 0 {
		t.Fatal("new store must start with zero expected sample count")
	}
}
