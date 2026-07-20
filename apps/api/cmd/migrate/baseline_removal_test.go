package main

import (
	"os"
	"strings"
	"testing"
)

func TestMigrateCommandDoesNotExposeBaselineMode(t *testing.T) {
	t.Parallel()

	sourceBytes, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read migrate command source: %v", err)
	}

	source := string(sourceBytes)
	for _, forbidden := range []string{
		`"baseline"`,
		"runner.Baseline(",
		"record existing migrations as applied without executing SQL",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("migrate command exposes forbidden baseline token %q", forbidden)
		}
	}
}
