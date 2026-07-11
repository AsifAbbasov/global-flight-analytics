BEGIN;

ALTER TABLE derived_reconciliation_tasks
    ADD COLUMN next_attempt_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN processing_started_at timestamptz,
    ADD COLUMN signal_version bigint NOT NULL DEFAULT 1,
    ADD COLUMN claimed_signal_version bigint;

UPDATE derived_reconciliation_tasks
SET
    processing_started_at = updated_at,
    claimed_signal_version = signal_version
WHERE status = 'processing';

ALTER TABLE derived_reconciliation_tasks
    ADD CONSTRAINT derived_reconciliation_tasks_signal_version_check
        CHECK (signal_version >= 1),
    ADD CONSTRAINT derived_reconciliation_tasks_processing_metadata_check
        CHECK (
            (
                status = 'processing'
                AND processing_started_at IS NOT NULL
                AND claimed_signal_version IS NOT NULL
            )
            OR
            (
                status <> 'processing'
                AND processing_started_at IS NULL
                AND claimed_signal_version IS NULL
            )
        ),
    ADD CONSTRAINT derived_reconciliation_tasks_claimed_signal_version_check
        CHECK (
            claimed_signal_version IS NULL
            OR claimed_signal_version <= signal_version
        );

DROP INDEX derived_reconciliation_tasks_status_idx;

CREATE INDEX derived_reconciliation_tasks_available_idx
    ON derived_reconciliation_tasks (
        next_attempt_at,
        created_at,
        id
    )
    WHERE status = 'pending';

CREATE INDEX derived_reconciliation_tasks_processing_started_idx
    ON derived_reconciliation_tasks (
        processing_started_at,
        id
    )
    WHERE status = 'processing';

COMMIT;
