package migrationfile

import "testing"

func TestParseAcceptsCanonicalMigrationFileNames(t *testing.T) {
	testCases := []struct {
		fileName string
		version  string
		name     string
	}{
		{
			fileName: "001_initial_schema.sql",
			version:  "001",
			name:     "initial_schema",
		},
		{
			fileName: "018_trajectory_relational_integrity.sql",
			version:  "018",
			name:     "trajectory_relational_integrity",
		},
		{
			fileName: "019_данные_2.sql",
			version:  "019",
			name:     "данные_2",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.fileName, func(t *testing.T) {
			identity, err := Parse(testCase.fileName)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if identity.Version != testCase.version ||
				identity.Name != testCase.name ||
				identity.FileName != testCase.fileName {
				t.Fatalf("Parse() identity = %#v", identity)
			}
		})
	}
}

func TestParseRejectsNonCanonicalMigrationFileNames(t *testing.T) {
	for _, fileName := range []string{
		"",
		" 001_initial_schema.sql",
		"001_initial_schema.sql ",
		"database/001_initial_schema.sql",
		`database\\001_initial_schema.sql`,
		"001_initial_schema.SQL",
		"10_short.sql",
		"ABC_letters.sql",
		"١٢٣_unicode_digits.sql",
		"010.sql",
		"010_.sql",
		"010_invalid-name.sql",
	} {
		t.Run(fileName, func(t *testing.T) {
			if _, err := Parse(fileName); err == nil {
				t.Fatalf("Parse(%q) expected error", fileName)
			}
		})
	}
}

func TestMustParsePanicsForInvalidSourceOwnedName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustParse() did not panic")
		}
	}()

	MustParse("invalid.sql")
}
