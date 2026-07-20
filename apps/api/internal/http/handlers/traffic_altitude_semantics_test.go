package handlers

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
)

func TestToCurrentTrafficItemsPreservesObservedZeroAltitude(
	t *testing.T,
) {
	value := 0.0

	result := toCurrentTrafficItems(
		[]traffic.CurrentTrafficItem{
			{
				ICAO24:         "abc123",
				AltitudeM:      &value,
				AltitudeStatus: flightstate.AltitudeStatusObserved,
				AltitudeSource: traffic.AltitudeSourceGeometric,
			},
		},
	)

	if len(result) != 1 {
		t.Fatalf(
			"traffic DTO count = %d, want 1",
			len(result),
		)
	}
	if result[0].AltitudeM == nil ||
		*result[0].AltitudeM != 0 {
		t.Fatalf(
			"observed zero altitude was not preserved: %#v",
			result[0].AltitudeM,
		)
	}
	if result[0].AltitudeStatus !=
		flightstate.AltitudeStatusObserved {
		t.Fatalf(
			"altitude status = %q",
			result[0].AltitudeStatus,
		)
	}
	if result[0].AltitudeSource !=
		traffic.AltitudeSourceGeometric {
		t.Fatalf(
			"altitude source = %q",
			result[0].AltitudeSource,
		)
	}
}

func TestToCurrentTrafficItemsPreservesMissingAltitude(
	t *testing.T,
) {
	result := toCurrentTrafficItems(
		[]traffic.CurrentTrafficItem{
			{
				ICAO24:         "def456",
				AltitudeM:      nil,
				AltitudeStatus: flightstate.AltitudeStatusUnknown,
				AltitudeSource: traffic.AltitudeSourceNone,
			},
		},
	)

	if result[0].AltitudeM != nil {
		t.Fatalf(
			"missing altitude became a numeric value: %v",
			*result[0].AltitudeM,
		)
	}
	if result[0].AltitudeStatus !=
		flightstate.AltitudeStatusUnknown {
		t.Fatalf(
			"altitude status = %q",
			result[0].AltitudeStatus,
		)
	}
}
