package historicalaggregate

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestHistoricalAggregatePaginationMatchesCompleteOrdering(
	t *testing.T,
) {
	content := readHistoricalAggregateSource(
		t,
		"postgres.go",
	)

	for _, required := range []string{
		"window_end_unix_nano DESC",
		"window_start_unix_nano DESC",
		"as_of_time_unix_nano DESC",
		"id ASC",
		"window_end_unix_nano < $5",
		"window_start_unix_nano < $6",
		"as_of_time_unix_nano < $7",
		"id > $8",
	} {
		if !strings.Contains(
			content,
			required,
		) {
			t.Fatalf(
				"historical aggregate pagination is missing %q",
				required,
			)
		}
	}

	legacyField := "Before" + "WindowEnd"
	if strings.Contains(
		content,
		"normalized."+legacyField,
	) {
		t.Fatal(
			"legacy single-field pagination remains in PostgreSQL store",
		)
	}
}

func TestHistoricalAggregateContractHasNoLegacyCursorField(
	t *testing.T,
) {
	contract := readHistoricalAggregateContractSource(
		t,
		"contracts.go",
	)
	pagination := readHistoricalAggregateContractSource(
		t,
		"pagination.go",
	)

	legacyField := "Before" + "WindowEnd"
	if strings.Contains(
		contract,
		legacyField,
	) {
		t.Fatal(
			"historical aggregate contract still exposes the legacy cursor field",
		)
	}
	for _, required := range []string{
		"WindowEnd",
		"WindowStart",
		"AsOfTime",
		"ID",
	} {
		if !strings.Contains(
			pagination,
			required,
		) {
			t.Fatalf(
				"composite cursor is missing %q",
				required,
			)
		}
	}
}

func readHistoricalAggregateSource(
	t *testing.T,
	name string,
) string {
	t.Helper()

	_, currentFile, _, ok :=
		runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve historical aggregate source path",
		)
	}
	content, err := os.ReadFile(
		filepath.Join(
			filepath.Dir(currentFile),
			name,
		),
	)
	if err != nil {
		t.Fatalf(
			"read %s: %v",
			name,
			err,
		)
	}
	return string(content)
}

func readHistoricalAggregateContractSource(
	t *testing.T,
	name string,
) string {
	t.Helper()

	_, currentFile, _, ok :=
		runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve historical aggregate contract source path",
		)
	}
	content, err := os.ReadFile(
		filepath.Join(
			filepath.Dir(currentFile),
			"..",
			"historicalaggregatecontract",
			name,
		),
	)
	if err != nil {
		t.Fatalf(
			"read contract %s: %v",
			name,
			err,
		)
	}
	return string(content)
}
