package flightfeatures

import (
	"reflect"
	"strings"
	"testing"
)

func TestCurrentSchemaIsVersionedAndDeterministic(t *testing.T) {
	first := CurrentSchema()
	second := CurrentSchema()

	if first.Version != SchemaVersionV1 {
		t.Fatalf(
			"schema version = %q, want %q",
			first.Version,
			SchemaVersionV1,
		)
	}
	if len(first.Definitions) == 0 {
		t.Fatal("schema definitions must not be empty")
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("CurrentSchema() is not deterministic")
	}
}

func TestCurrentSchemaReturnsIndependentDefinitions(t *testing.T) {
	first := CurrentSchema()
	first.Definitions[0].Name = "changed"

	second := CurrentSchema()
	if second.Definitions[0].Name == "changed" {
		t.Fatal(
			"CurrentSchema() exposed mutable package schema state",
		)
	}
}

func TestCurrentSchemaDefinitionsAreUniqueAndComplete(
	t *testing.T,
) {
	schema := CurrentSchema()
	names := make(map[string]struct{}, len(schema.Definitions))
	groupCounts := map[FeatureGroup]int{}

	for index, definition := range schema.Definitions {
		if strings.TrimSpace(definition.Name) == "" {
			t.Fatalf("definition %d has an empty name", index)
		}
		if _, exists := names[definition.Name]; exists {
			t.Fatalf(
				"duplicate feature definition %q",
				definition.Name,
			)
		}
		names[definition.Name] = struct{}{}

		if definition.Group == "" {
			t.Fatalf(
				"feature %q has an empty group",
				definition.Name,
			)
		}
		if definition.ValueType == "" {
			t.Fatalf(
				"feature %q has an empty value type",
				definition.Name,
			)
		}
		if strings.TrimSpace(definition.Description) == "" {
			t.Fatalf(
				"feature %q has an empty description",
				definition.Name,
			)
		}
		if !strings.HasPrefix(
			definition.Name,
			string(definition.Group)+".",
		) {
			t.Fatalf(
				"feature %q does not match group %q",
				definition.Name,
				definition.Group,
			)
		}

		groupCounts[definition.Group]++
	}

	expectedGroups := []FeatureGroup{
		FeatureGroupTemporal,
		FeatureGroupGeographical,
		FeatureGroupOperational,
		FeatureGroupTrajectory,
		FeatureGroupAircraft,
	}
	for _, group := range expectedGroups {
		if groupCounts[group] == 0 {
			t.Fatalf(
				"schema has no definitions for group %q",
				group,
			)
		}
	}
}

func TestRequiredFeatureDefinitionsHaveUnitsWhenApplicable(
	t *testing.T,
) {
	unitlessValueTypes := map[FeatureValueType]bool{
		FeatureValueTypeBoolean: true,
		FeatureValueTypeString:  true,
	}

	for _, definition := range CurrentSchema().Definitions {
		if !definition.Required {
			continue
		}
		if unitlessValueTypes[definition.ValueType] {
			continue
		}
		if strings.TrimSpace(definition.Unit) == "" {
			t.Fatalf(
				"required numeric feature %q has no unit",
				definition.Name,
			)
		}
	}
}

func TestDefinitionByNameReturnsKnownDefinition(t *testing.T) {
	definition, found := DefinitionByName(
		"trajectory.quality_score",
	)
	if !found {
		t.Fatal("expected trajectory.quality_score definition")
	}
	if definition.Group != FeatureGroupTrajectory ||
		definition.ValueType != FeatureValueTypeFloat64 ||
		definition.Unit != "ratio" ||
		!definition.Required {
		t.Fatalf(
			"unexpected definition: %#v",
			definition,
		)
	}
}

func TestDefinitionByNameRejectsUnknownDefinition(t *testing.T) {
	definition, found := DefinitionByName(
		"trajectory.unknown_feature",
	)
	if found {
		t.Fatalf(
			"unexpected definition: %#v",
			definition,
		)
	}
}
