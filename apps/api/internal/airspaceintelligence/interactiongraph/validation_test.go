package interactiongraph

import (
	"testing"
)

func TestValidateRejectsTamperedDerivedFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Result)
		field  string
	}{
		{
			name: "metrics",
			mutate: func(result *Result) {
				result.Metrics.EdgeCount++
			},
			field: "metrics",
		},
		{
			name: "degree",
			mutate: func(result *Result) {
				result.Nodes[0].Degree++
			},
			field: "nodes[0].degree",
		},
		{
			name: "status",
			mutate: func(result *Result) {
				result.Status = ResultStatusLimited
			},
			field: "status",
		},
		{
			name: "confidence level",
			mutate: func(result *Result) {
				result.Confidence.Level = ConfidenceLevelLow
			},
			field: "confidence.level",
		},
		{
			name: "fingerprint",
			mutate: func(result *Result) {
				result.Provenance.InputFingerprint = "sha256:tampered"
			},
			field: "provenance.input_fingerprint",
		},
		{
			name: "scope guard",
			mutate: func(result *Result) {
				result.ScopeGuard = ""
			},
			field: "scope_guard",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result, err := Build(completeRequest())
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}
			test.mutate(&result)
			report := Validate(result)
			if report.Status != ValidationStatusInvalid {
				t.Fatalf("Validate() = %#v, want invalid", report)
			}
			if !hasIssue(report.Issues, test.field) {
				t.Fatalf("issues = %#v, want field %q", report.Issues, test.field)
			}
		})
	}
}

func TestPolicyBoundaries(t *testing.T) {
	t.Parallel()

	statusTests := []struct {
		nodes int
		edges int
		want  ResultStatus
	}{
		{nodes: 0, edges: 0, want: ResultStatusUnavailable},
		{nodes: 1, edges: 0, want: ResultStatusLimited},
		{nodes: MinimumCompleteNodeCount, edges: 0, want: ResultStatusLimited},
		{
			nodes: MinimumCompleteNodeCount,
			edges: MinimumCompleteEdgeCount,
			want:  ResultStatusComplete,
		},
	}
	for _, test := range statusTests {
		if got := statusForCounts(test.nodes, test.edges); got != test.want {
			t.Fatalf(
				"statusForCounts(%d, %d) = %q, want %q",
				test.nodes,
				test.edges,
				got,
				test.want,
			)
		}
	}

	confidenceTests := []struct {
		score float64
		want  ConfidenceLevel
	}{
		{score: 0, want: ConfidenceLevelNone},
		{score: 0.01, want: ConfidenceLevelLow},
		{
			score: MediumConfidenceMinimumScore,
			want:  ConfidenceLevelMedium,
		},
		{
			score: HighConfidenceMinimumScore,
			want:  ConfidenceLevelHigh,
		},
	}
	for _, test := range confidenceTests {
		if got := confidenceLevelForScore(test.score); got != test.want {
			t.Fatalf(
				"confidenceLevelForScore(%f) = %q, want %q",
				test.score,
				got,
				test.want,
			)
		}
	}
}

func TestFingerprintChangesWhenEvidenceChanges(t *testing.T) {
	t.Parallel()

	first, err := Build(completeRequest())
	if err != nil {
		t.Fatalf("Build(first) error = %v", err)
	}
	request := completeRequest()
	request.Nodes[0].Latitude += 0.01
	second, err := Build(request)
	if err != nil {
		t.Fatalf("Build(second) error = %v", err)
	}
	if first.Provenance.InputFingerprint ==
		second.Provenance.InputFingerprint {
		t.Fatal("fingerprint did not change after evidence changed")
	}
}

func hasIssue(issues []ValidationIssue, field string) bool {
	for _, issue := range issues {
		if issue.Field == field {
			return true
		}
	}
	return false
}
