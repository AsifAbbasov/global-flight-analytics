package migrator

import (
	"os"
	"strings"
	"testing"
)

func TestPrepareMigrationSQLRemovesOuterTransactionEnvelope(
	t *testing.T,
) {
	t.Parallel()

	actual, err := prepareMigrationSQL(`
		BEGIN;

		CREATE TABLE example (
			id bigint PRIMARY KEY
		);

		COMMIT;
	`)
	if err != nil {
		t.Fatalf("prepare migration SQL: %v", err)
	}

	expected := `CREATE TABLE example (
			id bigint PRIMARY KEY
		);`
	if actual != expected {
		t.Fatalf(
			"unexpected migration body\nexpected:\n%s\nactual:\n%s",
			expected,
			actual,
		)
	}
}

func TestPrepareMigrationSQLPreservesUnwrappedSQL(
	t *testing.T,
) {
	t.Parallel()

	expected := "CREATE INDEX example_idx ON example (id);"
	actual, err := prepareMigrationSQL(expected)
	if err != nil {
		t.Fatalf("prepare migration SQL: %v", err)
	}

	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

func TestPrepareMigrationSQLRejectsIncompleteTransactionEnvelope(
	t *testing.T,
) {
	t.Parallel()

	cases := []string{
		"BEGIN; CREATE TABLE example (id bigint);",
		"CREATE TABLE example (id bigint); COMMIT;",
	}

	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase, func(t *testing.T) {
			t.Parallel()

			if _, err := prepareMigrationSQL(testCase); err == nil {
				t.Fatal("expected incomplete transaction envelope to fail")
			}
		})
	}
}

func TestPrepareMigrationSQLRejectsNestedTransactionControl(
	t *testing.T,
) {
	t.Parallel()

	_, err := prepareMigrationSQL(`
		BEGIN;
		CREATE TABLE example (id bigint);
		COMMIT;
		BEGIN;
		CREATE TABLE second_example (id bigint);
		COMMIT;
	`)
	if err == nil {
		t.Fatal("expected nested transaction control to fail")
	}
}

func TestApplyMigrationAtomicallyKeepsDDLAndHistoryInOneTransaction(
	t *testing.T,
) {
	t.Parallel()

	source := readRunnerSource(t)
	function := extractFunctionSource(
		t,
		source,
		"func (\n\trunner *Runner,\n) applyMigrationAtomically",
	)

	requireTokensInOrder(
		t,
		function,
		"conn.BeginTx(",
		"tx.Exec(ctx, sqlBody)",
		"INSERT INTO schema_migrations",
		"tx.Commit(ctx)",
	)

	if strings.Contains(function, "runner.pool.Exec") {
		t.Fatal("atomic migration function bypasses its transaction")
	}
}

func TestApplyPendingUsesAProcessWidePostgreSQLAdvisoryLock(
	t *testing.T,
) {
	t.Parallel()

	source := readRunnerSource(t)
	applyPending := extractFunctionSource(
		t,
		source,
		"func (runner *Runner) ApplyPending",
	)
	if !strings.Contains(applyPending, "runner.withMigrationLock(") {
		t.Fatal("ApplyPending does not use the migration advisory lock")
	}

	lockFunction := extractFunctionSource(
		t,
		source,
		"func (\n\trunner *Runner,\n) withMigrationLock",
	)
	if !strings.Contains(lockFunction, "pg_advisory_lock") {
		t.Fatal("migration lock function does not acquire pg_advisory_lock")
	}

	unlockFunction := extractFunctionSource(
		t,
		source,
		"func releaseMigrationLock",
	)
	if !strings.Contains(unlockFunction, "pg_advisory_unlock") {
		t.Fatal("migration lock function does not release pg_advisory_unlock")
	}
}

func readRunnerSource(t *testing.T) string {
	t.Helper()

	content, err := os.ReadFile("runner.go")
	if err != nil {
		t.Fatalf("read runner.go: %v", err)
	}

	return string(content)
}

func extractFunctionSource(
	t *testing.T,
	source string,
	functionPrefix string,
) string {
	t.Helper()

	start := strings.Index(source, functionPrefix)
	if start < 0 {
		t.Fatalf("function prefix %q not found", functionPrefix)
	}

	openingOffset := strings.Index(source[start:], "{")
	if openingOffset < 0 {
		t.Fatalf("opening brace for %q not found", functionPrefix)
	}

	opening := start + openingOffset
	depth := 0
	for index := opening; index < len(source); index++ {
		switch source[index] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return source[start : index+1]
			}
		}
	}

	t.Fatalf("closing brace for %q not found", functionPrefix)
	return ""
}

func requireTokensInOrder(
	t *testing.T,
	source string,
	tokens ...string,
) {
	t.Helper()

	remaining := source
	for _, token := range tokens {
		index := strings.Index(remaining, token)
		if index < 0 {
			t.Fatalf("required token %q not found in order", token)
		}

		remaining = remaining[index+len(token):]
	}
}
