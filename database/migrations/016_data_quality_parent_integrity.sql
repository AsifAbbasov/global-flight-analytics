BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM data_quality_reports
        WHERE state_id IS NULL
    ) THEN
        RAISE EXCEPTION
            'data_quality_reports rows without state_id must be repaired before migration 016';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM data_quality_reports
        WHERE flight_state_id IS NOT NULL
          AND state_id <> flight_state_id
    ) THEN
        RAISE EXCEPTION
            'data_quality_reports contains mismatched state identities';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM data_quality_reports
        WHERE flight_state_id IS NULL
          AND validation_status <> 'invalid'
    ) THEN
        RAISE EXCEPTION
            'non-invalid data_quality_reports rows without a Flight State require manual repair';
    END IF;
END;
$$;

CREATE TABLE rejected_flight_state_quality_reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    state_id uuid NOT NULL,
    icao24 varchar(10) NOT NULL DEFAULT '',
    callsign text NOT NULL DEFAULT '',
    observed_at timestamptz NOT NULL DEFAULT now(),
    source_name text NOT NULL DEFAULT '',
    ingestion_run_id uuid REFERENCES ingestion_runs(id) ON DELETE SET NULL,
    validation_status text NOT NULL,
    completeness text NOT NULL,
    confidence text NOT NULL,
    score numeric NOT NULL,
    missing_fields text[] NOT NULL DEFAULT '{}',
    warnings_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    calculated_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT rejected_flight_state_quality_validation_status_check
        CHECK (validation_status = 'invalid'),
    CONSTRAINT rejected_flight_state_quality_completeness_check
        CHECK (
            completeness IN (
                'complete',
                'partial',
                'position_only',
                'insufficient'
            )
        ),
    CONSTRAINT rejected_flight_state_quality_confidence_check
        CHECK (confidence IN ('none', 'low', 'medium', 'high')),
    CONSTRAINT rejected_flight_state_quality_score_check
        CHECK (score >= 0 AND score <= 1),
    CONSTRAINT rejected_flight_state_quality_warnings_check
        CHECK (jsonb_typeof(warnings_json) = 'array')
);

INSERT INTO rejected_flight_state_quality_reports (
    id,
    state_id,
    validation_status,
    completeness,
    confidence,
    score,
    missing_fields,
    warnings_json,
    calculated_at,
    created_at
)
SELECT
    id,
    state_id,
    validation_status,
    completeness,
    confidence,
    score,
    missing_fields,
    warnings_json,
    calculated_at,
    created_at
FROM data_quality_reports
WHERE flight_state_id IS NULL;

DELETE FROM data_quality_reports
WHERE flight_state_id IS NULL;

ALTER TABLE data_quality_reports
    DROP CONSTRAINT data_quality_reports_flight_state_fk,
    DROP CONSTRAINT data_quality_reports_flight_state_identity_check;

ALTER TABLE data_quality_reports
    ALTER COLUMN state_id SET NOT NULL,
    ALTER COLUMN flight_state_id SET NOT NULL;

ALTER TABLE data_quality_reports
    ADD CONSTRAINT data_quality_reports_flight_state_fk
        FOREIGN KEY (flight_state_id)
        REFERENCES flight_states(id)
        ON DELETE CASCADE,
    ADD CONSTRAINT data_quality_reports_flight_state_identity_check
        CHECK (state_id = flight_state_id);

CREATE INDEX rejected_flight_state_quality_reports_state_id_idx
    ON rejected_flight_state_quality_reports (state_id);

CREATE INDEX rejected_flight_state_quality_reports_icao24_observed_at_idx
    ON rejected_flight_state_quality_reports (icao24, observed_at DESC);

CREATE INDEX rejected_flight_state_quality_reports_ingestion_run_id_idx
    ON rejected_flight_state_quality_reports (ingestion_run_id);

COMMIT;
