package featurestore

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPostgresStoreOwnsTimestampMirrorConsistency(
	t *testing.T,
) {
	_, fileName, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}

	contents, err := os.ReadFile(
		filepath.Join(filepath.Dir(fileName), "postgres.go"),
	)
	if err != nil {
		t.Fatalf("read postgres.go: %v", err)
	}

	source := string(contents)
	if strings.Count(
		source,
		"as_of_time,\n\t\t\tas_of_time_unix_nano",
	) != 6 {
		t.Fatalf(
			"as-of timestamp mirror pair count changed",
		)
	}
	if strings.Count(
		source,
		"stored_at,\n\t\t\tstored_at_unix_nano",
	) != 6 {
		t.Fatalf(
			"stored timestamp mirror pair count changed",
		)
	}
	if !strings.Contains(
		source,
		"validateTimestampMirror(\n\t\t\"as_of_time\"",
	) || !strings.Contains(
		source,
		"validateTimestampMirror(\n\t\t\"stored_at\"",
	) {
		t.Fatal("scan path no longer validates both timestamp mirrors")
	}
	if !strings.Contains(
		source,
		"as_of_time_unix_nano < $3",
	) || !strings.Contains(
		source,
		"as_of_time_unix_nano DESC",
	) {
		t.Fatal("exact Unix-nanosecond key ownership changed")
	}
}
