package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTypeScriptSetValues(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(
		directory,
		"parser.ts",
	)
	content := `const values = new Set<Example>([
  'beta',
  'alpha',
])
`
	if err := os.WriteFile(
		path,
		[]byte(content),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	values, err := parseTypeScriptSetValues(
		path,
		"values",
	)
	if err != nil {
		t.Fatalf(
			"parse set values: %v",
			err,
		)
	}

	if len(values) != 2 ||
		values[0] != "alpha" ||
		values[1] != "beta" {
		t.Fatalf(
			"values = %#v",
			values,
		)
	}
}

func TestParseTypeScriptSetValuesRejectsDuplicates(
	t *testing.T,
) {
	directory := t.TempDir()
	path := filepath.Join(
		directory,
		"parser.ts",
	)
	content := `const values = new Set<Example>([
  'alpha',
  'alpha',
])
`
	if err := os.WriteFile(
		path,
		[]byte(content),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := parseTypeScriptSetValues(
		path,
		"values",
	); err == nil {
		t.Fatal(
			"expected duplicate Set value error",
		)
	}
}

func TestAuditTrajectoryRuntimeParserAcceptsGroupedTypeImport(
	t *testing.T,
) {
	repositoryRoot := t.TempDir()
	path := filepath.Join(
		repositoryRoot,
		"apps/web/lib/api/trajectory.ts",
	)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	content := `import type {
  FlightIdentityBasis,
  FlightSplitReason,
} from '@/types/trajectory'

const flightIdentityBases = new Set<FlightIdentityBasis>([
  'source_flight_id',
  'callsign_and_start_time',
  'aircraft_and_start_time',
])

const flightSplitReasons = new Set<FlightSplitReason>([
  'initial_observation',
  'source_flight_id_changed',
  'callsign_changed',
  'ground_cycle',
  'continued_from_previous_batch',
])

function parseAircraftTrajectory(record: Record<string, unknown>) {
  return {
    identity_key: requireString(
      record.identity_key,
      'identity_key'
    ),
    identity_basis: identityBasis as FlightIdentityBasis,
    split_reason: splitReason as FlightSplitReason,
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	goEnums := map[string][]string{
		"FlightIdentityBasis": {
			"aircraft_and_start_time",
			"callsign_and_start_time",
			"source_flight_id",
		},
		"FlightSplitReason": {
			"callsign_changed",
			"continued_from_previous_batch",
			"ground_cycle",
			"initial_observation",
			"source_flight_id_changed",
		},
	}

	if err := auditTrajectoryRuntimeParser(repositoryRoot, goEnums); err != nil {
		t.Fatalf("audit grouped type import parser: %v", err)
	}
}

// STAGE-14-1-AUDIT-FALSE-POSITIVE-FIX
