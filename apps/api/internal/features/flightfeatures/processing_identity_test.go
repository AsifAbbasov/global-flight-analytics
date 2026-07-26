package flightfeatures

import (
	"reflect"
	"testing"
	"time"
)

func TestProcessingIdentityIsValueComparable(t *testing.T) {
	identity := ProcessingIdentity{
		Versions: ProcessingComponentVersions{
			Composition: "composition-v1",
			Extractor:   "extractor-v1",
		},
		GeographicCellPrecision:       2,
		AircraftEnrichmentMode:        AircraftEnrichmentModeEnabled,
		AircraftCacheMode:             "enabled",
		AircraftPositiveCacheTTL:      time.Minute,
		AircraftNegativeCacheTTL:      30 * time.Second,
		AircraftNotFoundPolicyVersion: "not-found-v1",
		AircraftMetadataSourceName:    "aircraft-reference",
	}
	copied := identity
	if !reflect.DeepEqual(copied, identity) {
		t.Fatalf("processing identity copy changed value: got=%#v want=%#v", copied, identity)
	}
}
