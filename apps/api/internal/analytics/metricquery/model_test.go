package metricquery

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestRecentRequestNormalizeDefaults(t *testing.T) {
	now := time.Date(2026, time.July, 14, 10, 0, 0, 0, time.UTC)

	window, err := (RecentRequest{}).Normalize(now)
	if err != nil {
		t.Fatalf("expected default window, got %v", err)
	}

	if window.ObservedTo != now ||
		window.ObservedFrom != now.Add(-15*time.Minute) ||
		window.Limit != DefaultResultLimit {
		t.Fatalf("unexpected normalized window: %#v", window)
	}
}

func TestRecentRequestNormalizeRejectsInvalidValues(t *testing.T) {
	_, err := (RecentRequest{WindowMinutes: 181}).Normalize(time.Now())
	if !errors.Is(err, ErrWindowMinutesInvalid) {
		t.Fatalf("expected window error, got %v", err)
	}

	_, err = (RecentRequest{Limit: 5001}).Normalize(time.Now())
	if !errors.Is(err, ErrResultLimitInvalid) {
		t.Fatalf("expected limit error, got %v", err)
	}
}

func TestNormalizeTrajectoryIDs(t *testing.T) {
	ids, err := NormalizeTrajectoryIDs([]string{"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222", "11111111-1111-4111-8111-111111111111"})
	if err != nil {
		t.Fatalf("expected normalized ids, got %v", err)
	}

	expected := []string{"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"}
	if !reflect.DeepEqual(ids, expected) {
		t.Fatalf("expected %#v, got %#v", expected, ids)
	}
}

func TestNormalizeTrajectoryIDsRejectsMissingAndBlankValues(t *testing.T) {
	_, err := NormalizeTrajectoryIDs(nil)
	if !errors.Is(err, ErrTrajectoryIDsMissing) {
		t.Fatalf("expected missing ids error, got %v", err)
	}

	_, err = NormalizeTrajectoryIDs([]string{"11111111-1111-4111-8111-111111111111", " "})
	if !errors.Is(err, ErrTrajectoryIDInvalid) {
		t.Fatalf("expected invalid id error, got %v", err)
	}
}
