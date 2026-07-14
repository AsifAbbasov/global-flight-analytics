package migrator

import (
	"errors"
	"reflect"
	"testing"
)

func TestValidateUniqueMigrationVersionsAcceptsUnique(
	t *testing.T,
) {
	err := validateUniqueMigrationVersions(
		[]Migration{
			{
				Version: "001",
				Path:    "001_first.sql",
			},
			{
				Version: "002",
				Path:    "002_second.sql",
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"validateUniqueMigrationVersions() error = %v",
			err,
		)
	}
}

func TestValidateUniqueMigrationVersionsRejectsDuplicate(
	t *testing.T,
) {
	err := validateUniqueMigrationVersions(
		[]Migration{
			{
				Version: "010",
				Path:    "/migrations/010_second.sql",
			},
			{
				Version: "010",
				Path:    "/migrations/010_first.sql",
			},
			{
				Version: "011",
				Path:    "/migrations/011_next.sql",
			},
		},
	)
	if !errors.Is(
		err,
		ErrDuplicateMigrationVersion,
	) {
		t.Fatalf(
			"error = %v, want %v",
			err,
			ErrDuplicateMigrationVersion,
		)
	}

	var duplicateErr *DuplicateMigrationVersionError
	if !errors.As(err, &duplicateErr) {
		t.Fatalf(
			"error = %T, want *DuplicateMigrationVersionError",
			err,
		)
	}
	if duplicateErr.Version != "010" ||
		!reflect.DeepEqual(
			duplicateErr.FileNames,
			[]string{
				"010_first.sql",
				"010_second.sql",
			},
		) {
		t.Fatalf(
			"duplicate error = %#v",
			duplicateErr,
		)
	}
}

func TestValidateUniqueMigrationVersionsReportsLowestVersion(
	t *testing.T,
) {
	err := validateUniqueMigrationVersions(
		[]Migration{
			{Version: "020", Path: "020_b.sql"},
			{Version: "020", Path: "020_a.sql"},
			{Version: "010", Path: "010_b.sql"},
			{Version: "010", Path: "010_a.sql"},
		},
	)

	var duplicateErr *DuplicateMigrationVersionError
	if !errors.As(err, &duplicateErr) {
		t.Fatalf(
			"error = %T, want *DuplicateMigrationVersionError",
			err,
		)
	}
	if duplicateErr.Version != "010" {
		t.Fatalf(
			"version = %q, want 010",
			duplicateErr.Version,
		)
	}
}
