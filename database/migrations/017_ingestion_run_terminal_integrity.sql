BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM ingestion_runs
        WHERE
            (
                status = 'running'
                AND finished_at IS NOT NULL
            )
            OR
            (
                status IN ('success', 'failed', 'partial')
                AND finished_at IS NULL
            )
    ) THEN
        RAISE EXCEPTION
            'inconsistent ingestion run lifecycle rows exist';
    END IF;
END;
$$;

ALTER TABLE ingestion_runs
    ADD CONSTRAINT ingestion_runs_lifecycle_check
        CHECK (
            (
                status = 'running'
                AND finished_at IS NULL
            )
            OR
            (
                status IN ('success', 'failed', 'partial')
                AND finished_at IS NOT NULL
            )
        );

CREATE OR REPLACE FUNCTION enforce_ingestion_run_terminal_immutability()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF
        OLD.status IN ('success', 'failed', 'partial')
        AND NEW IS DISTINCT FROM OLD
    THEN
        RAISE EXCEPTION
            'completed ingestion run % is immutable',
            OLD.id
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS ingestion_runs_terminal_immutability
    ON ingestion_runs;

CREATE TRIGGER ingestion_runs_terminal_immutability
    BEFORE UPDATE ON ingestion_runs
    FOR EACH ROW
    EXECUTE FUNCTION enforce_ingestion_run_terminal_immutability();

COMMIT;
