package historicalcontract

import "testing"

func TestCurrentSchemaIsStableAndDefensive(
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
	if len(first.Definitions) != 36 {
		t.Fatalf(
			"definition count = %d, want 36",
			len(first.Definitions),
		)
	}

	seen := make(map[string]struct{})
	for _, definition := range first.Definitions {
		if definition.Name == "" ||
			definition.Description == "" {
			t.Fatalf(
				"incomplete definition: %#v",
				definition,
			)
		}
		if _, exists := seen[definition.Name]; exists {
			t.Fatalf(
				"duplicate definition: %s",
				definition.Name,
			)
		}
		seen[definition.Name] = struct{}{}
	}

	first.Definitions[0].Name = "changed"
	if second.Definitions[0].Name == "changed" {
		t.Fatal(
			"CurrentSchema() returned shared definitions",
		)
	}
}

func TestDefinitionByName(
	t *testing.T,
) {
	definition, exists := DefinitionByName(
		"points.coverage_ratio",
	)
	if !exists {
		t.Fatal(
			"points.coverage_ratio definition was not found",
		)
	}
	if definition.Group != FieldGroupSeries ||
		definition.ValueType !=
			FieldValueTypeFloat64 ||
		definition.Unit != "ratio" ||
		!definition.Required {
		t.Fatalf(
			"unexpected definition: %#v",
			definition,
		)
	}

	if _, exists := DefinitionByName(
		"missing",
	); exists {
		t.Fatal(
			"missing definition unexpectedly exists",
		)
	}
}
