package featurestore

import (
	"strings"
	"testing"
)

func TestFeatureTimestampSnapshotTableDDLOwnsProcessingIdentity(
	t *testing.T,
) {
	required := []string{
		"processing_version text NOT NULL",
		"trajectory_id,\n" +
			"\t\tschema_version,\n" +
			"\t\tprocessing_version,\n" +
			"\t\tas_of_time_unix_nano",
	}
	for _, fragment := range required {
		if !strings.Contains(
			featureTimestampSnapshotTableDDL,
			fragment,
		) {
			t.Fatalf(
				"feature timestamp fixture missing %q:\n%s",
				fragment,
				featureTimestampSnapshotTableDDL,
			)
		}
	}

	legacyIdentity := "trajectory_id,\n" +
		"\t\tschema_version,\n" +
		"\t\tas_of_time_unix_nano"
	if strings.Contains(
		featureTimestampSnapshotTableDDL,
		legacyIdentity,
	) {
		t.Fatalf(
			"feature timestamp fixture retains legacy uniqueness:\n%s",
			featureTimestampSnapshotTableDDL,
		)
	}
}
