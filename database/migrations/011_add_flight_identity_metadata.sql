BEGIN;

ALTER TABLE flight_trajectories
    ADD COLUMN identity_key text NOT NULL DEFAULT '',
    ADD COLUMN identity_basis text NOT NULL DEFAULT '',
    ADD COLUMN split_reason text NOT NULL DEFAULT '';

ALTER TABLE flight_trajectories
    ADD CONSTRAINT flight_trajectories_identity_completeness_check
        CHECK (
            (
                identity_key = ''
                AND identity_basis = ''
                AND split_reason = ''
            )
            OR
            (
                identity_key <> ''
                AND identity_basis <> ''
                AND split_reason <> ''
            )
        ),

    ADD CONSTRAINT flight_trajectories_identity_key_check
        CHECK (
            identity_key = ''
            OR identity_key ~ '^flight-identity-[0-9a-f]{64}$'
        ),

    ADD CONSTRAINT flight_trajectories_identity_basis_check
        CHECK (
            identity_basis IN (
                '',
                'source_flight_id',
                'callsign_and_start_time',
                'aircraft_and_start_time'
            )
        ),

    ADD CONSTRAINT flight_trajectories_split_reason_check
        CHECK (
            split_reason IN (
                '',
                'initial_observation',
                'source_flight_id_changed',
                'callsign_changed',
                'ground_cycle'
            )
        );

CREATE INDEX flight_trajectories_identity_key_time_idx
    ON flight_trajectories (
        identity_key,
        end_time DESC,
        start_time DESC
    )
    WHERE identity_key <> '';

COMMIT;
