package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditRulesAcceptsValidFile(
	t *testing.T,
) {
	root := t.TempDir()
	path := filepath.Join(
		root,
		"sample.go",
	)
	err := os.WriteFile(
		path,
		[]byte(
			"required\nrequired\nsafe\n",
		),
		0o600,
	)
	if err != nil {
		t.Fatalf(
			"write sample: %v",
			err,
		)
	}

	failures := auditRules(
		root,
		[]fileRule{
			{
				Name:     "sample",
				Path:     "sample.go",
				Required: []string{"safe"},
				Forbidden: []string{
					"forbidden",
				},
				Counts: []fragmentCount{
					{
						Fragment: "required",
						Minimum:  2,
						Maximum:  2,
					},
				},
			},
		},
	)
	if len(failures) != 0 {
		t.Fatalf(
			"unexpected failures: %#v",
			failures,
		)
	}
}

func TestAuditRulesReportsEveryInvariantFailure(
	t *testing.T,
) {
	root := t.TempDir()
	path := filepath.Join(
		root,
		"sample.go",
	)
	err := os.WriteFile(
		path,
		[]byte(
			"forbidden\ncount\ncount\ncount\n",
		),
		0o600,
	)
	if err != nil {
		t.Fatalf(
			"write sample: %v",
			err,
		)
	}

	failures := auditRules(
		root,
		[]fileRule{
			{
				Name:      "sample",
				Path:      "sample.go",
				Required:  []string{"missing"},
				Forbidden: []string{"forbidden"},
				Counts: []fragmentCount{
					{
						Fragment: "count",
						Minimum:  1,
						Maximum:  2,
					},
					{
						Fragment: "absent",
						Minimum:  1,
						Maximum:  -1,
					},
				},
			},
		},
	)
	if len(failures) != 4 {
		t.Fatalf(
			"failure count = %d, want 4: %#v",
			len(failures),
			failures,
		)
	}

	combined := ""
	for _, failure := range failures {
		combined += failure.Detail + "\n"
	}
	for _, expected := range []string{
		"missing required fragment",
		"contains forbidden fragment",
		"maximum is 2",
		"minimum is 1",
	} {
		if !strings.Contains(
			combined,
			expected,
		) {
			t.Fatalf(
				"failure output is missing %q: %s",
				expected,
				combined,
			)
		}
	}
}

func TestRunFailsForInvalidRepositoryRoot(
	t *testing.T,
) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{
			"-root",
			t.TempDir(),
			"-strict",
		},
		&stdout,
		&stderr,
	)
	if exitCode != 1 {
		t.Fatalf(
			"exit code = %d, want 1",
			exitCode,
		)
	}
	if !strings.Contains(
		stderr.String(),
		"required repository file",
	) {
		t.Fatalf(
			"unexpected stderr: %s",
			stderr.String(),
		)
	}
}

func TestRulesCoverAllFinalCorrectnessBoundaries(
	t *testing.T,
) {
	groups := map[string][]fileRule{
		"projection": projectionSnapshotRules(),
		"telemetry":  nullableTelemetryRules(),
		"pagination": historicalPaginationRules(),
		"weather":    weatherCompositionRules(),
		"evidence":   evidenceRules(),
	}

	for name, rules := range groups {
		if len(rules) == 0 {
			t.Fatalf(
				"%s rules are empty",
				name,
			)
		}
		for _, rule := range rules {
			if strings.TrimSpace(rule.Name) == "" ||
				strings.TrimSpace(rule.Path) == "" {
				t.Fatalf(
					"%s contains incomplete rule: %#v",
					name,
					rule,
				)
			}
		}
	}
}
