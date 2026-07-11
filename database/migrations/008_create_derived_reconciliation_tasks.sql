BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE derived_reconciliation_tasks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    deduplication_key text NOT NULL UNIQUE,
    ingestion_run_id uuid,
    icao24 text NOT NULL,
    derivation_type text NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    observed_from timestamptz NOT NULL,
    observed_to timestamptz NOT NULL,
    attempt_count integer NOT NULL DEFAULT 0,
    last_error text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,

    CONSTRAINT derived_reconciliation_tasks_icao24_not_blank
        CHECK (btrim(icao24) <> ''),
    CONSTRAINT derived_reconciliation_tasks_derivation_type_check
        CHECK (derivation_type IN (
            'flight_state_quality',
            'trajectory'
        )),
    CONSTRAINT derived_reconciliation_tasks_status_check
        CHECK (status IN (
            'pending',
            'processing',
            'completed',
            'failed'
        )),
    CONSTRAINT derived_reconciliation_tasks_observed_range_check
        CHECK (observed_from <= observed_to),
    CONSTRAINT derived_reconciliation_tasks_attempt_count_check
        CHECK (attempt_count >= 0),
    CONSTRAINT derived_reconciliation_tasks_completed_at_check
        CHECK (
            (
                status = 'completed'
                AND completed_at IS NOT NULL
            )
            OR
            (
                status <> 'completed'
                AND completed_at IS NULL
            )
        )
);

CREATE INDEX derived_reconciliation_tasks_status_idx
    ON derived_reconciliation_tasks (status, updated_at);

CREATE INDEX derived_reconciliation_tasks_ingestion_run_idx
    ON derived_reconciliation_tasks (ingestion_run_id)
    WHERE ingestion_run_id IS NOT NULL;

CREATE INDEX derived_reconciliation_tasks_icao24_idx
    ON derived_reconciliation_tasks (icao24, derivation_type, status);

COMMIT;
