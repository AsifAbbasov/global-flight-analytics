package extractor

import (
	"reflect"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestCanonicalFingerprintInputFieldOrderRemainsExplicit(t *testing.T) {
	assertStructFieldNames(
		t,
		reflect.TypeOf(canonicalFingerprintInput{}),
		[]string{
			"ProcessingIdentity",
			"AircraftMetadataSourceName",
			"AircraftMetadataProviderVersion",
			"Trajectory",
			"Aircraft",
		},
	)
}

func TestCanonicalFingerprintEvidenceMirrorsDomainContracts(t *testing.T) {
	tests := []struct {
		name      string
		source    reflect.Type
		canonical reflect.Type
		renames   map[string]string
	}{
		{
			name:      "trajectory",
			source:    reflect.TypeOf(trajectory.FlightTrajectory{}),
			canonical: reflect.TypeOf(canonicalTrajectory{}),
		},
		{
			name:      "track point",
			source:    reflect.TypeOf(trajectory.TrackPoint4D{}),
			canonical: reflect.TypeOf(canonicalTrackPoint{}),
		},
		{
			name:      "segment",
			source:    reflect.TypeOf(trajectory.TrajectorySegment{}),
			canonical: reflect.TypeOf(canonicalSegment{}),
		},
		{
			name:      "coverage gap",
			source:    reflect.TypeOf(trajectory.CoverageGap{}),
			canonical: reflect.TypeOf(canonicalCoverageGap{}),
			renames: map[string]string{
				"DistanceKm": "DistanceKM",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.source.NumField() != test.canonical.NumField() {
				t.Fatalf("field count changed: source=%d canonical=%d", test.source.NumField(), test.canonical.NumField())
			}
			for index := 0; index < test.source.NumField(); index++ {
				sourceName := test.source.Field(index).Name
				wantName := sourceName
				if renamed, exists := test.renames[sourceName]; exists {
					wantName = renamed
				}
				gotName := test.canonical.Field(index).Name
				if gotName != wantName {
					t.Fatalf("field %d = %q, want %q mirrored from %q", index, gotName, wantName, sourceName)
				}
			}
		})
	}
}

func TestTrajectoryClonePreservesFingerprintInput(t *testing.T) {
	item := validRequest().Trajectory
	original, err := fingerprintTrajectory(item)
	if err != nil {
		t.Fatalf("fingerprint original: %v", err)
	}
	cloned, err := fingerprintTrajectory(item.Clone())
	if err != nil {
		t.Fatalf("fingerprint clone: %v", err)
	}
	if original != cloned {
		t.Fatalf("clone changed fingerprint: original=%q cloned=%q", original, cloned)
	}
}

func assertStructFieldNames(t *testing.T, item reflect.Type, want []string) {
	t.Helper()
	if item.NumField() != len(want) {
		t.Fatalf("field count = %d, want %d", item.NumField(), len(want))
	}
	for index, wantName := range want {
		if got := item.Field(index).Name; got != wantName {
			t.Fatalf("field %d = %q, want %q", index, got, wantName)
		}
	}
}
