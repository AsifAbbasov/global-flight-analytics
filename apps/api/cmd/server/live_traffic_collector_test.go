package main

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
)

func TestComposeLiveTrafficCollectorDisabledNeedsNoDatabase(t *testing.T) {
	composition, err := composeLiveTrafficCollector(
		nil,
		nil,
		config.LiveTrafficCollectorConfig{Enabled: false},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if composition.Store != nil || composition.Collector != nil {
		t.Fatalf("unexpected disabled composition: %+v", composition)
	}
}

func TestComposeLiveTrafficCollectorEnabledRequiresDatabase(t *testing.T) {
	_, err := composeLiveTrafficCollector(
		nil,
		nil,
		config.LiveTrafficCollectorConfig{
			Enabled:  true,
			Provider: config.LiveTrafficProviderADSBLOL,
		},
		nil,
	)
	if !errors.Is(err, errLiveTrafficDatabaseRequired) {
		t.Fatalf("expected database requirement, got %v", err)
	}
}
