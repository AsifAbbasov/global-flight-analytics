package postgres

import (
	"os"
	"strings"
	"testing"
)

func TestAirportPaginationContractRemainsKeysetBounded(t *testing.T) {
	t.Parallel()

	queries := readAirportPaginationSource(t, "airport_read_queries.go")
	for _, required := range []string{
		"ORDER BY a.name ASC, a.id ASC",
		"a.name > $1",
		"a.name = $1",
		"a.id > $2::uuid",
		"LIMIT $1",
		"LIMIT $3",
	} {
		if !strings.Contains(queries, required) {
			t.Fatalf("airport keyset queries are missing %q", required)
		}
	}
	if strings.Contains(queries, "OFFSET") {
		t.Fatal("airport pagination must not use offset pagination")
	}

	pagination := readAirportPaginationSource(t, "airport_pagination_read.go")
	for _, required := range []string{
		"normalized.Limit + 1",
		"scanAirportRecord(rows)",
		"buildAirportPage(records, normalized.Limit)",
	} {
		if !strings.Contains(pagination, required) {
			t.Fatalf("airport page reader is missing %q", required)
		}
	}
}

func TestAirportReadPathsShareOneRowScanner(t *testing.T) {
	t.Parallel()

	repository := readAirportPaginationSource(t, "airport_repository.go")
	if !strings.Contains(repository, "repository.ListPage(ctx, request)") {
		t.Fatal("legacy List does not delegate to bounded ListPage")
	}
	if !strings.Contains(repository, "scanAirportRecord(") {
		t.Fatal("GetByICAO does not use the canonical airport scanner")
	}
	if strings.Contains(repository, "rows.Scan(") ||
		strings.Contains(repository, "ORDER BY a.name ASC;") {
		t.Fatal("airport repository regained duplicated scan or unbounded SQL ownership")
	}

	scanner := readAirportPaginationSource(t, "airport_row_scan.go")
	if strings.Count(scanner, "func scanAirportRecord(") != 1 {
		t.Fatal("airport row scanner must have exactly one owner")
	}
}

func readAirportPaginationSource(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(content)
}
