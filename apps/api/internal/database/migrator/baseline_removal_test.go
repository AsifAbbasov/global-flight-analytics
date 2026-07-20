package migrator

import (
	"strings"
	"testing"
)

func TestRunnerDoesNotExposeBaselineOperation(t *testing.T) {
	t.Parallel()

	source := readRunnerSource(t)
	for _, forbidden := range []string{
		"func (runner *Runner) Baseline",
		"begin migration baseline transaction",
		"baseline migration",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("migration runner exposes forbidden baseline token %q", forbidden)
		}
	}
}
