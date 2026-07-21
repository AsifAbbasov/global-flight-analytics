BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM ingestion_runs
        WHERE
            records_inserted > records_received
            OR records_updated > records_received - records_inserted
    ) THEN
        RAISE EXCEPTION
            'ingestion run processed counts exceed received records';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM ingestion_runs
        WHERE
            (
                status IN ('running', 'success')
                AND error_message IS NOT NULL
            )
            OR
            (
                status IN ('failed', 'partial')
                AND NULLIF(BTRIM(error_message), '') IS NULL
            )
    ) THEN
        RAISE EXCEPTION
            'ingestion run error-message semantics are inconsistent with status';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM flight_route_results
        WHERE
            ABS(
                EXTRACT(EPOCH FROM as_of_time) * 1000000000
                - as_of_time_unix_nano::numeric
            ) >= 1000
            OR ABS(
                EXTRACT(EPOCH FROM stored_at) * 1000000000
                - stored_at_unix_nano::numeric
            ) >= 1000
    ) THEN
        RAISE EXCEPTION
            'flight route result timestamp mirrors are inconsistent';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM historical_aggregate_results
        WHERE
            ABS(
                EXTRACT(EPOCH FROM window_start) * 1000000000
                - window_start_unix_nano::numeric
            ) >= 1000
            OR ABS(
                EXTRACT(EPOCH FROM window_end) * 1000000000
                - window_end_unix_nano::numeric
            ) >= 1000
            OR ABS(
                EXTRACT(EPOCH FROM as_of_time) * 1000000000
                - as_of_time_unix_nano::numeric
            ) >= 1000
            OR ABS(
                EXTRACT(EPOCH FROM stored_at) * 1000000000
                - stored_at_unix_nano::numeric
            ) >= 1000
    ) THEN
        RAISE EXCEPTION
            'historical aggregate timestamp mirrors are inconsistent';
    END IF;
END;
$$;

ALTER TABLE ingestion_runs
    ADD CONSTRAINT ingestion_runs_processed_counts_check
        CHECK (
            records_inserted <= records_received
            AND records_updated <= records_received - records_inserted
        ),
    ADD CONSTRAINT ingestion_runs_error_message_status_check
        CHECK (
            (
                status IN ('running', 'success')
                AND error_message IS NULL
            )
            OR
            (
                status IN ('failed', 'partial')
                AND NULLIF(BTRIM(error_message), '') IS NOT NULL
            )
        );

ALTER TABLE flight_route_results
    ADD CONSTRAINT flight_route_results_as_of_time_mirror_check
        CHECK (
            ABS(
                EXTRACT(EPOCH FROM as_of_time) * 1000000000
                - as_of_time_unix_nano::numeric
            ) < 1000
        ),
    ADD CONSTRAINT flight_route_results_stored_at_mirror_check
        CHECK (
            ABS(
                EXTRACT(EPOCH FROM stored_at) * 1000000000
                - stored_at_unix_nano::numeric
            ) < 1000
        );

ALTER TABLE historical_aggregate_results
    ADD CONSTRAINT historical_aggregate_results_window_start_mirror_check
        CHECK (
            ABS(
                EXTRACT(EPOCH FROM window_start) * 1000000000
                - window_start_unix_nano::numeric
            ) < 1000
        ),
    ADD CONSTRAINT historical_aggregate_results_window_end_mirror_check
        CHECK (
            ABS(
                EXTRACT(EPOCH FROM window_end) * 1000000000
                - window_end_unix_nano::numeric
            ) < 1000
        ),
    ADD CONSTRAINT historical_aggregate_results_as_of_time_mirror_check
        CHECK (
            ABS(
                EXTRACT(EPOCH FROM as_of_time) * 1000000000
                - as_of_time_unix_nano::numeric
            ) < 1000
        ),
    ADD CONSTRAINT historical_aggregate_results_stored_at_mirror_check
        CHECK (
            ABS(
                EXTRACT(EPOCH FROM stored_at) * 1000000000
                - stored_at_unix_nano::numeric
            ) < 1000
        );

COMMIT;
