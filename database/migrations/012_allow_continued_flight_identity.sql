BEGIN;

ALTER TABLE flight_trajectories
    DROP CONSTRAINT flight_trajectories_split_reason_check;

ALTER TABLE flight_trajectories
    ADD CONSTRAINT flight_trajectories_split_reason_check
        CHECK (
            split_reason IN (
                '',
                'initial_observation',
                'source_flight_id_changed',
                'callsign_changed',
                'ground_cycle',
                'continued_from_previous_batch'
            )
        );

COMMIT;
