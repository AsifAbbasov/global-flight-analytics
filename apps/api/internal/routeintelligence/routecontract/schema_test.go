package routecontract

import (
	"reflect"
	"testing"
)

func TestCurrentSchemaIsDeterministicAndUnique(
	t *testing.T,
) {
	first := CurrentSchema()
	second := CurrentSchema()

	if first.Version != SchemaVersionV1 {
		t.Fatalf(
			"schema version = %q",
			first.Version,
		)
	}
	if len(first.Definitions) != 22 {
		t.Fatalf(
			"definitions = %d, want 22",
			len(first.Definitions),
		)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("CurrentSchema() is not deterministic")
	}

	seen := make(map[string]struct{})
	for index, definition := range first.Definitions {
		if definition.Name == "" ||
			definition.Group == "" ||
			definition.ValueType == "" ||
			definition.Description == "" {
			t.Fatalf(
				"definition %d is incomplete: %#v",
				index,
				definition,
			)
		}
		if _, exists := seen[definition.Name]; exists {
			t.Fatalf(
				"duplicate definition %q",
				definition.Name,
			)
		}
		seen[definition.Name] = struct{}{}
	}
}

func TestCurrentSchemaReturnsDefensiveCopy(
	t *testing.T,
) {
	schema := CurrentSchema()
	schema.Definitions[0].Name = "changed"

	next := CurrentSchema()
	if next.Definitions[0].Name == "changed" {
		t.Fatal(
			"CurrentSchema() shared definitions",
		)
	}
}

func TestDefinitionByName(t *testing.T) {
	definition, ok := DefinitionByName(
		"confidence.score",
	)
	if !ok {
		t.Fatal(
			"confidence.score definition not found",
		)
	}
	if definition.Group !=
		FieldGroupConfidence ||
		definition.ValueType !=
			FieldValueTypeFloat64 ||
		definition.Unit != "ratio" ||
		!definition.Required {
		t.Fatalf(
			"unexpected definition: %#v",
			definition,
		)
	}

	if _, ok := DefinitionByName(
		"unknown",
	); ok {
		t.Fatal(
			"unexpected unknown definition",
		)
	}
}
