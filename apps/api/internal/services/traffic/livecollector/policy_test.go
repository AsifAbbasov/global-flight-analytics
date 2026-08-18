package livecollector

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestAirplanesLiveMinimumRequestSpacingUsesDailyFreeBudget(t *testing.T) {
	spacing, err := MinimumRequestSpacing(
		providerpolicy.ProviderAirplanesLive,
	)
	if err != nil {
		t.Fatal(err)
	}

	want := 24 * time.Hour / 500
	if spacing != want {
		t.Fatalf("spacing = %s, want %s", spacing, want)
	}
	if spacing != 172800*time.Millisecond {
		t.Fatalf("spacing = %s, want 172.8s", spacing)
	}
}

func TestAirplanesLivePollIntervalScalesWithTargetCount(t *testing.T) {
	interval, err := MinimumPollInterval(
		providerpolicy.ProviderAirplanesLive,
		3,
	)
	if err != nil {
		t.Fatal(err)
	}

	want := (24 * time.Hour / 500) * 3
	if interval != want {
		t.Fatalf("interval = %s, want %s", interval, want)
	}
}

func TestMinimumPollIntervalRejectsInvalidTargetCount(t *testing.T) {
	if _, err := MinimumPollInterval(
		providerpolicy.ProviderAirplanesLive,
		0,
	); err == nil {
		t.Fatal("expected invalid target-count error")
	}
}

func TestMinimumRequestSpacingRejectsNonFixedWindowProvider(t *testing.T) {
	if _, err := MinimumRequestSpacing(
		providerpolicy.ProviderOpenSky,
	); err == nil {
		t.Fatal("expected non-fixed-window provider error")
	}
}
