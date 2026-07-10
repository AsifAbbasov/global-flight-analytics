BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM data_quality_reports
        WHERE object_type <> 'flight_state'
    ) THEN
        RAISE EXCEPTION
            'unsupported data_quality_reports object_type values exist';
    END IF;
END;
$$;

ALTER TABLE data_quality_reports
    ADD COLUMN state_id uuid,
    ADD COLUMN flight_state_id uuid;

UPDATE data_quality_reports AS report
SET
    state_id = report.object_id,
    flight_state_id = (
        SELECT state.id
        FROM flight_states AS state
        WHERE state.id = report.object_id
    )
WHERE report.object_type = 'flight_state';

DROP INDEX IF EXISTS data_quality_reports_object_idx;

ALTER TABLE data_quality_reports
    DROP COLUMN object_type,
    DROP COLUMN object_id;

ALTER TABLE data_quality_reports
    ADD CONSTRAINT data_quality_reports_flight_state_fk
        FOREIGN KEY (flight_state_id)
        REFERENCES flight_states(id)
        ON DELETE SET NULL,
    ADD CONSTRAINT data_quality_reports_flight_state_identity_check
        CHECK (
            flight_state_id IS NULL
            OR (
                state_id IS NOT NULL
                AND state_id = flight_state_id
            )
        );

CREATE INDEX data_quality_reports_state_id_idx
    ON data_quality_reports (state_id);

CREATE INDEX data_quality_reports_flight_state_id_idx
    ON data_quality_reports (flight_state_id);

COMMIT;
