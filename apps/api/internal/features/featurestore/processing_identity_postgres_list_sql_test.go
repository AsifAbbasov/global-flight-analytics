package featurestore

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestListSnapshotsSQLBindsProcessingVersionBeforeLimit(
	t *testing.T,
) {
	required := []string{
		"AND processing_version = $3",
		"LIMIT $4;",
	}
	for _, fragment := range required {
		if !strings.Contains(listSnapshotsSQL, fragment) {
			t.Fatalf(
				"listSnapshotsSQL missing %q:\n%s",
				fragment,
				listSnapshotsSQL,
			)
		}
	}

	if strings.Contains(listSnapshotsSQL, "LIMIT $3;") {
		t.Fatalf(
			"listSnapshotsSQL reuses processing-version placeholder for limit:\n%s",
			listSnapshotsSQL,
		)
	}
	if strings.Contains(
		listSnapshotsSQL,
		"as_of_time_unix_nano <",
	) {
		t.Fatalf(
			"non-cursor list query contains cursor boundary:\n%s",
			listSnapshotsSQL,
		)
	}
}

func TestPostgresStoreListWithoutCursorPassesProcessingVersion(
	t *testing.T,
) {
	client := &fakePostgresClient{
		query: func(
			_ context.Context,
			query string,
			args ...any,
		) (rowIterator, error) {
			if query != listSnapshotsSQL {
				t.Fatalf(
					"query = %q, want listSnapshotsSQL",
					query,
				)
			}
			if len(args) != 4 {
				t.Fatalf(
					"list args = %d, want 4: %#v",
					len(args),
					args,
				)
			}
			canonicalTrajectoryID := strings.ToLower(
				testTrajectoryID,
			)
			if args[0] != canonicalTrajectoryID {
				t.Fatalf(
					"trajectory argument = %#v, want canonical %q",
					args[0],
					canonicalTrajectoryID,
				)
			}
			if args[1] != string(
				flightfeatures.SchemaVersionV1,
			) {
				t.Fatalf(
					"schema argument = %#v",
					args[1],
				)
			}
			if args[2] != string(
				flightfeatures.CurrentProcessingVersion,
			) {
				t.Fatalf(
					"processing argument = %#v",
					args[2],
				)
			}
			if args[3] != 3 {
				t.Fatalf(
					"limit-with-sentinel argument = %#v",
					args[3],
				)
			}

			return rowsFromRecords(t, nil), nil
		},
	}
	store := newPostgresStore(client, time.Now)

	page, err := store.List(
		context.Background(),
		ListQuery{
			TrajectoryID:  testTrajectoryID,
			SchemaVersion: flightfeatures.SchemaVersionV1,
			Limit:         2,
		},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(page.Records) != 0 || page.HasMore {
		t.Fatalf("page = %#v, want empty page", page)
	}
}
