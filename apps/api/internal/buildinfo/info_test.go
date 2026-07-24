package buildinfo

import (
	"os"
	"testing"
)

func TestCurrentNormalizesBuildMetadata(
	t *testing.T,
) {
	originalVersion := version
	originalRevision := revision
	originalBuiltAt := builtAt
	t.Cleanup(
		func() {
			version = originalVersion
			revision = originalRevision
			builtAt = originalBuiltAt
		},
	)

	version = " 1.2.3 "
	revision = " abcdef123456 "
	builtAt = "2026-07-24T04:00:00+04:00"

	actual := Current()
	if actual.Version != "1.2.3" {
		t.Fatalf(
			"unexpected version: %q",
			actual.Version,
		)
	}
	if actual.Revision != "abcdef123456" {
		t.Fatalf(
			"unexpected revision: %q",
			actual.Revision,
		)
	}
	if actual.BuiltAt !=
		"2026-07-24T00:00:00Z" {
		t.Fatalf(
			"unexpected build timestamp: %q",
			actual.BuiltAt,
		)
	}
}

func TestCurrentFailsClosedForMissingOrInvalidMetadata(
	t *testing.T,
) {
	originalVersion := version
	originalRevision := revision
	originalBuiltAt := builtAt
	t.Cleanup(
		func() {
			version = originalVersion
			revision = originalRevision
			builtAt = originalBuiltAt
		},
	)

	version = " "
	revision = ""
	builtAt = "invalid"

	actual := Current()
	if actual.Version != DefaultVersion ||
		actual.Revision != UnknownRevision ||
		actual.BuiltAt != UnknownBuiltAt {
		t.Fatalf(
			"unexpected fail-closed metadata: %+v",
			actual,
		)
	}
}

func TestLinkerInjectedBuildMetadata(
	t *testing.T,
) {
	if os.Getenv(
		"GFA_EXPECT_LINKER_BUILD_INFO",
	) != "1" {
		t.Skip(
			"linker metadata verification is enabled by the installer",
		)
	}

	actual := Current()
	if actual.Version != "installer-test" ||
		actual.Revision != "0123456789abcdef" ||
		actual.BuiltAt != "2026-07-24T00:00:00Z" {
		t.Fatalf(
			"unexpected linker metadata: %+v",
			actual,
		)
	}
}
