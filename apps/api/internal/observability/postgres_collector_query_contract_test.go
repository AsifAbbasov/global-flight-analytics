package observability

import (
	"strings"
	"testing"
)

func TestReconciliationOldestPendingAgeQueryMeasuresDueWaitTime(t *testing.T) {
	t.Parallel()

	for _, expected := range []string{
		"MIN(next_attempt_at)",
		"status = 'pending'",
		"next_attempt_at <= now()",
	} {
		if !strings.Contains(reconciliationOldestPendingAgeQuery, expected) {
			t.Fatalf("expected reconciliation backlog query to contain %q", expected)
		}
	}

	if strings.Contains(reconciliationOldestPendingAgeQuery, "MIN(created_at)") {
		t.Fatal("reconciliation backlog age must not use immutable row creation time")
	}
}
