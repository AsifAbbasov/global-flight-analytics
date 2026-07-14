BEGIN;

CREATE TABLE flight_route_results (
    id varchar(77) PRIMARY KEY,
    trajectory_id uuid NOT NULL
        REFERENCES flight_trajectories(id)
        ON DELETE CASCADE,
    schema_version text NOT NULL,
    as_of_time timestamptz NOT NULL,
    as_of_time_unix_nano bigint NOT NULL,
    input_fingerprint varchar(71) NOT NULL,
    route_status text NOT NULL,
    confidence_level text NOT NULL,
    validation_warning_count integer NOT NULL,
    route_json jsonb NOT NULL,
    stored_at timestamptz NOT NULL,
    stored_at_unix_nano bigint NOT NULL,

    CONSTRAINT flight_route_results_record_id_check
        CHECK (
            id ~ '^route-record-[0-9a-f]{64}$'
        ),

    CONSTRAINT flight_route_results_schema_version_check
        CHECK (
            schema_version = 'route-intelligence-v1'
        ),

    CONSTRAINT flight_route_results_input_fingerprint_check
        CHECK (
            input_fingerprint ~ '^sha256:[0-9a-f]{64}$'
        ),

    CONSTRAINT flight_route_results_route_status_check
        CHECK (
            route_status IN (
                'unavailable',
                'partial',
                'complete'
            )
        ),

    CONSTRAINT flight_route_results_confidence_level_check
        CHECK (
            confidence_level IN (
                'none',
                'low',
                'medium',
                'high'
            )
        ),

    CONSTRAINT flight_route_results_warning_count_check
        CHECK (
            validation_warning_count >= 0
        ),

    CONSTRAINT flight_route_results_route_json_check
        CHECK (
            jsonb_typeof(route_json) = 'object'
        ),

    CONSTRAINT flight_route_results_result_key_unique
        UNIQUE (
            trajectory_id,
            schema_version,
            as_of_time_unix_nano
        )
);

CREATE INDEX flight_route_results_history_idx
    ON flight_route_results (
        trajectory_id,
        schema_version,
        as_of_time_unix_nano DESC,
        id ASC
    );

CREATE INDEX flight_route_results_status_time_idx
    ON flight_route_results (
        route_status,
        confidence_level,
        stored_at DESC
    );

COMMIT;
