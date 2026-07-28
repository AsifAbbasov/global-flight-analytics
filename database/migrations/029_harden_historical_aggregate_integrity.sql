BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM historical_aggregate_results
        WHERE scope_type = 'region'
    ) THEN
        RAISE EXCEPTION
            'historical aggregate region rows must be rematerialized before migration 029';
    END IF;
END;
$$;

ALTER TABLE historical_aggregate_results
    DROP CONSTRAINT historical_aggregate_results_scope_check;

ALTER TABLE historical_aggregate_results
    ADD CONSTRAINT historical_aggregate_results_scope_check
    CHECK (
        (
            scope_type = 'global'
            AND scope_key = 'global'
            AND region_code = ''
            AND airport_icao_code = ''
            AND origin_icao_code = ''
            AND destination_icao_code = ''
        )
        OR
        (
            scope_type = 'region'
            AND scope_key = 'region:' || region_code
            AND region_code ~ '^[a-z0-9][a-z0-9_-]{1,31}$'
            AND airport_icao_code = ''
            AND origin_icao_code = ''
            AND destination_icao_code = ''
        )
        OR
        (
            scope_type = 'airport'
            AND scope_key = 'airport:' || airport_icao_code
            AND region_code = ''
            AND airport_icao_code ~ '^[A-Z0-9]{4}$'
            AND origin_icao_code = ''
            AND destination_icao_code = ''
        )
        OR
        (
            scope_type = 'route'
            AND scope_key =
                'route:' ||
                origin_icao_code ||
                ':' ||
                destination_icao_code
            AND region_code = ''
            AND airport_icao_code = ''
            AND origin_icao_code ~ '^[A-Z0-9]{4}$'
            AND destination_icao_code ~ '^[A-Z0-9]{4}$'
        )
    );

ALTER TABLE historical_aggregate_results
    ADD CONSTRAINT historical_aggregate_results_timestamp_mirror_check
    CHECK (
        abs(
            extract(epoch FROM window_start) * 1000000000 -
            window_start_unix_nano::numeric
        ) < 1000
        AND abs(
            extract(epoch FROM window_end) * 1000000000 -
            window_end_unix_nano::numeric
        ) < 1000
        AND abs(
            extract(epoch FROM as_of_time) * 1000000000 -
            as_of_time_unix_nano::numeric
        ) < 1000
        AND abs(
            extract(epoch FROM stored_at) * 1000000000 -
            stored_at_unix_nano::numeric
        ) < 1000
    );

ALTER TABLE historical_aggregate_results
    ADD CONSTRAINT historical_aggregate_results_json_metadata_check
    CHECK (
        (result_json ->> 'SchemaVersion')
            IS NOT DISTINCT FROM schema_version
        AND (result_json #>> '{Metric,Name}')
            IS NOT DISTINCT FROM metric_name
        AND (result_json #>> '{Scope,Type}')
            IS NOT DISTINCT FROM scope_type
        AND (result_json #>> '{Scope,RegionCode}')
            IS NOT DISTINCT FROM region_code
        AND (result_json #>> '{Scope,AirportICAOCode}')
            IS NOT DISTINCT FROM airport_icao_code
        AND (result_json #>> '{Scope,OriginICAOCode}')
            IS NOT DISTINCT FROM origin_icao_code
        AND (result_json #>> '{Scope,DestinationICAOCode}')
            IS NOT DISTINCT FROM destination_icao_code
        AND (
            result_json #>> '{Window,StartTime}'
        )::timestamptz IS NOT DISTINCT FROM window_start
        AND (
            result_json #>> '{Window,EndTime}'
        )::timestamptz IS NOT DISTINCT FROM window_end
        AND (
            result_json #>> '{Window,AsOfTime}'
        )::timestamptz IS NOT DISTINCT FROM as_of_time
        AND (result_json ->> 'Granularity')
            IS NOT DISTINCT FROM granularity
        AND (result_json ->> 'Status')
            IS NOT DISTINCT FROM series_status
        AND (result_json #>> '{Confidence,Level}')
            IS NOT DISTINCT FROM confidence_level
        AND (
            result_json #>> '{Provenance,InputFingerprint}'
        ) IS NOT DISTINCT FROM input_fingerprint
    );

ALTER TABLE historical_aggregate_results
    ADD CONSTRAINT historical_aggregate_results_stored_at_causality_check
    CHECK (
        stored_at >=
            (result_json ->> 'GeneratedAt')::timestamptz
    );

COMMIT;
