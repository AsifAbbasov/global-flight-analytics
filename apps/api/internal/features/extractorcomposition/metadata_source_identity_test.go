package extractorcomposition

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/aircraftprovider"
)

func TestProcessingIdentityIncludesAircraftMetadataSource(t *testing.T) {
	composition, err := New(DefaultConfig(&fakeAircraftLookup{}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if composition.ProcessingIdentity.AircraftMetadataSourceName !=
		aircraftprovider.MetadataSourceName {
		t.Fatalf(
			"aircraft metadata source = %q",
			composition.ProcessingIdentity.AircraftMetadataSourceName,
		)
	}
}
