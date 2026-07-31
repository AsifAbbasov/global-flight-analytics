package projectionread

import "testing"

func TestPolicyRejectsWhitespaceSourceName(
	t *testing.T,
) {
	policy := DefaultPolicy()
	policy.DataSource.SourceName = "   "

	if err := policy.Validate(); err == nil {
		t.Fatal(
			"whitespace-only data source name passed validation",
		)
	}
}
