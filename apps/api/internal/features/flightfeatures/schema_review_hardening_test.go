package flightfeatures

import (
	"reflect"
	"sort"
	"testing"
)

func TestCurrentSchemaMatchesExactGroupFieldContract(t *testing.T) {
	counts := CurrentGroupRequirementCounts()
	want := map[FeatureGroup]GroupRequirementCounts{
		FeatureGroupTemporal: {
			Required: TemporalRequiredFeatureFieldCount,
		},
		FeatureGroupGeographical: {
			Required: GeographicalRequiredFeatureFieldCount,
		},
		FeatureGroupOperational: {
			Required: OperationalRequiredFeatureFieldCount,
		},
		FeatureGroupTrajectory: {
			Required: TrajectoryRequiredFeatureFieldCount,
		},
		FeatureGroupAircraft: {
			Optional: AircraftOptionalFeatureFieldCount,
		},
	}
	if !reflect.DeepEqual(counts, want) {
		t.Fatalf("group requirement counts = %#v, want %#v", counts, want)
	}
}

func TestGeographicalSchemaContainsEveryAnalyticalField(t *testing.T) {
	want := []string{
		"geographical.crosses_antimeridian",
		"geographical.end_latitude",
		"geographical.end_longitude",
		"geographical.great_circle_distance_km",
		"geographical.latitude_span_degrees",
		"geographical.longitude_span_degrees",
		"geographical.maximum_displacement_km",
		"geographical.maximum_latitude",
		"geographical.maximum_longitude",
		"geographical.minimum_latitude",
		"geographical.minimum_longitude",
		"geographical.observed_path_distance_km",
		"geographical.start_latitude",
		"geographical.start_longitude",
		"geographical.unique_geographic_cell_count",
	}
	got := make([]string, 0, len(want))
	for _, definition := range CurrentSchema().Definitions {
		if definition.Group == FeatureGroupGeographical {
			got = append(got, definition.Name)
		}
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("geographical definitions = %#v, want %#v", got, want)
	}
	if _, found := DefinitionByName(
		"geographical.geographic_cell_precision",
	); found {
		t.Fatal("processing precision must not be counted as an analytical feature")
	}
}

func TestSchemaRegistrySupportsExplicitVersionLookup(t *testing.T) {
	versions := SupportedSchemaVersions()
	if !reflect.DeepEqual(versions, []SchemaVersion{SchemaVersionV1}) {
		t.Fatalf("supported schema versions = %#v", versions)
	}
	versions[0] = "mutated"
	if !reflect.DeepEqual(
		SupportedSchemaVersions(),
		[]SchemaVersion{SchemaVersionV1},
	) {
		t.Fatal("SupportedSchemaVersions() exposed mutable state")
	}

	schema, found := SchemaForVersion(SchemaVersionV1)
	if !found || !reflect.DeepEqual(schema, CurrentSchema()) {
		t.Fatalf("SchemaForVersion(v1) = %#v, found=%v", schema, found)
	}
	if _, found := SchemaForVersion("future-schema"); found {
		t.Fatal("unknown schema version was accepted")
	}
	if CompatibilityBetween(SchemaVersionV1, SchemaVersionV1) !=
		SchemaCompatibilityIdentical {
		t.Fatal("identical registered schemas were not classified as identical")
	}
	if CompatibilityBetween(SchemaVersionV1, "future-schema") !=
		SchemaCompatibilityUnsupported {
		t.Fatal("unknown schema compatibility was not rejected")
	}
}

func TestStringEnumerationsExposeCentralValidityContracts(t *testing.T) {
	if !AvailabilityStatusAvailable.IsValid() ||
		AvailabilityStatus("maybe").IsValid() ||
		!ValidationStatusUnvalidated.IsValid() ||
		ValidationStatus("approved").IsValid() ||
		!FeatureGroupGeographical.IsValid() ||
		FeatureGroup("unknown").IsValid() ||
		!FeatureValueTypeFloat64.IsValid() ||
		FeatureValueType("decimal").IsValid() ||
		!ValidationAuditStateComplete.IsValid() ||
		ValidationAuditState("unknown").IsValid() ||
		!ValidationIssueSeverityWarning.IsValid() ||
		ValidationIssueSeverity("notice").IsValid() ||
		!AircraftEnrichmentModeEnabled.IsValid() ||
		AircraftEnrichmentMode("automatic").IsValid() {
		t.Fatal("enumeration validity contract is inconsistent")
	}
}

func TestGeographicCellPrecisionLivesInProcessingIdentity(t *testing.T) {
	identity := ProcessingIdentity{GeographicCellPrecision: 3}
	if identity.GeographicCellPrecision != 3 {
		t.Fatalf("processing precision = %d", identity.GeographicCellPrecision)
	}
}
