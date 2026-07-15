BEGIN;

CREATE TABLE historical_aggregate_results (
    id varchar(92) PRIMARY KEY,

    schema_version text NOT NULL,
    metric_name text NOT NULL,

    scope_type text NOT NULL,
    scope_key text NOT NULL,
    region_code text NOT NULL DEFAULT '',
    airport_icao_code text NOT NULL DEFAULT '',
    origin_icao_code text NOT NULL DEFAULT '',
    destination_icao_code text NOT NULL DEFAULT '',

    granularity text NOT NULL,

    window_start timestamptz NOT NULL,
    window_start_unix_nano bigint NOT NULL,
    window_end timestamptz NOT NULL,
    window_end_unix_nano bigint NOT NULL,
    as_of_time timestamptz NOT NULL,
    as_of_time_unix_nano bigint NOT NULL,

    input_fingerprint varchar(71) NOT NULL,
    series_status text NOT NULL,
    confidence_level text NOT NULL,

    result_json jsonb NOT NULL,

    stored_at timestamptz NOT NULL,
    stored_at_unix_nano bigint NOT NULL,

    CONSTRAINT historical_aggregate_results_id_check
        CHECK (
            id ~ '^historical-aggregate-record-[0-9a-f]{64}$'
        ),

    CONSTRAINT historical_aggregate_results_schema_version_check
        CHECK (
            schema_version = 'historical-intelligence-v1'
        ),

    CONSTRAINT historical_aggregate_results_metric_name_check
        CHECK (
            metric_name IN (
                'active_aircraft',
                'flight_count',
                'trajectory_count',
                'observation_count',
                'peak_activity',
                'average_activity',
                'traffic_density',
                'data_freshness',
                'coverage_score',
                'airport_departures',
                'airport_arrivals',
                'airport_operations',
                'unique_aircraft',
                'active_routes',
                'route_observations',
                'route_confidence',
                'complete_route_ratio',
                'partial_route_ratio',
                'unavailable_route_ratio',
                'great_circle_distance_km'
            )
        ),

    CONSTRAINT historical_aggregate_results_scope_check
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
                AND region_code ~ '^[A-Z0-9_-]{2,32}$'
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
        ),

    CONSTRAINT historical_aggregate_results_granularity_check
        CHECK (
            granularity IN (
                'hour',
                'day',
                'week',
                'custom'
            )
        ),

    CONSTRAINT historical_aggregate_results_window_check
        CHECK (
            window_start < window_end
            AND window_end <= as_of_time
            AND window_start_unix_nano < window_end_unix_nano
            AND window_end_unix_nano <= as_of_time_unix_nano
        ),

    CONSTRAINT historical_aggregate_results_input_fingerprint_check
        CHECK (
            input_fingerprint ~ '^sha256:[0-9a-f]{64}$'
        ),

    CONSTRAINT historical_aggregate_results_series_status_check
        CHECK (
            series_status IN (
                'unavailable',
                'partial',
                'complete'
            )
        ),

    CONSTRAINT historical_aggregate_results_confidence_level_check
        CHECK (
            confidence_level IN (
                'none',
                'low',
                'medium',
                'high'
            )
        ),

    CONSTRAINT historical_aggregate_results_json_check
        CHECK (
            jsonb_typeof(result_json) = 'object'
        ),

    CONSTRAINT historical_aggregate_results_key_unique
        UNIQUE (
            schema_version,
            metric_name,
            scope_key,
            granularity,
            window_start_unix_nano,
            window_end_unix_nano,
            as_of_time_unix_nano
        )
);

CREATE INDEX historical_aggregate_results_history_idx
    ON historical_aggregate_results (
        schema_version,
        metric_name,
        scope_key,
        granularity,
        window_end_unix_nano DESC,
        window_start_unix_nano DESC,
        as_of_time_unix_nano DESC,
        id ASC
    );

CREATE INDEX historical_aggregate_results_status_time_idx
    ON historical_aggregate_results (
        series_status,
        confidence_level,
        stored_at DESC
    );

COMMIT;
