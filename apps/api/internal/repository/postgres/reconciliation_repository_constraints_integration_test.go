package postgres

import (
	"context"
	"testing"
	"time"
)

func TestReconciliationSchemaRejectsPendingTaskWithCompletedTimestamp(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	observedAt := time.Date(
		2026,
		time.July,
		11,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	_, err := fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO derived_reconciliation_tasks (
				deduplication_key,
				icao24,
				derivation_type,
				status,
				observed_from,
				observed_to,
				completed_at
			)
			VALUES (
				'pending-with-completed-at',
				'abc123',
				'trajectory',
				'pending',
				$1,
				$1,
				now()
			);
		`,
		observedAt,
	)
	if err == nil {
		t.Fatal(
			"expected pending task with completed_at to violate schema constraint",
		)
	}
}

func TestReconciliationSchemaRejectsReversedObservedRange(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	observedAt := time.Date(
		2026,
		time.July,
		11,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	_, err := fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO derived_reconciliation_tasks (
				deduplication_key,
				icao24,
				derivation_type,
				status,
				observed_from,
				observed_to
			)
			VALUES (
				'reversed-observed-range',
				'abc123',
				'trajectory',
				'pending',
				$1,
				$2
			);
		`,
		observedAt,
		observedAt.Add(-time.Minute),
	)
	if err == nil {
		t.Fatal(
			"expected reversed observed range to violate schema constraint",
		)
	}
}

func TestReconciliationSchemaRejectsProcessingWithoutLifecycleMetadata(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	observedAt := time.Date(
		2026,
		time.July,
		11,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	_, err := fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO derived_reconciliation_tasks (
				deduplication_key,
				icao24,
				derivation_type,
				status,
				observed_from,
				observed_to
			)
			VALUES (
				'processing-without-lifecycle-metadata',
				'abc123',
				'trajectory',
				'processing',
				$1,
				$1
			);
		`,
		observedAt,
	)
	if err == nil {
		t.Fatal(
			"expected processing task without lifecycle metadata to violate schema constraint",
		)
	}
}

func TestReconciliationSchemaRejectsClaimVersionAboveSignalVersion(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	observedAt := time.Date(
		2026,
		time.July,
		11,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	_, err := fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO derived_reconciliation_tasks (
				deduplication_key,
				icao24,
				derivation_type,
				status,
				observed_from,
				observed_to,
				signal_version,
				claimed_signal_version,
				processing_started_at
			)
			VALUES (
				'claim-version-above-signal-version',
				'abc123',
				'trajectory',
				'processing',
				$1,
				$1,
				1,
				2,
				now()
			);
		`,
		observedAt,
	)
	if err == nil {
		t.Fatal(
			"expected claimed signal version above signal version to violate schema constraint",
		)
	}
}
