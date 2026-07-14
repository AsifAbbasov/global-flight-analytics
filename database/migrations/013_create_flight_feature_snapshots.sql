BEGIN;

CREATE TABLE flight_feature_snapshots (
    id varchar(79) PRIMARY KEY,
    trajectory_id uuid NOT NULL
        REFERENCES flight_trajectories(id)
        ON DELETE CASCADE,
    schema_version text NOT NULL,
    as_of_time timestamptz NOT NULL,
    as_of_time_unix_nano bigint NOT NULL,
    input_fingerprint varchar(71) NOT NULL,
    validation_status text NOT NULL,
    features_json jsonb NOT NULL,
    stored_at timestamptz NOT NULL,
    stored_at_unix_nano bigint NOT NULL,

    CONSTRAINT flight_feature_snapshots_record_id_check
        CHECK (
            id ~ '^feature-record-[0-9a-f]{64}$'
        ),

    CONSTRAINT flight_feature_snapshots_schema_version_check
        CHECK (
            schema_version = 'flight-features-v1'
        ),

    CONSTRAINT flight_feature_snapshots_input_fingerprint_check
        CHECK (
            input_fingerprint ~ '^sha256:[0-9a-f]{64}$'
        ),

    CONSTRAINT flight_feature_snapshots_validation_status_check
        CHECK (
            validation_status IN ('valid', 'limited')
        ),

    CONSTRAINT flight_feature_snapshots_features_json_check
        CHECK (
            jsonb_typeof(features_json) = 'object'
        ),

    CONSTRAINT flight_feature_snapshots_snapshot_key_unique
        UNIQUE (
            trajectory_id,
            schema_version,
            as_of_time_unix_nano
        )
);

CREATE INDEX flight_feature_snapshots_history_idx
    ON flight_feature_snapshots (
        trajectory_id,
        schema_version,
        as_of_time_unix_nano DESC,
        id ASC
    );

CREATE INDEX flight_feature_snapshots_validation_time_idx
    ON flight_feature_snapshots (
        validation_status,
        stored_at DESC
    );

COMMIT;
